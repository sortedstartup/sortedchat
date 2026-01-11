import { Button } from "@/components/ui/button";
import {
    CornerDownLeft,
    Copy,
    Check,
    Loader2,
    Square,
    BotIcon,
    CheckCircle2,
    XCircle,
    ChevronDown,
    ChevronRight,
    Brain,
    FileText,
    Code2,
    ExternalLink
} from "lucide-react";
import React, { useEffect, useRef, useState, useMemo } from "react";
import { useStore } from "@nanostores/react";
import { useParams } from "react-router-dom";
import {
    $agentMessages,
    $isAgentStreaming,
    $agentStreamingMessage,
    $agentStreamingEvents,
    $currentSessionId,
    getAgentMessages,
    sendAgentMessage,
    agentStream,
    type StreamEvent
} from "@/store/agents";
import { EnhancedMarkdown } from "@/components/enhanced-markdown";
import { type AgentMessage, AgentChatErrorType } from "../../proto/chatservice";

interface MessageProps {
    message: AgentMessage & { isStreaming?: boolean };
    streamEvents?: StreamEvent[];
    onCopyMessage: (content: string, messageId: string) => void;
    isCopied: boolean;
}

// Combined tool execution info
interface ToolExecution {
    toolCallId?: string;
    toolName: string;
    argumentsJson?: string;
    resultJson?: string;
    success?: boolean;
    errorMessage?: string;
    durationMs?: number;
    isComplete: boolean;
}

// Compact expandable code block with line limit
function ExpandableCode({ content, defaultLines = 50 }: { content: string; defaultLines?: number }) {
    const [isFullyExpanded, setIsFullyExpanded] = useState(false);
    const lines = content.split('\n');
    const needsTruncation = lines.length > defaultLines;

    const displayContent = needsTruncation && !isFullyExpanded
        ? lines.slice(0, defaultLines).join('\n') + '\n...'
        : content;

    return (
        <div>
            <pre className="text-xs bg-muted/50 p-2 rounded overflow-x-auto max-h-[300px] overflow-y-auto">
                <code>{displayContent}</code>
            </pre>
            {needsTruncation && (
                <button
                    onClick={() => setIsFullyExpanded(!isFullyExpanded)}
                    className="mt-1 text-xs text-primary hover:underline"
                >
                    {isFullyExpanded ? 'Show less' : `Show all ${lines.length} lines`}
                </button>
            )}
        </div>
    );
}

// Compact inline tool chip - clickable to expand
function ToolChip({ execution, isExpanded, onToggle }: {
    execution: ToolExecution;
    isExpanded: boolean;
    onToggle: () => void;
}) {
    const isSuccess = execution.success !== false;
    const isComplete = execution.isComplete;

    const getBgColor = () => {
        if (!isComplete) return "bg-muted hover:bg-muted/80";
        if (isSuccess) return "bg-green-100 dark:bg-green-900/40 hover:bg-green-200 dark:hover:bg-green-900/60";
        return "bg-red-100 dark:bg-red-900/40 hover:bg-red-200 dark:hover:bg-red-900/60";
    };

    const getIconColor = () => {
        if (!isComplete) return "text-muted-foreground";
        if (isSuccess) return "text-green-600 dark:text-green-400";
        return "text-red-600 dark:text-red-400";
    };

    return (
        <button
            onClick={onToggle}
            className={`inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs ${getBgColor()} transition-colors`}
        >
            {!isComplete ? (
                <Loader2 className={`w-3 h-3 animate-spin ${getIconColor()}`} />
            ) : isSuccess ? (
                <CheckCircle2 className={`w-3 h-3 ${getIconColor()}`} />
            ) : (
                <XCircle className={`w-3 h-3 ${getIconColor()}`} />
            )}
            <span className="font-medium">{execution.toolName}</span>
            {isComplete && execution.durationMs !== undefined && (
                <span className="text-muted-foreground text-[10px]">{execution.durationMs}ms</span>
            )}
            {isExpanded ? (
                <ChevronDown className="w-3 h-3 text-muted-foreground" />
            ) : (
                <ChevronRight className="w-3 h-3 text-muted-foreground" />
            )}
        </button>
    );
}

