import { Button } from "@/components/ui/button";
import {
  ChatBubble,
  ChatBubbleAvatar,
  ChatBubbleMessage,
} from "@/components/ui/chat/chat-bubble";
import { ChatInput } from "@/components/ui/chat/chat-input";
import { ChatMessageList } from "@/components/ui/chat/chat-message-list";
import { CornerDownLeft, Mic, Paperclip } from "lucide-react";

export function Chat() {
  return (
    <div className="flex flex-col mx-auto max-w-full sm:max-w-2xl md:max-w-3xl lg:max-w-4xl xl:max-w-5xl w-full h-screen">
      <div className="flex-1 overflow-y-auto px-2 sm:px-4">
        <ChatMessageList className="flex flex-col gap-4 py-4 min-h-full">
          <div className="flex justify-end">
            <ChatBubble variant="sent" className="max-w-[75%] sm:max-w-sm md:max-w-md lg:max-w-lg mr-2 sm:mr-4">
              <ChatBubbleAvatar fallback="US" />
              <ChatBubbleMessage variant="sent">
                Hello, how has your day been? I hope you are doing well.
              </ChatBubbleMessage>
            </ChatBubble>
          </div>

          <div className="flex justify-start">
            <ChatBubble variant="received" className="max-w-[75%] sm:max-w-sm md:max-w-md lg:max-w-lg ml-2 sm:ml-4">
              <ChatBubbleAvatar fallback="AI" />
              <ChatBubbleMessage variant="received">
                Hi, I am doing well, thank you for asking. How can I help you
                today?
              </ChatBubbleMessage>
            </ChatBubble>
          </div>

          <div className="flex justify-start">
            <ChatBubble variant="received" className="max-w-[75%] sm:max-w-sm md:max-w-md lg:max-w-lg ml-2 sm:ml-4">
              <ChatBubbleAvatar fallback="AI" />
              <ChatBubbleMessage isLoading />
            </ChatBubble>
          </div>
        </ChatMessageList>
      </div>
      <div className="sticky bottom-0 bg-background p-2 sm:p-4 border-t">
        <div className="relative rounded-lg border bg-background focus-within:ring-1 focus-within:ring-ring p-1">
          <ChatInput
            placeholder="Type your message here..."
            className="min-h-12 resize-none rounded-lg bg-background border-0 p-3 shadow-none focus-visible:ring-0"
          />
          <div className="flex items-center p-3 pt-0">
            <Button variant="ghost" size="icon">
              <Paperclip className="size-4" />
              <span className="sr-only">Attach file</span>
            </Button>

            <Button variant="ghost" size="icon">
              <Mic className="size-4" />
              <span className="sr-only">Use Microphone</span>
            </Button>

            <Button size="sm" className="ml-auto gap-1.5">
              Send Message
              <CornerDownLeft className="size-3.5" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}