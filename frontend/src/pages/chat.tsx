import { Button } from "@/components/ui/button";
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
  ImageIcon,
  X,
} from "lucide-react";
import React, { useEffect, useRef, useState } from "react";
import { useStore } from "@nanostores/react";
import { useParams, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import {
  $currentChatId,
  doChat,
  $selectedModel,
  $availableModels,
  $currentChatMessages,
  $streamingMessage,
  $currentChatMessage,
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
  $currentProject,
  $currentProjectId,
  $currentChatError,
} from "@/store/chat";
import { EnhancedMarkdown } from "@/components/enhanced-markdown";
import { ModelSelector } from "@/components/ModelSelector";

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
          <div key={index} className="bg-muted rounded-lg p-4">
            <div className="flex justify-between items-center mb-2">
              <div className="text-xs text-muted-foreground font-medium">
                Bytes {chunk.start_byte || 0} - {chunk.end_byte || 0}
              </div>
              <div className="text-xs bg-green-500/10 text-green-600 dark:text-green-400 px-2 py-1 rounded">
                Similarity: {chunk.simillarity?.toFixed(3) || 'N/A'}
              </div>
            </div>
            <div className="text-sm text-foreground leading-relaxed whitespace-pre-wrap">
              {displayText}
              {shouldTruncate && !isExpanded && (
                <span className="text-muted-foreground/50">...</span>
              )}
            </div>
            {shouldTruncate && (
              <button
                onClick={() => toggleChunk(index)}
                className="mt-2 text-xs text-primary hover:text-primary/80 hover:underline focus:outline-none"
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
  // onBranchChat, //temporarily hiding it for this release only
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
        ? "bg-muted border-b border-border"
        : "bg-card border-b border-border"
        } py-6 px-4`}
    >
      <div
        className={`w-full max-w-none px-4 flex items-start space-x-4 ${isUser ? "justify-end" : "justify-start"}`}
      >
        {!isUser && (
          <div className="flex-shrink-0 w-8 h-8 rounded-full bg-green-600 text-white flex items-center justify-center text-sm font-medium">
            AI
          </div>
        )}

        <div className={`flex-1 min-w-0 ${isUser ? "text-right" : "text-left"}`}>
          {isProgress ? (
            <div className="flex items-center space-x-2 text-sm text-muted-foreground">
              <Loader2 className="h-4 w-4 animate-spin" />
              <span>{chatProgress?.message || getProgressText(chatProgress?.state || 0)}</span>
            </div>
          ) : (
            <div className="space-y-2">
              {/* Render multi-modal content if available */}
              {message.contents && message.contents.length > 0 ? (
                message.contents.map((content, idx) => {
                  if (content.type === "text" && content.text) {
                    return <EnhancedMarkdown key={idx}>{content.text}</EnhancedMarkdown>;
                  } else if (content.type === "image_url" && content.image_url) {
                    return (
                      <div key={idx} className="my-2">
                        <img
                          src={content.image_url.url}
                          alt="Message image"
                          className="max-w-md rounded-lg shadow-md border border-border"
                          loading="lazy"
                          style={{ maxHeight: '400px', objectFit: 'contain' }}
                        />
                      </div>
                    );
                  }
                  return null;
                })
              ) : (
                /* Fallback to old text-only format */
                <>
                  <EnhancedMarkdown>{message.content}</EnhancedMarkdown>
                </>
              )}
            </div>
          )}

          {!isUser && !isProgress && projectId && message.rag_enabled == false && (
            <div className="mt-2 inline-flex items-center px-2 py-1 rounded-full text-xs bg-destructive/10 text-destructive">
              <FileX className="h-3 w-3 mr-1" />
              RAG not enabled
            </div>
          )}

          {!isUser && !isProgress && message.references && message?.references.length > 0 && (
            <div className="mt-3">
              <div className="text-xs text-muted-foreground mb-2">Sources:</div>
              <div className="flex flex-wrap gap-2">
                {message.references.map((docRef: RAGDocumentReference, idx: number) => (
                  <Button
                    key={`${docRef.doc_id}-${idx}`}
                    variant="outline"
                    size="sm"
                    className="text-xs h-6 px-2 bg-primary/10 border-primary/20 text-primary hover:bg-primary/20"
                    onClick={() =>
                      onViewRAGDetails(message.message_id, docRef.doc_id, docRef.file_name)
                    }
                  >
                    <FileText className="h-3 w-3 mr-1" />
                    {docRef.file_name}
                    {docRef.Chunks?.length > 0 && (
                      <span className="ml-1 bg-primary/20 text-primary px-1 rounded text-xs">
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
                  className="h-8 px-2 text-xs text-muted-foreground hover:text-foreground"
                >
                  {isCopied ? (
                    <Check className="h-4 w-4 text-green-500" />
                  ) : (
                    <Copy className="h-3 w-3" />
                  )}
                </Button>

                <Button
                  variant="ghost"
                  size="sm"
                  onClick={onToggleExpand}
                  className="h-8 px-2 text-xs text-muted-foreground hover:text-foreground"
                >
                  {isExpanded ? (
                    <Minimize2 className="h-4 w-4" />
                  ) : (
                    <Maximize2 className="h-4 w-4" />
                  )}
                </Button>

                {/* Temporarily hiding it for this release only */}
                {/* <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => onBranchChat(message.message_id)}
                  className="h-8 px-2 text-xs text-muted-foreground hover:text-foreground"
                >
                  Branch Chat
                </Button> */}
              </div>

              <div
                className="flex items-center space-x-2 text-xs text-muted-foreground"
                onMouseEnter={() => setIsHovered(true)}
                onMouseLeave={() => setIsHovered(false)}
              >
                {isHovered ? (
                  <>
                    <div className="flex items-center space-x-1">
                      <ArrowUp className="size-3" />
                      <span>
                        {messageSummary?.input_tokens ?? message.input_tokens}
                        {cachedTokensDisplay ? `/${cachedTokensDisplay}` : ""}
                      </span>
                    </div>
                    <div className="flex items-center space-x-1">
                      <ArrowDown className="size-3" />
                      <span>{messageSummary?.output_tokens ?? message.output_tokens}</span>
                    </div>
                    <div className="flex items-center space-x-1">
                      <DollarSign className="size-3" />
                      <span>{costDisplay}</span>
                    </div>
                    <div className="flex items-center space-x-1">
                      <span>{messageSummary?.model ?? message.model}</span>
                    </div>
                  </>
                ) : (
                  <Info className="h-4 w-4 hover:text-foreground" />
                )}
              </div>
            </div>
          )}
        </div>

        {isUser && (
          <div className="flex-shrink-0 w-8 h-8 rounded-full bg-primary text-primary-foreground flex items-center justify-center text-sm font-medium">
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
  onSendMessage: (message: string, images?: File[], imageDetail?: string) => void;
}) {
  const MIN_TEXTAREA_HEIGHT = 48;
  const MAX_TEXTAREA_HEIGHT = 200;
  const [inputValue, setInputValue] = useState("");
  const [selectedImages, setSelectedImages] = useState<File[]>([]);
  const [imageDetail, setImageDetail] = useState<string>("auto");
  const [showDetailedTokens, setShowDetailedTokens] = useState(() => {
    const saved = localStorage.getItem('showDetailedTokens');
    return saved ? JSON.parse(saved) : false;
  });
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

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

  // Get current model capabilities
  const modelInfo = availableModels.find(m => m.id === selectedModel.model_name);
  const supportsImageInput = modelInfo?.capabilities?.image?.input ?? false;

  // Auto-resize textarea based on content
  useEffect(() => {
    const textarea = textareaRef.current;
    if (textarea) {
      // Reset height to auto to get the correct scrollHeight
      textarea.style.height = 'auto';

      // Calculate new height (min 48px, max 200px)
      const newHeight = Math.min(Math.max(textarea.scrollHeight, MIN_TEXTAREA_HEIGHT), MAX_TEXTAREA_HEIGHT);
      textarea.style.height = `${newHeight}px`;
    }
  }, [inputValue]);

  const handleImageSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []);

    // Validate file types
    const validFiles = files.filter(f => f.type.startsWith('image/'));

    // Validate file sizes (20MB limit)
    const validSizes = validFiles.filter(f => f.size <= 20 * 1024 * 1024);

    // Track validation issues
    const invalidTypeCount = files.length - validFiles.length;
    const invalidSizeCount = validFiles.length - validSizes.length;

    // Check total image count limit (current + new)
    const currentCount = selectedImages.length;
    const newValidCount = validSizes.length;
    const totalCount = currentCount + newValidCount;
    const maxImages = 10;

    let finalImages = validSizes;
    let wasLimited = false;

    if (totalCount > maxImages) {
      const availableSlots = maxImages - currentCount;
      if (availableSlots <= 0) {
        toast.error(`Maximum ${maxImages} images allowed. Remove some images first.`);
        return;
      }
      finalImages = validSizes.slice(0, availableSlots);
      wasLimited = true;
    }

    // Show appropriate error messages
    const errors = [];
    if (invalidTypeCount > 0) {
      errors.push(`${invalidTypeCount} file(s) skipped (invalid type)`);
    }
    if (invalidSizeCount > 0) {
      errors.push(`${invalidSizeCount} file(s) skipped (too large, max 20MB)`);
    }
    if (wasLimited) {
      const skippedCount = newValidCount - finalImages.length;
      errors.push(`${skippedCount} image(s) skipped (max ${maxImages} images allowed)`);
    }

    if (errors.length > 0) {
      toast.error(errors.join(", "));
    }

    // Only add images if we have valid ones to add
    if (finalImages.length > 0) {
      setSelectedImages(prev => [...prev, ...finalImages]);
      toast.success(`${finalImages.length} image(s) added`);
    }
  };

  const removeImage = (index: number) => {
    setSelectedImages(prev => prev.filter((_, i) => i !== index));
  };

  const handleSend = () => {
    if ((inputValue.trim() || selectedImages.length > 0) && !isStreaming) {
      onSendMessage(inputValue, selectedImages, imageDetail);
      setInputValue("");
      setSelectedImages([]);
      setImageDetail("auto"); // Reset to default
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

  const handleModelSelect = (model: string, provider: string) => {
    $selectedModel.set({ model_name: model, provider });
  };

  const { costDisplay, cachedTokensDisplay } = formatCostAndTokens(
    chatMetadata?.cost,
    chatMetadata?.cached_token_count
  );

  return (
    <>
      <div className="flex-shrink-0 bg-card border-t border-border p-4">
        <div className="w-full max-w-none px-4">
          {/* Image Preview Area */}
          {selectedImages.length > 0 && (
            <div className="mb-3 flex flex-wrap gap-2">
              {selectedImages.map((img, idx) => (
                <div key={idx} className="relative group">
                  <img
                    src={URL.createObjectURL(img)}
                    alt={`Preview ${idx}`}
                    className="h-20 w-20 object-cover rounded border border-border"
                  />
                  <button
                    onClick={() => removeImage(idx)}
                    className="absolute -top-2 -right-2 bg-destructive text-destructive-foreground rounded-full w-5 h-5 flex items-center justify-center opacity-0 group-hover:opacity-100 transition-opacity"
                  >
                    <X className="h-3 w-3" />
                  </button>
                </div>
              ))}
            </div>
          )}
          {/* RAG Toggle for Project Chats */}
          {projectId && (
            <div className="flex items-center mb-3">
              <label className="flex items-center space-x-2 text-sm text-foreground cursor-pointer">
                <input
                  type="checkbox"
                  checked={ragEnabled}
                  onChange={toggleRagEnabled}
                  className="rounded border-input text-primary focus:ring-ring"
                  disabled={isStreaming}
                />
                <span>Enable RAG (Retrieval-Augmented Generation)</span>
              </label>
            </div>
          )}

          <div className="relative rounded-lg border border-border bg-card focus-within:ring-2 focus-within:ring-ring focus-within:border-ring">
            <textarea
              ref={textareaRef}
              placeholder={isStreaming ? "Response is being generated..." : "Ask anything"}
              className="w-full min-h-[48px] max-h-[200px] resize-none rounded-lg bg-transparent border-0 p-3 shadow-none focus-visible:ring-0 focus:outline-none overflow-y-auto text-foreground placeholder:text-muted-foreground"
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyDown={handleKeyDown}
              disabled={isStreaming}
              rows={1}
              aria-label="Chat message input"
            />
            <div className="flex items-center justify-between p-3 pt-0">
              <div className="flex items-center space-x-2">
                {/* Image Upload Button - Only show if model supports it */}
                {supportsImageInput && (
                  <>
                    <input
                      ref={fileInputRef}
                      type="file"
                      accept="image/*"
                      multiple
                      className="hidden"
                      onChange={handleImageSelect}
                    />
                    <Button
                      variant="ghost"
                      size="sm"
                      onClick={() => fileInputRef.current?.click()}
                      disabled={isStreaming}
                      title="Add images"
                    >
                      <ImageIcon className="h-4 w-4" />
                    </Button>
                  </>
                )}

                <ModelSelector
                  selectedModelId={selectedModel.model_name}
                  onSelectModel={handleModelSelect}
                />
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
                  className="bg-primary hover:bg-primary/90 text-primary-foreground px-4"
                  onClick={handleSend}
                  disabled={!inputValue.trim() && selectedImages.length === 0}
                >
                  <CornerDownLeft className="size-3.5" />
                </Button>
              )}
            </div>
          </div>

          {/* Show warning if images selected but model doesn't support */}
          {!supportsImageInput && selectedImages.length > 0 && (
            <div className="mt-2 text-xs text-destructive">
              Selected model doesn't support images. Choose a vision-capable model.
            </div>
          )}

          {/* Optional: Detail level selector for advanced users */}
          {selectedImages.length > 0 && supportsImageInput && (
            <div className="mt-2 flex items-center space-x-2 text-xs text-muted-foreground">
              <span>Image detail:</span>
              <select
                className="text-xs border border-border rounded px-1 bg-card text-foreground"
                value={imageDetail}
                onChange={(e) => setImageDetail(e.target.value)}
              >
                <option value="auto">Auto (recommended)</option>
                <option value="low">Low (faster, cheaper)</option>
                <option value="high">High (slower, more detailed)</option>
              </select>
            </div>
          )}
        </div>
        <div className="text-sm text-muted-foreground mt-2 flex flex-row gap-2 px-6">
          <div className="flex items-center gap-1 px-2 py-1 rounded-full bg-muted border border-border">
            <ArrowUp className="size-3" />
            <span>{chatMetadata?.input_token_count}</span>
          </div>
          <div className="flex items-center gap-1 px-2 py-1 rounded-full bg-muted border border-border">
            <ArrowDown className="size-3" />
            <span>{chatMetadata?.output_token_count}</span>
          </div>
          <button
            onClick={toggleDetailedTokens}
            className="flex items-center gap-1 text-muted-foreground hover:text-foreground transition-colors px-1"
            aria-label={showDetailedTokens ? "Hide detailed token usage" : "Show detailed token usage"}
          >
            <ChevronRight className={`size-3 transition-transform ${showDetailedTokens ? 'rotate-90' : ''}`} />
          </button>
          {showDetailedTokens && (
            <>
              {cachedTokensDisplay && <div className="flex items-center gap-1 px-2 py-1 rounded-full bg-muted border border-border">
                <span>{cachedTokensDisplay} cached tokens</span>
              </div>}
              <div className="flex items-center gap-1 px-2 py-1 rounded-full bg-muted border border-border">
                <DollarSign className="size-3" />
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
  const currentProject = useStore($currentProject);
  const currentChatError = useStore($currentChatError);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const messagesContainerRef = useRef<HTMLDivElement>(null);
  const [isUserScrolledUp, setIsUserScrolledUp] = useState(false);
  const shouldAutoScroll = !isUserScrolledUp;

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
    if (projectId) {
      $currentProjectId.set(projectId);
    } else {
      setRagEnabledForProject(false);
    }
  }, [projectId]);

  // Scroll handler to detect when user scrolls up
  useEffect(() => {
    const container = messagesContainerRef.current;
    if (!container) return;

    const handleScroll = () => {
      const {
        scrollTop: scrollFromTop, //how far you’ve scrolled from the very top.
        scrollHeight: totalContentHeight, //total content height (like the full length of a long chat).
        clientHeight: viewportHeight, //visible window height (the viewport).
      } = container;
      const isAtBottom = totalContentHeight - scrollFromTop - viewportHeight < 100;
      //if user is at bottom of the chat, it will autoscroll, else it will not.

      setIsUserScrolledUp(!isAtBottom);
    };

    container.addEventListener("scroll", handleScroll);
    return () => container.removeEventListener("scroll", handleScroll);
  }, []);

  // Auto-scroll only when user is at bottom and content changes
  useEffect(() => {
    if (shouldAutoScroll && messagesEndRef.current) {
      messagesEndRef.current.scrollIntoView({ behavior: "smooth" });
    }
  }, [data, streamingMessage, currentChatMessage, shouldAutoScroll]);



  const handleSendMessage = (message: string, images?: File[], imageDetail?: string) => {
    setIsUserScrolledUp(false);
    doChat(message, projectId, images, imageDetail);
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
      message_id: "progress-indicator", //this should be unique

      role: "assistant",
      content: "",
      isProgress: true
    }] : []),
    ...(streamingMessage?.trim() ? [{ message_id: currentAssistantMessageId, role: "assistant", content: streamingMessage }] : []),
  ];

  return (
    <div className="flex flex-col h-full w-full">
      {projectId && currentProject && (
        <div className="p-4 border-b border-gray-200 flex-shrink-0 bg-white">
          <div className="flex items-center gap-3">
            <div className="w-8 h-8 bg-gray-100 rounded flex items-center justify-center">
              <FileText className="size-5 text-orange-500" />
            </div>
            <h1 className="text-xl font-bold">{currentProject}</h1>
          </div>
        </div>
      )}
      <div ref={messagesContainerRef} className="flex-1 overflow-y-auto min-h-0">
        <div className={`mx-auto w-full transition-all ${isExpanded ? "max-w-7xl" : "max-w-4xl"}`}>
          {loading ? (
            <div className="flex items-center justify-center h-full text-muted-foreground">
              Loading messages...
            </div>
          ) : combinedMessages.length === 0 ? (
            <div className="flex items-center justify-center h-full text-muted-foreground">
              No messages yet
            </div>
          ) : (
            combinedMessages.map((message, index) => {
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
          
          {/* Error Display */}
          {currentChatError && (
            <div className="w-full px-4 py-4">
              <div className="bg-red-50 dark:bg-red-950/20 border border-red-200 dark:border-red-800 rounded-lg p-4 flex items-start gap-3">
                {/* Error icon */}
                <div className="flex-shrink-0 mt-0.5">
                  <svg className="w-5 h-5 text-red-600 dark:text-red-400" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                  </svg>
                </div>
                
                {/* Error content */}
                <div className="flex-1">
                  <h3 className="text-sm font-semibold text-red-800 dark:text-red-300">
                    {currentChatError.type === 0 ? '⚠️ Provider Configuration Error' : 'Error'}
                  </h3>
                  <p className="mt-1 text-sm text-red-700 dark:text-red-400">
                    {currentChatError.message}
                  </p>
                  
                  {/* Action button for provider configuration errors */}
                  {currentChatError.type === 0 && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => navigate('/models')}
                      className="mt-3 text-xs bg-red-100 dark:bg-red-900/30 border-red-300 dark:border-red-700 text-red-800 dark:text-red-300 hover:bg-red-200 dark:hover:bg-red-900/50"
                    >
                      Configure API Keys →
                    </Button>
                  )}
                </div>
                
                {/* Close button */}
                <button
                  onClick={() => $currentChatError.set(null)}
                  className="flex-shrink-0 text-red-400 hover:text-red-600 dark:hover:text-red-300 transition-colors"
                  aria-label="Dismiss error"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>
            </div>
          )}
          
          {listChatBranch.length > 0 && (
            <div className="bg-muted border-t border-border py-4 px-4">
              <div className="w-full max-w-none px-4">
                <h3 className="text-sm font-medium text-foreground mb-2">Related Chats:</h3>
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
      </div>
      <div className="mx-auto w-full max-w-4xl">
        <ChatInputBox projectId={projectId} onSendMessage={handleSendMessage} />
      </div>
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
              <div className="text-sm text-muted-foreground">Loading document details...</div>
            </div>
          )}
          {ragDocumentDetails.error && (
            <div className="text-destructive text-sm p-4 bg-destructive/10 rounded">
              Error: {ragDocumentDetails.error}
            </div>
          )}
          {ragDocumentDetails.data && (
            <div>
              <div className="text-sm text-muted-foreground mb-4">
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