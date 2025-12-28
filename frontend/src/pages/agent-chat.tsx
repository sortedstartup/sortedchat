import { Button } from "@/components/ui/button";
import {
    CornerDownLeft,
    Copy,
    Check,
    Loader2,
    Square,
    BotIcon
} from "lucide-react";
import React, { useEffect, useRef, useState } from "react";
import { useStore } from "@nanostores/react";
import { useParams } from "react-router-dom";
import {
    $agentMessages,
    $isAgentStreaming,
    $agentStreamingMessage,
    getAgentMessages,
    sendAgentMessage,
    agentStream
} from "@/store/agents";
import { EnhancedMarkdown } from "@/components/enhanced-markdown";
import type { AgentMessage } from "../../proto/chatservice";

interface MessageProps {
    message: AgentMessage & { isStreaming?: boolean };
    onCopyMessage: (content: string, messageId: string) => void;
    isCopied: boolean;
}

function Message({
    message,
    onCopyMessage,
    isCopied,
}: MessageProps) {
    const isUser = message.role === "user";

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
                    <div className="space-y-2">
                        <EnhancedMarkdown>{message.content}</EnhancedMarkdown>
                    </div>

                    {!isUser && (
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
                                onCopyMessage={handleCopyMessage}
                                isCopied={copiedMessageId === msg.id}
                            />
                        ))}

                        {/* Streaming Message */}
                        {isStreaming && streamingMessage && (
                            <Message
                                message={{
                                    role: "assistant",
                                    content: streamingMessage,
                                    type: "text",
                                    id: "streaming",
                                    sequence_number: 0,
                                    created_at: 0
                                } as unknown as AgentMessage}
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