// Expanded detail panel for a tool execution
function ToolDetailPanel({ execution }: { execution: ToolExecution }) {
    // Parse arguments and result
    let args = {};
    let result = {};
    try {
        if (execution.argumentsJson) {
            args = JSON.parse(execution.argumentsJson);
        }
    } catch (e) { /* ignore */ }

    try {
        if (execution.resultJson) {
            result = JSON.parse(execution.resultJson);
        }
    } catch (e) {
        result = execution.resultJson || {};
    }

    return (
        <div className="mt-2 p-2 border border-border rounded-lg bg-muted/20 text-left">
            <div className="text-xs font-medium mb-1">{execution.toolName}</div>

            {/* Arguments */}
            <div className="mb-2">
                <div className="text-[10px] text-muted-foreground mb-0.5">Arguments:</div>
                <ExpandableCode content={JSON.stringify(args, null, 2)} defaultLines={50} />
            </div>

            {/* Result (if complete) */}
            {execution.isComplete && (
                <div>
                    <div className="text-[10px] text-muted-foreground mb-0.5">Result:</div>
                    <ExpandableCode content={JSON.stringify(result, null, 2)} defaultLines={50} />
                </div>
            )}

            {/* Error message */}
            {execution.errorMessage && (
                <div className="mt-1 text-xs text-red-600 dark:text-red-400">
                    Error: {execution.errorMessage}
                </div>
            )}
        </div>
    );
}

// Container for multiple tool chips in a horizontal row
function ToolExecutionsRow({ executions }: { executions: ToolExecution[] }) {
    const [expandedIndex, setExpandedIndex] = useState<number | null>(null);

    const handleToggle = (index: number) => {
        setExpandedIndex(expandedIndex === index ? null : index);
    };

    return (
        <div className="mb-3">
            {/* Horizontal row of chips */}
            <div className="flex flex-wrap gap-1.5">
                {executions.map((execution, idx) => (
                    <ToolChip
                        key={execution.toolCallId || idx}
                        execution={execution}
                        isExpanded={expandedIndex === idx}
                        onToggle={() => handleToggle(idx)}
                    />
                ))}
            </div>

            {/* Expanded detail panel (shown below chips) */}
            {expandedIndex !== null && executions[expandedIndex] && (
                <ToolDetailPanel execution={executions[expandedIndex]} />
            )}
        </div>
    );
}

function ThinkingCard({ event }: { event: StreamEvent }) {
    return (
        <div className="my-2 px-3 py-2 bg-purple-50 dark:bg-purple-950/30 border border-purple-200 dark:border-purple-800 rounded-lg">
            <div className="flex items-center space-x-2">
                <Brain className="w-3 h-3 text-purple-600 dark:text-purple-400 animate-pulse" />
                <span className="text-[10px] font-mono bg-purple-200 dark:bg-purple-800 px-1 rounded text-purple-800 dark:text-purple-200">
                    thinking
                </span>
                {event.phase && (
                    <span className="text-[10px] font-mono bg-purple-100 dark:bg-purple-900 px-1 rounded text-purple-600 dark:text-purple-400">
                        {event.phase}
                    </span>
                )}
                <span className="text-xs text-purple-700 dark:text-purple-300 italic">
                    {event.text}
                </span>
                {event.model && (
                    <span className="text-[10px] font-mono text-purple-500 dark:text-purple-500">
                        [{event.model}]
                    </span>
                )}
            </div>
        </div>
    );
}

