import { useState } from "react";
import { Button } from "@/components/ui/button";
import { ChatInput } from "@/components/ui/chat/chat-input";
import { CornerDownLeft } from "lucide-react";
// import { createNewChat, $currentChatId, doChat } from "@/store/chat";
import { useNavigate } from "react-router-dom";
import {DoChat,CreateNewChat}  from "../../wailsjs/go/main/App";
import { $currentChatMessage, $streamingMessage } from "@/store/chat";

export function Home() {
  const [message, setMessage] = useState("");
  const navigate = useNavigate();

  // useEffect(() => {
  //   // Initial greeting message
  //   Greet("hi").then((greeting) => {
  //     console.log(greeting);
  //   });
  //   DoChat("Hello, i am sanskar").then((response) => {
  //     console.log(response);
      
  //   });
  // }, []);

  const handleSendMessage = async () => {
    if (message.trim()) {
      const chatId = await CreateNewChat();
      console.log("New chat created with ID:", chatId);
      if (chatId) {
        // $currentChatId.set(chatId);
        navigate(`/chat/${chatId}`);
        $currentChatMessage.set(message);
        setTimeout(async() => {
          const res = await DoChat(message);
          $streamingMessage.set(res);
          console.log("Message sent:", res);
        }, 100);
        setMessage("");
      }
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
    >
      <div className="flex flex-1 items-center justify-center">
        <div className="w-full max-w-2xl mx-auto">
          <h2 className="text-2xl font-bold mb-6 text-center">
            How can I help you?
          </h2>
          <div className="bg-white p-6 rounded-lg border border-gray-200 shadow">
            <form>
              <ChatInput
                placeholder="Type your message here..."
                className="min-h-16 border-0 p-4 shadow-none focus-visible:ring-0 bg-gray-50 text-black text-lg font-semibold"
                value={message}
                onChange={(e) => setMessage(e.target.value)}
                onKeyDown={handleKeyDown}
              />
              <div className="flex items-center gap-2 pt-4">
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
            </form>
          </div>
        </div>
      </div>
    </div>
  );
}
