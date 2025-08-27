import { Button } from "@/components/ui/button";
import { ChatInput } from "@/components/ui/chat/chat-input";
import { CornerDownLeft, FileText, Eye, FileX, Copy, Check } from "lucide-react";
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
  $currentDocumentReferences,
  $ragEnabled,
  toggleRagEnabled,
  setRagEnabledForProject,
  $ragDocumentDetails,
  fetchRAGDocumentReference,
} from "@/store/chat";
import { EnhancedMarkdown } from "@/components/enhanced-markdown";
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
} from "@/components/ui/dialog";
import type { RAGDocumentReference, RAGDocumentReferenceChunk } from "proto/chatservice";

// Collapsible Chunks Display Component
function ChunksDisplay({ chunks }: { chunks: RAGDocumentReferenceChunk[] | undefined }) {
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
    <div className="max-h-[60vh] overflow-auto space-y-4 w-full">
      {chunks.map((chunk: RAGDocumentReferenceChunk, index: number) => {
        const isExpanded = expandedChunks.has(index);
        const chunkText = chunk.chunk_text || 'No content available';
        const words = chunkText.split(/\s+/);
        const shouldTruncate = words.length > 20;
        const displayText = shouldTruncate && !isExpanded 
          ? words.slice(0, 20).join(' ') 
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
                {isExpanded ? 'Show less' : 'Show more'}
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

  useEffect(() => {
    if (!projectId) {
      setRagEnabledForProject(false);
    }
  }, [projectId]);

  const { data, loading } = useStore($currentChatMessages);
  const streamingMessage = useStore($streamingMessage);
  const currentChatMessage = useStore($currentChatMessage);
  const availableModels = useStore($availableModels);
  const selectedModel = useStore($selectedModel);
  const listChatBranch = useStore($listChatBranch);
  const currentDocumentReferences = useStore($currentDocumentReferences);
  const ragEnabled = useStore($ragEnabled);
  const ragDocumentDetails = useStore($ragDocumentDetails);

  const [inputValue, setInputValue] = useState("");
  const [copiedMessageId, setCopiedMessageId] = useState<string | null>(null);
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

  const renderDocumentReferences = (references: RAGDocumentReference[], messageId?: string) => {
    if (!references || references.length === 0) return null;

    return (
      <div className="mt-3">
        <div className="text-xs text-gray-500 mb-2">Sources:</div>
        <div className="flex flex-wrap gap-2">
          {references.map((docRef, index) => (
            <Button
              key={`${docRef.doc_id}-${index}`}
              variant="outline"
              size="sm"
              className="text-xs h-6 px-2 bg-blue-50 border-blue-200 text-blue-700 hover:bg-blue-100"
              onClick={() => {
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


  const handleCopyMessage = async (content: string, messageId: string) => {
    try {
      await navigator.clipboard.writeText(content);
      setCopiedMessageId(messageId);
      setTimeout(() => setCopiedMessageId(null), 2000);
      console.log('Message copied to clipboard');
    } catch (error) {
      console.error('Failed to copy message:', error);
    }
  };

  return (
    <div className="flex flex-col h-full w-full px-6">
      <div className="flex-1 overflow-y-auto min-h-0">
        {loading ? (
          <div className="flex items-center justify-center h-full text-gray-500">
            Loading messages...
          </div>
        ) : data === undefined || data === null ? (
          <div className="flex items-center justify-center h-full text-gray-500">
            No messages yet
          </div>
        ) : (
          <div className="space-y-0">
            {data?.map((message, index) => (
              <div
                key={index}
                className={`w-full ${
                  message.role === "user" 
                    ? "bg-gray-50 border-b border-gray-200" 
                    : "bg-white border-b border-gray-200"
                } py-6 px-4`}
              >
                <div className="w-full max-w-none px-4">
                  <div className="flex items-start space-x-4">
                    {/* Avatar */}
                    <div className={`flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium ${
                      message.role === "user"
                        ? "bg-blue-600 text-white"
                        : "bg-green-600 text-white"
                    }`}>
                      {message.role === "user" ? "U" : "AI"}
                    </div>

                    {/* Message Content */}
                    <div className="flex-1 min-w-0">
                      <div className="prose prose-sm max-w-none">
                        <EnhancedMarkdown>
                          {message.content}
                        </EnhancedMarkdown>
                      </div>

                      {/* RAG Status Indicator */}
                      {projectId && message.role === "assistant" && !message.rag_enabled && (
                        <div className="mt-2 inline-flex items-center px-2 py-1 rounded-full text-xs bg-red-100 text-red-700">
                          <FileX className="h-3 w-3 mr-1" />
                          RAG not enabled
                        </div>
                      )}

                      {/* Document References */}
                      {message.role === "assistant" && message.references && (
                        renderDocumentReferences(message.references, message.message_id)
                      )}

                      {/* Action Buttons */}
                      {message.role === "assistant" && (
                        <div className="flex items-center space-x-2 mt-3">
                          <Button 
                            variant="ghost" 
                            size="sm"
                            onClick={() => handleCopyMessage(message.content, message.message_id)}
                            className="h-8 px-2 text-xs text-black-600 hover:text-gray-800"
                          >
                            {
                              copiedMessageId === message.message_id ?
                              <>
                                <Check className="h-4 w-4 text-green-400" /> 
                                <span className="text-xs text-green-400">Copied</span> 
                              </> : 
                              <>
                                <Copy className="h-3 w-3 text-gray-600" />
                                <span className="text-xs text-gray-600">Copy</span>
                              </>
                            }

                          </Button>
                          {message.message_id && (
                            <Button 
                              variant="ghost" 
                              size="sm" 
                              onClick={() => BranchChat(message.message_id)}
                              className="h-8 px-2 text-xs text-gray-600 hover:text-gray-800"
                            >
                              Branch Chat
                            </Button>
                          )}
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            ))}

            {/* Current user message */}
            {currentChatMessage && currentChatMessage.trim() && (
              <div className="w-full bg-gray-50 border-b border-gray-200 py-6 px-4">
                <div className="w-full max-w-none px-4">
                  <div className="flex items-start space-x-4">
                    <div className="flex-shrink-0 w-8 h-8 rounded-full bg-blue-600 text-white flex items-center justify-center text-sm font-medium">
                      U
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="prose prose-sm max-w-none">
                        <EnhancedMarkdown>
                          {currentChatMessage}
                        </EnhancedMarkdown>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* Streaming message */}
            {streamingMessage && streamingMessage.trim() && (
              <div className="w-full bg-white  border-gray-200 py-6 px-4">
                <div className="w-full max-w-none px-4">
                  <div className="flex items-start space-x-4">
                    <div className="flex-shrink-0 w-8 h-8 rounded-full bg-green-600 text-white flex items-center justify-center text-sm font-medium">
                      AI
                    </div>
                    <div className="flex-1 min-w-0">
                      <div className="prose prose-sm max-w-none">
                        <EnhancedMarkdown>
                          {streamingMessage}
                        </EnhancedMarkdown>
                      </div>

                      {/* RAG Status for streaming */}
                      {projectId && !ragEnabled && (
                        <div className="mt-2 inline-flex items-center px-2 py-1 rounded-full text-xs bg-red-100 text-red-700">
                          <FileX className="h-3 w-3 mr-1" />
                          RAG not enabled
                        </div>
                      )}

                      {/* Document references for streaming */}
                      {currentDocumentReferences.length > 0 && (
                        renderDocumentReferences(currentDocumentReferences)
                      )}
                    </div>
                  </div>
                </div>
              </div>
            )}
            <div ref={messagesEndRef} />
          </div>
        )}

        {/* Related Chats */}
        {listChatBranch.length > 0 && (
          <div className="bg-gray-50 border-t py-4 px-4">
            <div className="w-full max-w-none px-4">
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
          </div>
        )}
      </div>

      {/* Input Section */}
      <div className="flex-shrink-0 bg-white border-t border-gray-200 p-4">
        <div className="w-full max-w-none px-4">
          {/* RAG Toggle for Project Chats */}
          {projectId && (
            <div className="flex items-center mb-3">
              <label className="flex items-center space-x-2 text-sm text-gray-700 cursor-pointer">
                <input
                  type="checkbox"
                  checked={ragEnabled}
                  onChange={toggleRagEnabled}
                  className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                />
                <span>Enable RAG (Retrieval-Augmented Generation)</span>
              </label>
            </div>
          )}
          
          <div className="relative rounded-lg border border-gray-300 bg-white focus-within:ring-2 focus-within:ring-blue-500 focus-within:border-blue-500">
            <ChatInput
              placeholder="Ask anything"
              className="min-h-12 resize-none rounded-lg bg-transparent border-0 p-3 shadow-none focus-visible:ring-0"
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyDown={handleKeyDown}
            />
            <div className="flex items-center justify-between p-3 pt-0">
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="outline" size="sm" className="text-xs">
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
              <Button 
                size="sm" 
                className="bg-black hover:bg-gray-800 text-white px-4"
                onClick={handleSend}
                disabled={!inputValue.trim()}
              >
                <CornerDownLeft className="size-3.5" />
              </Button>
            </div>
          </div>
        </div>
      </div>

      {/* RAG Document Details Dialog */}
      <Dialog open={!!selectedDocumentForDetails} onOpenChange={() => setSelectedDocumentForDetails(null)}>
        <DialogContent className="max-w-[60vw] max-h-[80vh] overflow-hidden">
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