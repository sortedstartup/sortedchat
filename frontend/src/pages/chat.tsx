import { Button } from "@/components/ui/button";
import {
  ChatBubble,
  ChatBubbleAvatar,
  ChatBubbleMessage,
} from "@/components/ui/chat/chat-bubble";
import { ChatInput } from "@/components/ui/chat/chat-input";
import { ChatMessageList } from "@/components/ui/chat/chat-message-list";
import { CornerDownLeft, Mic } from "lucide-react";
import { useEffect, useRef, useState } from "react";
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
  $innerChatList,
} from "@/store/chat";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
} from "@/components/ui/dropdown-menu";

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
  const innerChatList = useStore($innerChatList);

  const [inputValue, setInputValue] = useState("");

  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    scrollToBottom();
  }, [data, streamingMessage, currentChatMessage]);

  const handleSend = () => {
    if (inputValue.trim()) {
      doChat(inputValue,projectId);
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
              {data?.map((message) => (
                <div
                  key={message.message_id}
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
                  {message.role === "assistant" && (
                    <div className="ml-2 sm:ml-4 mt-2">
                      <Button 
                        variant="outline" 
                        size="sm" 
                        onClick={() => BranchChat(message.message_id || "")}
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
              )}
              <div ref={messagesEndRef} />
            </>
          )}
        </ChatMessageList>
        
        {/* Inner Chat List */}
        {innerChatList.length > 0 && (
          <div className="mt-4 px-2 sm:px-4">
            <h3 className="text-sm font-medium text-gray-700 mb-2">Related Chats:</h3>
            <div className="flex flex-wrap gap-2">
              {innerChatList.map((chat) => (
                <Button
                  key={chat.chatId}
                  variant="outline"
                  size="sm"
                  onClick={() => navigate(`/chat/${chat.chatId}`)}
                  className="text-xs"
                >
                  {chat.name}
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
    </div>
  );
}