function ModelLoadingNotice() {
    const [isVisible, setIsVisible] = useState(true);

    useEffect(() => {
        const timer = setTimeout(() => setIsVisible(false), 5000);
        return () => clearTimeout(timer);
    }, []);

    if (!isVisible) return null;

    return (
        <div className="my-2 px-4 py-3 bg-blue-50 dark:bg-blue-950/30 border border-blue-200 dark:border-blue-800 rounded-xl animate-in fade-in slide-in-from-top-2 duration-500">
            <div className="flex items-center space-x-3">
                <div className="flex-shrink-0">
                    <Loader2 className="w-5 h-5 text-blue-600 dark:text-blue-400 animate-spin" />
                </div>
                <div className="flex flex-col">
                    <span className="text-sm font-semibold text-blue-900 dark:text-blue-100">
                        Initializing Model
                    </span>
                    <span className="text-xs text-blue-700 dark:text-blue-300">
                        The AI model is being loaded. This usually takes a few seconds.
                    </span>
                </div>
            </div>
        </div>
    );
}

// HTML File Preview Component
function HtmlFilePreview({ event }: { event: StreamEvent }) {
    const [showPreview, setShowPreview] = useState(false);
    const [isExpanded, setIsExpanded] = useState(false);
    const [enableScripts, setEnableScripts] = useState(true);

    const formatFileSize = (bytes?: number) => {
        if (!bytes) return '';
        if (bytes < 1024) return bytes + " B";
        if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
        return (bytes / (1024 * 1024)).toFixed(1) + " MB";
    };

    const openInNewTab = () => {
        const blob = new Blob([event.fileContent || ''], { type: 'text/html' });
        const url = URL.createObjectURL(blob);
        window.open(url, '_blank');
        // Clean up after a delay
        setTimeout(() => URL.revokeObjectURL(url), 100);
    };

    const sandboxPermissions = enableScripts
        ? "allow-same-origin allow-scripts"
        : "allow-same-origin";

    return (
        <div className="my-2 border border-border rounded-lg overflow-hidden bg-card">
            {/* Header */}
            <div className="flex items-center justify-between p-3 bg-muted/50">
                <div className="flex items-center gap-2">
                    <FileText className="w-4 h-4 text-blue-600 dark:text-blue-400" />
                    <span className="text-sm font-medium">{event.fileName}</span>
                    {event.fileSize && (
                        <span className="text-xs text-muted-foreground">
                            ({formatFileSize(event.fileSize)})
                        </span>
                    )}
                    <span className="text-xs bg-blue-100 dark:bg-blue-900/40 text-blue-800 dark:text-blue-200 px-2 py-0.5 rounded">
                        HTML
                    </span>
                </div>
                <div className="flex items-center gap-2">
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={openInNewTab}
                        className="h-7 gap-1"
                        title="Open in new tab"
                    >
                        <ExternalLink className="w-3 h-3" />
                    </Button>
                    <Button
                        variant={showPreview ? "default" : "outline"}
                        size="sm"
                        onClick={() => setShowPreview(!showPreview)}
                        className="h-7"
                    >
                        {showPreview ? 'Hide Preview' : 'Preview'}
                    </Button>
                    <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setIsExpanded(!isExpanded)}
                        className="h-7 gap-1"
                    >
                        <Code2 className="w-3 h-3" />
                        {isExpanded ? <ChevronDown className="w-3 h-3" /> : <ChevronRight className="w-3 h-3" />}
                    </Button>
                </div>
            </div>

            {/* Preview iframe */}
            {showPreview && (
                <div className="border-t border-border">
                    <div className="flex items-center justify-between px-3 py-2 bg-yellow-50 dark:bg-yellow-950/30 border-b border-yellow-200 dark:border-yellow-800">
                        <span className="text-xs text-yellow-800 dark:text-yellow-200">
                            Preview is sandboxed for security
                        </span>
                        <label className="flex items-center gap-2 text-xs cursor-pointer">
                            <input
                                type="checkbox"
                                checked={enableScripts}
                                onChange={(e) => setEnableScripts(e.target.checked)}
                                className="rounded"
                            />
                            <span className="text-yellow-800 dark:text-yellow-200">Enable JavaScript</span>
                        </label>
                    </div>
                    <iframe
                        key={enableScripts ? 'with-scripts' : 'no-scripts'}
                        srcDoc={event.fileContent}
                        sandbox={sandboxPermissions}
                        className="w-full h-96 bg-white"
                        title={`Preview: ${event.fileName}`}
                    />
                </div>
            )}

            {/* Code view (collapsible) */}
            {isExpanded && (
                <div className="border-t border-border p-3 bg-muted/20">
                    <div className="text-xs text-muted-foreground mb-2">HTML Source:</div>
                    <ExpandableCode content={event.fileContent || ''} defaultLines={20} />
                </div>
            )}
        </div>
    );
}

