import { Button } from "@/components/ui/button";
import {
    CornerDownLeft,
    Copy,
    Check,
    Loader2,
    Square,
    BotIcon,
    Wrench,
    CheckCircle2,
    XCircle,
    ChevronDown,
    ChevronRight,
    Brain,
    Clock
} from "lucide-react";
import React, { useEffect, useRef, useState } from "react";
import { useStore } from "@nanostores/react";
import { useParams } from "react-router-dom";
import {
    $agentMessages,
    $isAgentStreaming,
    $agentStreamingMessage,
    $agentStreamingEvents,
    getAgentMessages,
    sendAgentMessage,
    agentStream,
    type StreamEvent
} from "@/store/agents";
import { EnhancedMarkdown } from "@/components/enhanced-markdown";
import type { AgentMessage } from "../../proto/chatservice";

interface MessageProps {
    message: AgentMessage & { isStreaming?: boolean };
    streamEvents?: StreamEvent[];
    onCopyMessage: (content: string, messageId: string) => void;
    isCopied: boolean;
}

function ToolCallCard({ event }: { event: StreamEvent }) {
    const [isExpanded, setIsExpanded] = useState(true);

    // Parse arguments JSON
    let args = {};
    try {
        if (event.argumentsJson) {
            args = JSON.parse(event.argumentsJson);
        }
    } catch (e) {
        // Keep empty object
    }

    return (
        <div className="my-3 border border-blue-200 dark:border-blue-800 rounded-lg bg-blue-50 dark:bg-blue-950/30 overflow-hidden">
            <button
                onClick={() => setIsExpanded(!isExpanded)}
                className="w-full px-4 py-3 flex items-center justify-between hover:bg-blue-100 dark:hover:bg-blue-900/30 transition-colors"
            >
                <div className="flex items-center space-x-2">
                    <Wrench className="w-4 h-4 text-blue-600 dark:text-blue-400 animate-spin" />
                    <span className="text-[10px] font-mono bg-blue-200 dark:bg-blue-800 px-1 rounded text-blue-800 dark:text-blue-200">
                        tool_call
                    </span>
                    <span className="font-medium text-sm text-blue-900 dark:text-blue-100">
                        {event.toolName || 'Unknown'}
                    </span>
                    {event.toolCallId && (
                        <span className="text-[10px] font-mono text-blue-600 dark:text-blue-400">
                            #{event.toolCallId}
                        </span>
                    )}
                </div>
                {isExpanded ? (
                    <ChevronDown className="w-4 h-4 text-blue-600 dark:text-blue-400" />
                ) : (
                    <ChevronRight className="w-4 h-4 text-blue-600 dark:text-blue-400" />
                )}
            </button>
            {isExpanded && (
                <div className="px-4 pb-3 pt-1">
                    <div className="text-xs text-muted-foreground mb-1">Arguments:</div>
                    <pre className="text-xs bg-blue-100 dark:bg-blue-900/50 p-2 rounded overflow-x-auto">
                        <code>{JSON.stringify(args, null, 2)}</code>
                    </pre>
                </div>
            )}
        </div>
    );
}

