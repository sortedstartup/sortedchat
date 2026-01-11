import { useState, useEffect } from "react";
import { Button } from "@/components/ui/button";
import { ChatInput } from "@/components/ui/chat/chat-input";
import { CornerDownLeft } from "lucide-react";
import { createNewChat, $currentChatId, doChat } from "@/store/chat";
import { useNavigate, useSearchParams } from "react-router-dom";
import confetti from 'canvas-confetti';

export function Home() {
  const [message, setMessage] = useState("");
  const navigate = useNavigate();
  const [searchParams, setSearchParams] = useSearchParams();

  useEffect(() => {
    if (searchParams.get('welcome') === 'true') {
      const count = 200;
      const defaults = {
        origin: { y: 0.7 },
        zIndex: 10000
      };

      const fire = (particleRatio: number, opts: confetti.Options) => {
        confetti({
          ...defaults,
          ...opts,
          particleCount: Math.floor(count * particleRatio)
        });
      };

      fire(0.25, {
        spread: 26,
        startVelocity: 55,
      });
      fire(0.2, {
        spread: 60,
      });
      fire(0.35, {
        spread: 100,
        decay: 0.91,
        scalar: 0.8
      });
      fire(0.1, {
        spread: 120,
        startVelocity: 25,
        decay: 0.92,
        scalar: 1.2
      });
      fire(0.1, {
        spread: 120,
        startVelocity: 45,
      });

      // Clean up the URL
      setSearchParams({}, { replace: true });
    }
  }, [searchParams, setSearchParams]);

  const handleSendMessage = async () => {
    if (message.trim()) {
      const chatId = await createNewChat();
      if (chatId) {
        $currentChatId.set(chatId);
        navigate(`/chat/${chatId}`);
        setTimeout(() => {
          doChat(message, undefined);
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
          <div className="bg-card p-6 rounded-lg border border-border shadow">
            <form>
              <ChatInput
                placeholder="Type your message here..."
                className="min-h-16 border-0 p-4 shadow-none focus-visible:ring-0 bg-muted text-foreground text-lg font-semibold"
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
                  Send Messages
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