function Message({
    message,
    streamEvents,
    onCopyMessage,
    isCopied,
}: MessageProps) {
    const isUser = message.role === "user";

    // Prepare events to render: either from streamEvents or from message fields (for persisted messages)
    const eventsToRender: StreamEvent[] = [...(streamEvents || [])];

    // If no stream events but it's a tool-related message, create a synthetic event
    if (eventsToRender.length === 0) {
        if (message.type === 'tool_call') {
            eventsToRender.push({
                type: 'tool_call',
                timestamp: 0,
                toolName: message.tool_name,
                argumentsJson: message.tool_args,
            });
        } else if (message.type === 'tool_result') {
            eventsToRender.push({
                type: 'tool_result',
                timestamp: 0,
                toolName: message.tool_name,
                resultJson: message.content,
                success: true,
            });
        }
    }

    // Filter out 'text' and 'thinking' events (text shown as main content, thinking is noise)
    const toolEvents = eventsToRender.filter(e => e.type === 'tool_call' || e.type === 'tool_result');
    const errorEvents = eventsToRender.filter(e => e.type === 'error');
    const imageEvents = eventsToRender.filter(e => e.type === 'image');

    // Extract HTML files from write_file tool calls
    const htmlFileEvents: StreamEvent[] = [];
    toolEvents.forEach(event => {
        if (event.toolName === 'write_file' && event.type === 'tool_call') {
            try {
                const args = event.argumentsJson ? JSON.parse(event.argumentsJson) : {};
                const filePath = args.path || args.file_path || '';
                const content = args.content || '';
                const ext = filePath.split('.').pop()?.toLowerCase();

                if ((ext === 'html' || ext === 'htm') && content) {
                    htmlFileEvents.push({
                        type: 'write_file',
                        timestamp: event.timestamp,
                        fileName: filePath.split('/').pop() || filePath,
                        filePath: filePath,
                        fileContent: content,
                        fileSize: content.length,
                    });
                }
            } catch (e) {
                // Ignore parse errors
            }
        }
    });

    // Combine tool_call and tool_result events into unified executions
    const toolExecutions = useMemo(() => {
        const executions: ToolExecution[] = [];
        const callMap = new Map<string, ToolExecution>();

        for (const event of toolEvents) {
            const key = event.toolCallId || event.toolName || 'unknown';

            if (event.type === 'tool_call') {
                const execution: ToolExecution = {
                    toolCallId: event.toolCallId,
                    toolName: event.toolName || 'Unknown',
                    argumentsJson: event.argumentsJson,
                    isComplete: false
                };
                callMap.set(key, execution);
                executions.push(execution);
            } else if (event.type === 'tool_result') {
                // Find matching call or create new entry
                let execution = callMap.get(key);
                if (execution) {
                    // Update existing execution with result
                    execution.resultJson = event.resultJson;
                    execution.success = event.success;
                    execution.errorMessage = event.errorMessage;
                    execution.durationMs = event.durationMs;
                    execution.isComplete = true;
                } else {
                    // Result without matching call (shouldn't happen but handle gracefully)
                    executions.push({
                        toolCallId: event.toolCallId,
                        toolName: event.toolName || 'Unknown',
                        resultJson: event.resultJson,
                        success: event.success,
                        errorMessage: event.errorMessage,
                        durationMs: event.durationMs,
                        isComplete: true
                    });
                }
            }
        }

        return executions;
    }, [toolEvents]);

    const hasActiveTool = toolExecutions.some(e => !e.isComplete);

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
                    <div className={`flex-shrink-0 w-8 h-8 rounded-full ${message.isStreaming ? 'bg-purple-600' : 'bg-blue-600'} text-white flex items-center justify-center text-sm font-medium transition-colors relative`}>
                        {message.isStreaming ? (
                            <Brain className="w-5 h-5 animate-pulse" />
                        ) : (
                            <BotIcon className="w-5 h-5" />
                        )}
                    </div>
                )}

                <div className={`flex-1 min-w-0 ${isUser ? "text-right" : "text-left"}`}>
                    {/* Tool Executions - horizontal chips */}
                    {!isUser && toolExecutions.length > 0 && (
                        <ToolExecutionsRow executions={toolExecutions} />
                    )}

                    {/* Error Events */}
                    {!isUser && errorEvents.length > 0 && (
                        <div className="space-y-2 mb-4">
                            {errorEvents.map((event, idx) => {
                                const isModelLoading = event.errorType === AgentChatErrorType.MODEL_LOADING || 
                                                     (event.text?.toLowerCase().includes("500") && 
                                                      event.text?.toLowerCase().includes("loading model"));
                                
                                if (isModelLoading) {
                                    return <ModelLoadingNotice key={idx} />;
                                }

                                return (
                                    <div key={idx} className="my-2 px-3 py-2 bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800 rounded-lg">
                                        <div className="flex items-center space-x-2 text-xs text-red-700 dark:text-red-300">
                                            <XCircle className="w-4 h-4" />
                                            <span>{event.text}</span>
                                        </div>
                                    </div>
                                );
                            })}
                        </div>
                    )}

                    {/* Image Events */}
                    {!isUser && imageEvents.length > 0 && (
                        <div className="space-y-2 mb-4">
                            {imageEvents.map((event, idx) => (
                                <div key={idx} className="my-2">
                                    <img
                                        src={event.url}
                                        alt="Generated image"
                                        className="max-w-md rounded-lg border border-border"
                                    />
                                </div>
                            ))}
                        </div>
                    )}

                    {/* HTML File Events */}
                    {!isUser && htmlFileEvents.length > 0 && (
                        <div className="space-y-2 mb-4">
                            {htmlFileEvents.map((event, idx) => (
                                <HtmlFilePreview key={idx} event={event} />
                            ))}
                        </div>
                    )}

                    {/* Main message content */}
                    {message.content ? (
                        <div className="space-y-2">
                            <EnhancedMarkdown>{message.content}</EnhancedMarkdown>
                            {message.isStreaming && (
                                <span className="inline-block w-1.5 h-4 bg-primary/40 animate-pulse ml-1 align-middle" />
                            )}
                        </div>
                    ) : message.isStreaming && !hasActiveTool && (
                        <div className="flex items-center gap-2 text-muted-foreground text-sm italic py-2">
                            <Loader2 className="w-4 h-4 animate-spin" />
                            <span>Agent is thinking...</span>
                        </div>
                    )}

                    {!isUser && message.content && (
                        <div className="flex justify-between mt-3">
                            <div className="flex items-center space-x-2">
                                <Button
                                    variant="ghost"
                                    size="sm"
                                    onClick={() => onCopyMessage(message.content, message.id)}
                                    className="h-8 px-2 text-xs text-muted-foreground hover:text-foreground"
                                >
                                    {isCopied ? (
                                        <Check className="h-4 w-4 text-green-500" />
                                    ) : (
                                        <Copy className="h-3 w-3" />
                                    )}
                                </Button>
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
    onSendMessage,
    isStreaming
}: {
    onSendMessage: (message: string) => void;
    isStreaming: boolean;
}) {
    const [inputValue, setInputValue] = useState("");
    const textareaRef = useRef<HTMLTextAreaElement>(null);

    // Auto-resize textarea
    useEffect(() => {
        const textarea = textareaRef.current;
        if (textarea) {
            textarea.style.height = 'auto';
            const newHeight = Math.min(Math.max(textarea.scrollHeight, 48), 200);
            textarea.style.height = `${newHeight}px`;
        }
    }, [inputValue]);

    const handleSend = () => {
        if (inputValue.trim() && !isStreaming) {
            onSendMessage(inputValue);
            setInputValue("");
        }
    };

    const handleStop = () => {
        if (agentStream) {
            agentStream.cancel();
        }
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
        if (e.key === "Enter" && !e.shiftKey && !isStreaming) {
            e.preventDefault();
            handleSend();
        }
    };

    return (
        <div className="flex-shrink-0 bg-card border-t border-border p-4">
            <div className="w-full max-w-none px-4">
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
                    />
                    <div className="flex items-center justify-between p-3 pt-0">
                        <div className="flex items-center space-x-2">
                            {/* Optional: Add tools button here later */}
                        </div>

                        {isStreaming ? (
                            <Button
                                size="sm"
                                variant="destructive"
                                className="px-4"
                                onClick={handleStop}
                            >
                                <Square className="size-3.5" />
                            </Button>
                        ) : (
                            <Button
                                size="sm"
                                className="bg-primary hover:bg-primary/90 text-primary-foreground px-4"
                                onClick={handleSend}
                                disabled={!inputValue.trim()}
                            >
                                <CornerDownLeft className="size-3.5" />
                            </Button>
                        )}
                    </div>
                </div>
            </div>
        </div>
    );
}

