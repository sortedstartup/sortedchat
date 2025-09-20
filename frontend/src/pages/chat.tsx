import { Button } from "@/components/ui/button";
import { ChatInput } from "@/components/ui/chat/chat-input";
import {
  CornerDownLeft,
  FileText,
  Eye,
  FileX,
  Copy,
  Check,
  Maximize2,
  Minimize2,
  Info,
  ArrowUp,
  ArrowDown,
  DollarSign,
  ChevronRight,
  Loader2,
  Square,
} from "lucide-react";
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
  $ragEnabled,
  toggleRagEnabled,
  setRagEnabledForProject,
  $ragDocumentDetails,
  fetchRAGDocumentReference,
  $currentUserMessageId,
  $chatMetadata,
  $responseSummaries,
  $currentAssistantMessageId,
  $chatProgress,
  stream,
  $isStreaming,
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
import type {
  ChatMessage,
  RAGDocumentReference, 
  RAGDocumentReferenceChunk, 
  ResponseSummary,
  ChatProgress,
} from "proto/chatservice";

function getProgressText(state: number): string {
  switch (state) {
    case 0: // SENDING_REQUEST_TO_LLM
      return "Sending request...";
    case 1: // REQUEST_SENT_TO_LLM
      return "Request sent to LLM...";
    case 2: // FIRST_RESPONSE_RECEIVED
      return "Response received from LLM...";
    case 3: // FIRST_TOKEN_RECEIVED
      return "First token received from LLM...";
    case 4: // TOKENS_STREAMING
      return "Tokens streaming from LLM...";
    case 5: // TOKENS_STOPPED
      return "Finalizing...";
    default:
      return "Processing...";
  }
}

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

interface MessageProps {
  message: ChatMessage & { isProgress?: boolean };
  onCopyMessage: (content: string, messageId: string) => void;
  onViewRAGDetails: (messageId: string, docId: string, fileName: string) => void;
  onBranchChat: (messageId: string) => void;
  isCopied: boolean;
  projectId?: string;
  isExpanded: boolean;
  onToggleExpand: () => void;
  messageSummary?: ResponseSummary;
  chatProgress?: ChatProgress;
}

function formatCostAndTokens(
  cost: number | undefined,
  cachedTokens: number | undefined,
  showCachedTokens = true
): { costDisplay: string; cachedTokensDisplay: string } {
  let costDisplay = "";
  if (cost !== undefined) {
    costDisplay = cost < 1 ? `${(cost * 100).toFixed(3)} cents` : `${cost.toFixed(2)}`;
  }

  const cachedTokensDisplay = showCachedTokens && cachedTokens && cachedTokens > 0 ? cachedTokens.toString() : "";

  return { costDisplay, cachedTokensDisplay };
}

