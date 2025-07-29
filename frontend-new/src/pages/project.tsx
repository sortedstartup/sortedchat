import { useState } from "react";
import { Paperclip, Mic, CornerDownLeft, FileText } from "lucide-react";
import { Button } from "@/components/ui/button";
import { ChatInput } from "@/components/ui/chat/chat-input";

export function Project() {
  const [message, setMessage] = useState("");

  const handleSendMessage = () => {
    if (message.trim()) {
      console.log("Sending message:", message);
      setMessage("");
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  return (
    <div
      className="flex flex-col min-h-screen"
      style={{
        minHeight: "100vh",
        marginLeft: "1rem",
        marginRight: "1rem",
        width: "calc(100vw - 16rem)",
      }}
    >
      <div className="p-4 border-b border-gray-200">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 bg-gray-100 rounded flex items-center justify-center">
            <FileText className="size-5 text-orange-500" />
          </div>
          <h1 className="text-xl font-bold">Spy Novel</h1>
        </div>
      </div>

      <div className="flex-1 flex items-center justify-center p-6">
        <div className="text-center max-w-md">
          <h2 className="text-xl font-medium mb-3">Start your spy novel</h2>
          <p className="text-gray-500 mb-6 text-sm">
            Begin writing your sophisticated thriller below
          </p>
          <Button variant="outline" size="sm">
            <FileText className="size-4 mr-2" />
            Project Documents
          </Button>
        </div>
      </div>

      <div className="bg-white p-4 border-t border-gray-200">
        <div className="relative rounded-lg border border-gray-200 bg-gray-50 focus-within:ring-1 focus-within:ring-orange-500 p-1">
          <ChatInput
            placeholder="Type your message here..."
            className="min-h-12 border-0 p-3 shadow-none focus-visible:ring-0 bg-gray-50 text-black"
            value={message}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) => setMessage(e.target.value)}
            onKeyDown={handleKeyDown}
          />
          <div className="flex items-center p-3 pt-0">
            {/* <Button variant="ghost" size="icon" type="button">
              <Paperclip className="size-5 text-gray-400" />
              <span className="sr-only">Attach file</span>
            </Button>

            <Button variant="ghost" size="icon" type="button">
              <Mic className="size-5 text-gray-400" />
              <span className="sr-only">Use Microphone</span>
            </Button> */}
            <Button
              size="sm"
              className="ml-auto gap-1.5"
              onClick={handleSendMessage}
              type="button"
            >
              Send Message
              <CornerDownLeft className="size-4" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