export function AgentChat() {
    const { agentId, sessionId } = useParams();
    const [copiedMessageId, setCopiedMessageId] = useState<string | null>(null);

    const { data: messages, loading } = useStore($agentMessages);
    const isStreaming = useStore($isAgentStreaming);
    const streamingMessage = useStore($agentStreamingMessage);
    const streamingEvents = useStore($agentStreamingEvents);

    const messagesEndRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        if (sessionId && sessionId !== $currentSessionId.get()) {
            $currentSessionId.set(sessionId);
        } else if (sessionId) {
            // Ensure messages are loaded if we navigated back to same session or store was cleared
            // But relying on listener is safer.
            // If store already has this sessionId, listener won't fire if value didn't change?
            // Nanostores .set() triggers listeners even if value is same? 
            // Default atom does check equality.
            // So if we are already on this session, we might not refetch.
            // But if we navigated away, unmounted, and came back? 
            // The store persists in memory (single page app).
            // If we want to ensure fresh data, we might want to fetch anyway or rely on store state.
            // Let's assume store state is valid if set.
            // But for initial load/refresh, we might need to force fetch?
            // Actually, if we just mounted, we might want to refresh.
            // Let's call getAgentMessages if data is empty?
            if ($agentMessages.get().data.length === 0) {
                 getAgentMessages(sessionId);
            }
        }
    }, [sessionId]);

    // Auto-scroll to bottom
    useEffect(() => {
        if (messagesEndRef.current) {
            messagesEndRef.current.scrollIntoView({ behavior: "smooth" });
        }
    }, [messages, streamingMessage]);

    // Process messages to combine tool_call/tool_result with adjacent assistant messages
    const processedMessages = useMemo(() => {
        const result: Array<typeof messages[0] & { streamEvents?: StreamEvent[] }> = [];
        let pendingToolEvents: StreamEvent[] = [];

        for (let i = 0; i < messages.length; i++) {
            const msg = messages[i];

            if (msg.type === 'tool_call') {
                // Add to pending tool events
                pendingToolEvents.push({
                    type: 'tool_call',
                    timestamp: msg.created_at || 0,
                    toolCallId: msg.tool_call_id || undefined,
                    toolName: msg.tool_name || 'Unknown',
                    argumentsJson: msg.tool_args || undefined,
                });
            } else if (msg.type === 'tool_result') {
                // Add to pending tool events - mark as complete
                pendingToolEvents.push({
                    type: 'tool_result',
                    timestamp: msg.created_at || 0,
                    toolCallId: msg.tool_call_id || undefined,
                    toolName: msg.tool_name || 'Unknown',
                    resultJson: msg.content || undefined,
                    success: msg.success !== undefined ? msg.success : true,
                    errorMessage: msg.error_message || undefined,
                    durationMs: msg.run_time_ms || undefined,
                });
            } else {
                // Regular message (user or assistant text)
                // Attach any pending tool events to assistant messages
                if (msg.role === 'assistant' && pendingToolEvents.length > 0) {
                    result.push({
                        ...msg,
                        streamEvents: [...pendingToolEvents, ...(msg.streamEvents || [])],
                    });
                    pendingToolEvents = [];
                } else if (msg.role === 'user' && pendingToolEvents.length > 0) {
                    // If user message comes before assistant, create a synthetic assistant message for tools
                    result.push({
                        id: `tool-group-${i}`,
                        role: 'assistant',
                        content: '',
                        type: 'text',
                        sequence_number: msg.sequence_number - 0.5,
                        created_at: msg.created_at,
                        streamEvents: pendingToolEvents,
                    } as typeof messages[0] & { streamEvents?: StreamEvent[] });
                    pendingToolEvents = [];
                    result.push(msg);
                } else {
                    result.push(msg);
                }
            }
        }

        // Handle any remaining tool events (show at the end)
        if (pendingToolEvents.length > 0) {
            result.push({
                id: `tool-group-end`,
                role: 'assistant',
                content: '',
                type: 'text',
                sequence_number: 9999,
                created_at: Date.now(),
                streamEvents: pendingToolEvents,
            } as typeof messages[0] & { streamEvents?: StreamEvent[] });
        }

        return result;
    }, [messages]);

    const handleSendMessage = (content: string) => {
        if (sessionId) {
            sendAgentMessage(sessionId, content);
        }
    };

    const handleCopyMessage = async (content: string, messageId: string) => {
        await navigator.clipboard.writeText(content);
        setCopiedMessageId(messageId);
        setTimeout(() => setCopiedMessageId(null), 2000);
    };

    if (!sessionId || !agentId) {
        return <div className="flex items-center justify-center h-full">Select a session to start chatting</div>;
    }

    return (
        <div className="flex flex-col h-full w-full bg-background">
            {/* Messages Area */}
            <div className="flex-1 overflow-y-auto w-full min-h-0">
                {loading ? (
                    <div className="flex items-center justify-center h-full">
                        <Loader2 className="h-6 w-6 animate-spin text-primary" />
                    </div>
                ) : processedMessages.length === 0 && !isStreaming ? (
                    <div className="flex items-center justify-center h-full text-muted-foreground">
                        No messages yet. Start a conversation!
                    </div>
                ) : (
                    <div className="flex flex-col w-full max-w-none">
                        {processedMessages.map((msg, index) => (
                            <Message
                                key={msg.id || index}
                                message={msg}
                                streamEvents={msg.streamEvents}
                                onCopyMessage={handleCopyMessage}
                                isCopied={copiedMessageId === msg.id}
                            />
                        ))}

                        {/* Streaming Message */}
                        {isStreaming && (
                            <Message
                                message={{
                                    role: "assistant",
                                    content: streamingMessage,
                                    type: "text",
                                    id: "streaming",
                                    sequence_number: 0,
                                    created_at: 0,
                                    isStreaming: true
                                } as any}
                                streamEvents={streamingEvents}
                                onCopyMessage={() => { }}
                                isCopied={false}
                            />
                        )}

                        <div ref={messagesEndRef} />
                    </div>
                )}
            </div>

            {/* Input Area */}
            <ChatInputBox
                onSendMessage={handleSendMessage}
                isStreaming={isStreaming}
            />
        </div>
    );
}