function Message({
  message,
  onCopyMessage,
  onViewRAGDetails,
  onBranchChat,
  isCopied,
  projectId,
  isExpanded,
  onToggleExpand,
  messageSummary,
  chatProgress,
}: MessageProps) {
  const [isHovered, setIsHovered] = useState(false);

  const isUser = message.role === "user";
  const isProgress = message.isProgress;

  const { costDisplay, cachedTokensDisplay } = formatCostAndTokens(
    messageSummary?.cost ?? message?.cost,
    messageSummary?.cached_tokens ?? message.cached_tokens,
    true
  );

  return (
    <div
      className={`w-full ${isUser
          ? "bg-gray-50 border-b border-gray-200"
          : "bg-white border-b border-gray-200"
        } py-6 px-4`}
    >
      <div
        className={`w-full max-w-none px-4 flex items-start space-x-4 justify-${isUser ? "end" : "start"
          }`}
      >
        {!isUser && (
          <div className="flex-shrink-0 w-8 h-8 rounded-full bg-green-600 text-white flex items-center justify-center text-sm font-medium">
            AI
          </div>
        )}

        <div className={`flex-1 min-w-0 text-${isUser ? "right" : "left"}`}>
          {isProgress ? (
            <div className="flex items-center space-x-2 text-sm text-gray-600">
              <Loader2 className="h-4 w-4 animate-spin" />
              <span>{chatProgress?.message || getProgressText(chatProgress?.state || 0)}</span>
            </div>
          ) : (
            <EnhancedMarkdown>{message.content}</EnhancedMarkdown>
          )}

          {!isUser && !isProgress && projectId && message.rag_enabled == false && (
            <div className="mt-2 inline-flex items-center px-2 py-1 rounded-full text-xs bg-red-100 text-red-700">
              <FileX className="h-3 w-3 mr-1" />
              RAG not enabled
            </div>
          )}

          {!isUser && !isProgress && message.references && message?.references.length > 0 && (
            <div className="mt-3">
              <div className="text-xs text-gray-500 mb-2">Sources:</div>
              <div className="flex flex-wrap gap-2">
                {message.references.map((docRef: RAGDocumentReference, idx: number) => (
                  <Button
                    key={`${docRef.doc_id}-${idx}`}
                    variant="outline"
                    size="sm"
                    className="text-xs h-6 px-2 bg-blue-50 border-blue-200 text-blue-700 hover:bg-blue-100"
                    onClick={() =>
                      onViewRAGDetails(message.message_id, docRef.doc_id, docRef.file_name)
                    }
                  >
                    <FileText className="h-3 w-3 mr-1" />
                    {docRef.file_name}
                    {docRef.Chunks?.length > 0 && (
                      <span className="ml-1 bg-blue-200 text-blue-800 px-1 rounded text-xs">
                        {docRef.Chunks.length}
                      </span>
                    )}
                    <Eye className="h-3 w-3 ml-1" />
                  </Button>
                ))}
              </div>
            </div>
          )}

          {!isUser && !isProgress && (
            <div className="flex justify-between mt-3">
              <div className="flex items-center space-x-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onCopyMessage(message.content, message.message_id)}
                  className="h-8 px-2 text-xs text-gray-600 hover:text-gray-800"
                >
                  {isCopied ? (
                    <Check className="h-4 w-4 text-green-400" />
                  ) : (
                    <Copy className="h-3 w-3 text-gray-600" />
                  )}
                </Button>

                <Button
                  variant="ghost"
                  size="sm"
                  onClick={onToggleExpand}
                  className="h-8 px-2 text-xs text-gray-600 hover:text-gray-800"
                >
                  {isExpanded ? (
                    <Minimize2 className="h-4 w-4" />
                  ) : (
                    <Maximize2 className="h-4 w-4" />
                  )}
                </Button>

                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onBranchChat(message.message_id)}
                  className="h-8 px-2 text-xs text-gray-600 hover:text-gray-800"
                >
                  Branch Chat
                </Button>
              </div>

              <div
                className="flex items-center space-x-2 text-xs text-gray-600"
                onMouseEnter={() => setIsHovered(true)}
                onMouseLeave={() => setIsHovered(false)}
              >
                {isHovered ? (
                  <>
                    <div className="flex items-center space-x-1">
                      <ArrowUp className="size-3" />
                      <span>
                        {messageSummary?.input_tokens || message.input_tokens}
                        {cachedTokensDisplay ? `/${cachedTokensDisplay}` : ""}
                      </span>
                    </div>
                    <div className="flex items-center space-x-1">
                      <ArrowDown className="size-3" />
                      <span>{messageSummary?.output_tokens || message.output_tokens}</span>
                    </div>
                    <div className="flex items-center space-x-1">
                      <DollarSign className="size-3" />
                      <span>{costDisplay}</span>
                    </div>
                  </>
                ) : (
                  <Info className="h-4 w-4 text-gray-500 hover:text-gray-800" />
                )}
              </div>
            </div>
          )}
        </div>

        {isUser && (
          <div className="flex-shrink-0 w-8 h-8 rounded-full bg-blue-600 text-white flex items-center justify-center text-sm font-medium">
            U
          </div>
        )}
      </div>
    </div>
  );
}

