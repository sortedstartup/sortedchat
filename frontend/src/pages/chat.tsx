import { Button } from "@/components/ui/button";
import {
  ChatBubble,
  ChatBubbleAvatar,
  ChatBubbleMessage,
} from "@/components/ui/chat/chat-bubble";
import { ChatInput } from "@/components/ui/chat/chat-input";
import { ChatMessageList } from "@/components/ui/chat/chat-message-list";
import { CornerDownLeft, FileText, Eye } from "lucide-react"; // Add FileText and Eye icons
import React, { useEffect, useRef, useState } from "react";
import { useStore } from "@nanostores/react";
import { useParams, useNavigate } from "react-router-dom";
import {
  $currentChatId,
  $selectedModel,
  doChat,
  $currentChatMessages,
  $streamingMessage,
  $currentChatMessage,
  $availableModels,
  BranchChat,
  $listChatBranch,
  $currentDocumentReferences, // Add this import
  $ragDocumentDetails,
  fetchRAGDocumentReference,
} from "@/store/chat";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from "@/components/ui/dropdown-menu";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"; // Add Dialog imports
import type { RAGDocumentReference } from "proto/chatservice";

// Collapsible Chunks Display Component
function ChunksDisplay({ chunks }: { chunks: any[] | undefined }) {
  const [expandedChunks, setExpandedChunks] = useState<Set<number>>(new Set());

  const toggleChunk = (index: number) => {
    const newExpanded = new Set(expandedChunks);
    if (newExpanded.has(index)) {
      newExpanded.delete(index);
    } else {
      newExpanded.add(index);
    }
    setExpandedChunks(newExpanded);
  };

  if (!chunks || chunks.length === 0) {
    return (
      <div className="text-gray-500 text-center p-4">
        No chunks available
      </div>
    );
  }

  return (
    <div className="max-h-[60vh] overflow-auto space-y-4">
      {chunks.map((chunk: any, index: number) => {
        const isExpanded = expandedChunks.has(index);
        const chunkText = chunk.chunk_text || 'No content available';
        const lines = chunkText.split('\n');
        const shouldTruncate = lines.length > 4;
        const displayText = shouldTruncate && !isExpanded 
          ? lines.slice(0, 4).join('\n') 
          : chunkText;

        return (
          <div key={index} className="bg-gray-50 rounded-lg p-4">
            <div className="flex justify-between items-center mb-2">
              <div className="text-xs text-gray-600 font-medium">
                Bytes {chunk.start_byte || 0} - {chunk.end_byte || 0}
              </div>
              <div className="text-xs bg-green-100 text-green-800 px-2 py-1 rounded">
                Similarity: {chunk.simillarity?.toFixed(3) || 'N/A'}
              </div>
            </div>
            <div className="text-sm leading-relaxed whitespace-pre-wrap">
              {displayText}
              {shouldTruncate && !isExpanded && (
                <span className="text-gray-400">...</span>
              )}
            </div>
            {shouldTruncate && (
              <button
                onClick={() => toggleChunk(index)}
                className="mt-2 text-xs text-blue-600 hover:text-blue-800 hover:underline focus:outline-none"
              >
                {isExpanded ? 'Show less' : `Show more (${lines.length - 4} more lines)`}
              </button>
            )}
          </div>
        );
      })}
    </div>
  );
}