function ToolResultCard({ event }: { event: StreamEvent }) {
    const [isExpanded, setIsExpanded] = useState(true);

    // Parse result JSON
    let result = {};
    try {
        if (event.resultJson) {
            result = JSON.parse(event.resultJson);
        }
    } catch (e) {
        result = event.resultJson || {};
    }

    const isSuccess = event.success !== false;
    const borderColor = isSuccess 
        ? "border-green-200 dark:border-green-800" 
        : "border-red-200 dark:border-red-800";
    const bgColor = isSuccess 
        ? "bg-green-50 dark:bg-green-950/30" 
        : "bg-red-50 dark:bg-red-950/30";
    const hoverBg = isSuccess 
        ? "hover:bg-green-100 dark:hover:bg-green-900/30" 
        : "hover:bg-red-100 dark:hover:bg-red-900/30";
    const textColor = isSuccess 
        ? "text-green-600 dark:text-green-400" 
        : "text-red-600 dark:text-red-400";
    const badgeBg = isSuccess 
        ? "bg-green-200 dark:bg-green-800" 
        : "bg-red-200 dark:bg-red-800";
    const badgeText = isSuccess 
        ? "text-green-800 dark:text-green-200" 
        : "text-red-800 dark:text-red-200";
    const resultBg = isSuccess 
        ? "bg-green-100 dark:bg-green-900/50" 
        : "bg-red-100 dark:bg-red-900/50";

    return (
        <div className={`my-3 border ${borderColor} rounded-lg ${bgColor} overflow-hidden`}>
            <button
                onClick={() => setIsExpanded(!isExpanded)}
                className={`w-full px-4 py-3 flex items-center justify-between ${hoverBg} transition-colors`}
            >
                <div className="flex items-center space-x-2">
                    {isSuccess ? (
                        <CheckCircle2 className={`w-4 h-4 ${textColor}`} />
                    ) : (
                        <XCircle className={`w-4 h-4 ${textColor}`} />
                    )}
                    <span className={`text-[10px] font-mono ${badgeBg} px-1 rounded ${badgeText}`}>
                        tool_result
                    </span>
                    <span className={`font-medium text-sm ${isSuccess ? "text-green-900 dark:text-green-100" : "text-red-900 dark:text-red-100"}`}>
                        {event.toolName || 'Unknown'}
                    </span>
                    {event.toolCallId && (
                        <span className={`text-[10px] font-mono ${textColor}`}>
                            #{event.toolCallId}
                        </span>
                    )}
                    {event.durationMs !== undefined && (
                        <span className="flex items-center text-[10px] font-mono text-muted-foreground">
                            <Clock className="w-3 h-3 mr-1" />
                            {event.durationMs}ms
                        </span>
                    )}
                </div>
                {isExpanded ? (
                    <ChevronDown className={`w-4 h-4 ${textColor}`} />
                ) : (
                    <ChevronRight className={`w-4 h-4 ${textColor}`} />
                )}
            </button>
            {isExpanded && (
                <div className="px-4 pb-3 pt-1">
                    <div className="text-xs text-muted-foreground mb-1">Result:</div>
                    <pre className={`text-xs ${resultBg} p-2 rounded overflow-x-auto`}>
                        <code>{JSON.stringify(result, null, 2)}</code>
                    </pre>
                    {event.errorMessage && (
                        <div className="mt-2 text-xs text-red-600 dark:text-red-400">
                            Error: {event.errorMessage}
                        </div>
                    )}
                </div>
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

    // Filter out 'text' events from eventsToRender (they are shown as main content)
    const nonTextEvents = eventsToRender.filter(e => e.type !== 'text');

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
                    <div className="flex-shrink-0 w-8 h-8 rounded-full bg-blue-600 text-white flex items-center justify-center text-sm font-medium">
                        <BotIcon className="w-5 h-5" />
                    </div>
                )}

                <div className={`flex-1 min-w-0 ${isUser ? "text-right" : "text-left"}`}>
                    {/* Stream Events (thinking, tool calls, tool results) */}
                    {!isUser && nonTextEvents.length > 0 && (
                        <div className="space-y-2 mb-4">
                            {nonTextEvents.map((event, idx) => {
                                if (event.type === 'thinking') {
                                    return <ThinkingCard key={idx} event={event} />;
                                } else if (event.type === 'tool_call') {
                                    return <ToolCallCard key={idx} event={event} />;
                                } else if (event.type === 'tool_result') {
                                    return <ToolResultCard key={idx} event={event} />;
                                } else if (event.type === 'error') {
                                    return (
                                        <div key={idx} className="my-2 px-3 py-2 bg-red-50 dark:bg-red-950/30 border border-red-200 dark:border-red-800 rounded-lg">
                                            <div className="flex items-center space-x-2 text-xs text-red-700 dark:text-red-300">
                                                <span className="text-[10px] font-mono bg-red-200 dark:bg-red-800 px-1 rounded text-red-800 dark:text-red-200">
                                                    error
                                                </span>
                                                <span>❌ {event.text}</span>
                                            </div>
                                        </div>
                                    );
                                } else if (event.type === 'image' && event.url) {
                                    return (
                                        <div key={idx} className="my-2">
                                            <img 
                                                src={event.url} 
                                                alt="Generated image" 
                                                className="max-w-md rounded-lg border border-border"
                                            />
                                        </div>
                                    );
                                }
                                return null;
                            })}
                        </div>
                    )}

                    {/* Main message content */}
                    {message.content && (
                        <div className="space-y-2">
                            <EnhancedMarkdown>{message.content}</EnhancedMarkdown>
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
        if (sessionId) {
            getAgentMessages(sessionId);
        }
    }, [sessionId]);

    // Auto-scroll to bottom
    useEffect(() => {
        if (messagesEndRef.current) {
            messagesEndRef.current.scrollIntoView({ behavior: "smooth" });
        }
    }, [messages, streamingMessage]);

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
        <div className="flex flex-col h-[calc(100vh-theme(spacing.16))] md:h-screen w-full bg-background">
            {/* Messages Area */}
            <div className="flex-1 overflow-y-auto w-full">
                {loading ? (
                    <div className="flex items-center justify-center h-full">
                        <Loader2 className="h-6 w-6 animate-spin text-primary" />
                    </div>
                ) : messages.length === 0 && !isStreaming ? (
                    <div className="flex items-center justify-center h-full text-muted-foreground">
                        No messages yet. Start a conversation!
                    </div>
                ) : (
                    <div className="flex flex-col w-full max-w-none">
                        {messages.map((msg, index) => (
                            <Message
                                key={msg.id || index}
                                message={msg}
                                streamEvents={msg.streamEvents}
                                onCopyMessage={handleCopyMessage}
                                isCopied={copiedMessageId === msg.id}
                            />
                        ))}

                        {/* Streaming Message */}
                        {isStreaming && (streamingMessage || streamingEvents.length > 0) && (
                            <Message
                                message={{
                                    role: "assistant",
                                    content: streamingMessage,
                                    type: "text",
                                    id: "streaming",
                                    sequence_number: 0,
                                    created_at: 0
                                } as unknown as AgentMessage}
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