function ChatInputBox({
  projectId,
  onSendMessage,
}: {
  projectId?: string;
  onSendMessage: (message: string) => void;
}) {
  const [inputValue, setInputValue] = useState("");
  const [showDetailedTokens, setShowDetailedTokens] = useState(() => {
    const saved = localStorage.getItem('showDetailedTokens');
    return saved ? JSON.parse(saved) : false;
  });

  const toggleDetailedTokens = () => {
    const newValue = !showDetailedTokens;
    setShowDetailedTokens(newValue);
    localStorage.setItem('showDetailedTokens', JSON.stringify(newValue));
  };
  
  const availableModels = useStore($availableModels);
  const selectedModel = useStore($selectedModel);
  const ragEnabled = useStore($ragEnabled);
  const chatMetadata = useStore($chatMetadata);
  const isStreaming = useStore($isStreaming);

  const handleSend = () => {
    if (inputValue.trim() && !isStreaming) {
      onSendMessage(inputValue);
      setInputValue("");
    }
  };

  const handleStop = () => {
    $isStreaming.set(false);
    if (stream) {
      stream.cancel();
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    // Avoid sending while the user is composing text (IME) or while streaming
    if (
      e.key === "Enter" &&
      !e.shiftKey &&
      !isStreaming &&
      !(e.nativeEvent?.isComposing || (e as any).isComposing)
    ) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleModelSelect = (model: string) => {
    $selectedModel.set(model);
  };

  const { costDisplay, cachedTokensDisplay } = formatCostAndTokens(
    chatMetadata?.cost,
    chatMetadata?.cached_token_count
  );

  return (
    <>
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
                  disabled={isStreaming} // Disable during streaming
                />
                <span>Enable RAG (Retrieval-Augmented Generation)</span>
              </label>
            </div>
          )}

          <div className="relative rounded-lg border border-gray-300 bg-white focus-within:ring-2 focus-within:ring-blue-500 focus-within:border-blue-500">
            <ChatInput
              placeholder={isStreaming ? "Response is being generated..." : "Ask anything"}
              className="min-h-12 resize-none rounded-lg bg-transparent border-0 p-3 shadow-none focus-visible:ring-0"
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyDown={handleKeyDown}
              disabled={isStreaming} // Disable input during streaming
            />
            <div className="flex items-center justify-between p-3 pt-0">
              <div className="flex items-center space-x-2">
                <DropdownMenu>
                  <DropdownMenuTrigger asChild>
                    <Button 
                      variant="outline" 
                      size="sm" 
                      className="text-xs"
                    >
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
              </div>
              
              {isStreaming ? (
                <Button
                  size="sm"
                  variant="destructive"
                  className="px-4"
                  onClick={handleStop}
                  title="Stop Streaming"
                >
                  <Square className="size-3.5" />
                </Button>
              ) : (
                <Button
                  size="sm"
                  className="bg-black hover:bg-gray-800 text-white px-4"
                  onClick={handleSend}
                  disabled={!inputValue.trim()}
                >
                  <CornerDownLeft className="size-3.5" />
                </Button>
              )}
            </div>
          </div>
        </div>
        <div className="text-sm text-gray-500 mt-2 flex flex-row gap-2 px-6">
          <div className="flex items-center gap-1 px-2 py-1 rounded-full bg-gray-50 border border-gray-200">
            <ArrowDown className="size-3" />
            <span>{chatMetadata?.output_token_count}</span>
          </div>
          <button 
            onClick={toggleDetailedTokens}
            className="flex items-center gap-1 text-gray-400 hover:text-gray-600 transition-colors px-1"
            aria-label={showDetailedTokens ? "Hide detailed token usage" : "Show detailed token usage"}
          >
            <ChevronRight className={`size-3 transition-transform ${showDetailedTokens ? 'rotate-90' : ''}`} />
          </button>
          {showDetailedTokens && (
            <>
              
              {cachedTokensDisplay && <div className="flex items-center gap-1  px-2 py-1 rounded-full bg-gray-50 border border-gray-200">
                <span>{cachedTokensDisplay} cached tokens</span>
              </div>}
              <div className="flex items-center gap-1  px-2 py-1 rounded-full bg-gray-50 border border-gray-200">
                <DollarSign className="size-3"/>
                <span>{costDisplay}</span>
              </div>
            </>
          )}
        </div>
      </div>
    </>
  );
}