export function Chat() {
  const { projectId, chatId } = useParams();
  const navigate = useNavigate();

  useEffect(() => {
    if (chatId) {
      $currentChatId.set(chatId);
    }
  }, [chatId]);

  useEffect(() => {
    const unsub = $currentChatId.listen((newId) => {
      if (newId && newId !== chatId) {
        navigate(`/chat/${newId}`, { replace: true });
      }
    });
    return () => unsub();
  }, [chatId, navigate]);

  const { data, loading } = useStore($currentChatMessages);
  const streamingMessage = useStore($streamingMessage);
  const currentChatMessage = useStore($currentChatMessage);
  const availableModels = useStore($availableModels);
  const selectedModel = useStore($selectedModel);
  const listChatBranch = useStore($listChatBranch);
  const currentDocumentReferences = useStore($currentDocumentReferences); // Add this
  const ragDocumentDetails = useStore($ragDocumentDetails);

  const [inputValue, setInputValue] = useState("");
  const [selectedDocumentForDetails, setSelectedDocumentForDetails] = useState<{
    messageId: string;
    docId: string;
    fileName: string;
  } | null>(null);

  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [data, streamingMessage, currentChatMessage]);

  const handleSend = () => {
    if (inputValue.trim()) {
      doChat(inputValue, projectId);
      setInputValue("");
      setTimeout(scrollToBottom, 100);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleModelSelect = (model: string) => {
    $selectedModel.set(model);
  };

  const goToChatBranch = async (chatId: string) => {
    try {
      navigate(`/chat/${chatId}`);
    } catch (error) {
      console.error('Failed to navigate to chat with project:', error);
    }
  };

  // Function to render document references for a message
  const renderDocumentReferences = (references: RAGDocumentReference[], messageId?: string) => {
    if (!references || references.length === 0) return null;

    return (
      <div className="mt-2 ml-2 sm:ml-4">
        <div className="text-xs text-gray-500 mb-1">Sources:</div>
        <div className="flex flex-wrap gap-1">
          {references.map((docRef, index) => (
            <Button
              key={`${docRef.doc_id}-${index}`}
              variant="outline"
              size="sm"
              className="text-xs h-6 px-2 bg-blue-50 border-blue-200 text-blue-700 hover:bg-blue-100"
              onClick={() => {
                console.log('Button clicked:', { 
                  messageId, 
                  docId: docRef.doc_id, 
                  fileName: docRef.file_name,
                  docRef 
                });
                messageId && handleViewRAGDetails(messageId, docRef.doc_id, docRef.file_name);
              }}
            >
              <FileText className="h-3 w-3 mr-1" />
              {docRef.file_name}
              {docRef.Chunks && docRef.Chunks.length > 0 && (
                <span className="ml-1 bg-blue-200 text-blue-800 px-1 rounded text-xs">
                  {docRef.Chunks.length}
                </span>
              )}
              <Eye className="h-3 w-3 ml-1" />
            </Button>
          ))}
        </div>
      </div>
    );
  };

  // Function to handle fetching detailed RAG document references
  const handleViewRAGDetails = async (messageId: string, docId: string, fileName: string) => {
    if (!projectId || !messageId) {
      console.error("Project ID and message ID are required");
      return;
    }

    try {
      setSelectedDocumentForDetails({ messageId, docId, fileName });
      await fetchRAGDocumentReference(messageId, projectId, docId);
    } catch (error) {
      console.error("Failed to fetch RAG details:", error);
    }
  };

  return (
    <div className="flex flex-col h-full mx-auto max-w-full w-full">
      <div className="flex-1 overflow-y-auto px-2 sm:px-4 min-h-0">
        <ChatMessageList className="flex flex-col gap-4 py-4">
          {loading ? (
            <div className="flex items-center justify-center h-full text-gray-500">
              Loading messages...
            </div>
          ) : data === undefined || data === null ? (
            <div className="flex items-center justify-center h-full text-gray-500">
              No messages yet
            </div>
          ) : (
            <>
              {data?.map((message, index) => (
                <div
                  key={index}
                  className={`flex flex-col ${
                    message.role === "user" ? "items-end" : "items-start"
                  }`}
                >
                  <div className={`flex ${
                    message.role === "user" ? "justify-end" : "justify-start"
                  }`}>
                    <ChatBubble
                      variant={message.role === "user" ? "sent" : "received"}
                      className="max-w-[95%] sm:max-w-[90%] lg:max-w-[85%] xl:max-w-[80%] mx-2 sm:mx-4"
                    >
                      <ChatBubbleAvatar
                        fallback={message.role === "user" ? "US" : "AI"}
                      />
                      <ChatBubbleMessage
                        variant={message.role === "user" ? "sent" : "received"}
                      >
                        <ReactMarkdown remarkPlugins={[remarkGfm]}>
                          {message.content}
                        </ReactMarkdown>
                      </ChatBubbleMessage>
                    </ChatBubble>
                  </div>
                  
                  {/* Show document references for assistant messages */}
                  {message.role === "assistant" && message.references && (
                    renderDocumentReferences(message.references, message.message_id)
                  )}
                  
                  {message.role === "assistant" && message.message_id && (
                    <div className="ml-2 sm:ml-4 mt-2">
                      <Button 
                        variant="outline" 
                        size="sm" 
                        onClick={() => BranchChat(message.message_id)}
                        className="text-xs"
                      >
                        Branch Chat
                      </Button>
                    </div>
                  )}
                </div>
              ))}

              {currentChatMessage && currentChatMessage.trim() && (
                <div className="flex justify-end">
                  <ChatBubble
                    variant="sent"
                    className="max-w-[95%] sm:max-w-[90%] lg:max-w-[85%] xl:max-w-[80%] mr-2 sm:mr-4"
                  >
                    <ChatBubbleAvatar fallback="US" />
                    <ChatBubbleMessage variant="sent">
                      <ReactMarkdown remarkPlugins={[remarkGfm]}>
                        {currentChatMessage}
                      </ReactMarkdown>
                    </ChatBubbleMessage>
                  </ChatBubble>
                </div>
              )}

              {streamingMessage && streamingMessage.trim() && (
                <div className="flex flex-col items-start">
                <div className="flex justify-start">
                  <ChatBubble
                    variant="received"
                    className="max-w-[95%] sm:max-w-[90%] lg:max-w-[85%] xl:max-w-[80%] ml-2 sm:ml-4"
                  >
                    <ChatBubbleAvatar fallback="AI" />
                    <ChatBubbleMessage variant="received">
                      <ReactMarkdown remarkPlugins={[remarkGfm]}>
                        {streamingMessage}
                      </ReactMarkdown>
                    </ChatBubbleMessage>
                  </ChatBubble>
                  </div>
                  
                  {/* Show document references for currently streaming message */}
                  {currentDocumentReferences.length > 0 && (
                    renderDocumentReferences(currentDocumentReferences)
                  )}
                </div>
              )}
              <div ref={messagesEndRef} />
            </>
          )}
        </ChatMessageList>
        
        {/* Inner Chat List */}
        {listChatBranch.length > 0 && (
          <div className="mt-4 px-2 sm:px-4">
            <h3 className="text-sm font-medium text-gray-700 mb-2">Related Chats:</h3>
            <div className="flex flex-wrap gap-2">
              {listChatBranch.map((chat) => (
                <Button
                  key={chat.chatId}
                  variant="outline"
                  size="sm"
                  onClick={() => goToChatBranch(chat.chatId)}
                  className="text-xs"
                >
                  {chat.name || "New Branch"}
                </Button>
              ))}
            </div>
          </div>
        )}
      </div>

      <div className="flex-shrink-0 bg-background p-2 sm:p-4 border-t">
        <div className="relative rounded-lg border bg-background focus-within:ring-1 focus-within:ring-ring p-1">
          <ChatInput
            placeholder="Type your message here..."
            className="min-h-12 resize-none rounded-lg bg-background border-0 p-3 shadow-none focus-visible:ring-0"
            value={inputValue}
            onChange={(e) => setInputValue(e.target.value)}
            onKeyDown={handleKeyDown}
          />
          <div className="flex items-center p-3 pt-0">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="outline" size="sm" className="mr-2">
                  {selectedModel || "Select Model"}
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent>
                {availableModels.map((model) => (
                  <DropdownMenuItem
                    key={model.id || model.label}
                    onClick={() => handleModelSelect(model.id)}
                  >
                    {model.label}
                  </DropdownMenuItem>
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
            <Button size="sm" className="ml-auto gap-1.5" onClick={handleSend}>
              Send Message
              <CornerDownLeft className="size-3.5" />
            </Button>
          </div>
        </div>
      </div>

      {/* RAG Document Details Dialog */}
      <Dialog open={!!selectedDocumentForDetails} onOpenChange={() => setSelectedDocumentForDetails(null)}>
        <DialogContent className="max-w-4xl max-h-[80vh] overflow-hidden">
          <DialogHeader>
            <DialogTitle className="text-lg">
              <FileText className="inline h-5 w-5 mr-2" />
              Document Chunks: {selectedDocumentForDetails?.fileName}
            </DialogTitle>
          </DialogHeader>
          
          {ragDocumentDetails.loading && (
            <div className="flex items-center justify-center p-8">
              <div className="text-sm text-gray-500">Loading document details...</div>
            </div>
          )}
          
          {ragDocumentDetails.error && (
            <div className="text-red-600 text-sm p-4 bg-red-50 rounded">
              Error: {ragDocumentDetails.error}
            </div>
          )}
          
          {ragDocumentDetails.data && (
            <div>
              <div className="text-sm text-gray-500 mb-4">
                {/* Requested: {selectedDocumentForDetails?.docId} | Received: {ragDocumentDetails.data.doc_id} */}
                <br />
                Showing {ragDocumentDetails.data.Chunks?.length || 0} chunk{(ragDocumentDetails.data.Chunks?.length || 0) > 1 ? 's' : ''} used to generate this response
              </div>
              <ChunksDisplay chunks={ragDocumentDetails.data.Chunks} />
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}