export function Chat() {
  const { projectId, chatId } = useParams();
  const navigate = useNavigate();

  const [copiedMessageId, setCopiedMessageId] = useState<string | null>(null);
  const [isExpanded, setIsExpanded] = useState(false);
  const [selectedDocumentForDetails, setSelectedDocumentForDetails] = useState<{
    messageId: string;
    docId: string;
    fileName: string;
  } | null>(null);

  const { data, loading } = useStore($currentChatMessages);
  const streamingMessage = useStore($streamingMessage);
  const currentChatMessage = useStore($currentChatMessage);
  const listChatBranch = useStore($listChatBranch);
  const ragDocumentDetails = useStore($ragDocumentDetails);
  const currentUserMessageId = useStore($currentUserMessageId);
  const responseSummaries = useStore($responseSummaries);
  const currentAssistantMessageId = useStore($currentAssistantMessageId);
  const chatProgress = useStore($chatProgress);

  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Show progress as assistant message when there's progress but no streaming content yet
  const showProgressAsMessage = chatProgress && !streamingMessage?.trim();

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

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [data, streamingMessage, currentChatMessage]);

  const handleSendMessage = (message: string) => {
    doChat(message, projectId);
  };

  const handleCopyMessage = async (content: string, messageId: string) => {
    await navigator.clipboard.writeText(content);
    setCopiedMessageId(messageId);
    setTimeout(() => setCopiedMessageId(null), 2000);
  };

  const handleViewRAGDetails = async (
    messageId: string,
    docId: string,
    fileName: string
  ) => {
    if (!projectId || !messageId) return;

    setSelectedDocumentForDetails({ messageId, docId, fileName });
    await fetchRAGDocumentReference(messageId, projectId, docId);
  };

  const goToChatBranch = (chatId: string) => {
    navigate(`/chat/${chatId}`);
  };

  const handleToggleExpand = () => {
    setIsExpanded((prev) => !prev);
  };

  const combinedMessages = [
    ...(data || []),
    ...(currentChatMessage?.trim() ? [{ message_id: currentUserMessageId, role: "user", content: currentChatMessage }] : []),
    ...(showProgressAsMessage ? [{ 
      message_id: Math.random().toString(36).substring(2, 15), //this should be unique
      role: "assistant", 
      content: "",
      isProgress: true 
    }] : []),
    ...(streamingMessage?.trim() ? [{ message_id: currentAssistantMessageId, role: "assistant", content: streamingMessage }] : []),
  ];

  return (
    <div
      className={`flex flex-col h-full mx-auto w-full transition-all ${isExpanded ? "max-w-7xl" : "max-w-4xl"
        }`}
    >
      <div className="flex-1 overflow-y-auto min-h-0">
        {loading ? (
          <div className="flex items-center justify-center h-full text-gray-500">
            Loading messages...
          </div>
        ) : combinedMessages.length === 0 ? (
          <div className="flex items-center justify-center h-full text-gray-500">
            No messages yet
          </div>
        ) : (
          combinedMessages.map((message,index) => {
            const summaryForThis = responseSummaries[message.message_id || ""];

            return (
              <Message
                key={index}
                message={message as ChatMessage & { isProgress?: boolean }}
                onCopyMessage={handleCopyMessage}
                onViewRAGDetails={handleViewRAGDetails}
                onBranchChat={BranchChat}
                isCopied={copiedMessageId === message?.message_id}
                projectId={projectId}
                isExpanded={isExpanded}
                onToggleExpand={handleToggleExpand}
                messageSummary={summaryForThis || undefined}
                chatProgress={chatProgress || undefined}
              />
            );
          })
        )}
        
        <div ref={messagesEndRef} />
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
      <ChatInputBox projectId={projectId} onSendMessage={handleSendMessage} />
      <Dialog
        open={!!selectedDocumentForDetails}
        onOpenChange={() => setSelectedDocumentForDetails(null)}
      >
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
                Showing {ragDocumentDetails.data.Chunks?.length || 0} chunk
                {ragDocumentDetails.data.Chunks?.length !== 1 ? "s" : ""} used to
                generate this response
              </div>
              <ChunksDisplay chunks={ragDocumentDetails.data.Chunks} />
            </div>
          )}
        </DialogContent>
      </Dialog>
    </div>
  );
}