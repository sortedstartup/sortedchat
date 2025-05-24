import { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import logoDark from "./logo-dark.svg";
import logoLight from "./logo-light.svg";
import { $chatList, $currentChatId, createNewChat, doChat } from "~/store/chat";
import { useStore } from "@nanostores/react";

export function Welcome() {
  const navigate = useNavigate();
  const [inputValue, setInputValue] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const chatList = useStore($chatList);

  const handleChatSelect = (chatId: string) => {
    $currentChatId.set(chatId);
    navigate(`/chat/${chatId}`);
  };

  const handleNewChat = async () => {
    try {
      setIsLoading(true);
      const chatId = await createNewChat();
      if (chatId) {
        navigate(`/chat/${chatId}`);
      } else {
        setError("Failed to create new chat");
      }
    } catch (err) {
      console.error("Failed to create new chat:", err);
      setError("Failed to create new chat");
    } finally {
      setIsLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!inputValue.trim()) return;
    
    try {
      setIsLoading(true);
      setError(null);
      
      const chatId = await createNewChat();
    
      if (chatId) {
        $currentChatId.set(chatId);    
        doChat(inputValue);
        navigate(`/chat/${chatId}`);
      } else {
        setError("Failed to create new chat");
      }
    } catch (err) {
      console.error("Failed to start chat:", err);
      setError("Failed to start chat");
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e as unknown as React.FormEvent);
    }
  };

  const handleExampleClick = (exampleText: string) => {
    setInputValue(exampleText);
  };

  return (
    <div className="flex h-screen bg-white dark:bg-gray-900">
      {/* Left sidebar - Chat list */}
      <div className="w-64 border-r border-gray-200 dark:border-gray-700 flex flex-col">
        <div className="p-4 border-b border-gray-200 dark:border-gray-700">
          <div className="w-full p-2">
            <img
              src={logoLight}
              alt="SortedChat"
              className="block w-32 mx-auto dark:hidden"
            />
            <img
              src={logoDark}
              alt="SortedChat"
              className="hidden w-32 mx-auto dark:block"
            />
          </div>
        </div>
        <div className="flex-1 overflow-y-auto">
          <div className="p-3">
            <button 
              onClick={handleNewChat}
              disabled={isLoading}
              className="w-full flex items-center justify-center gap-2 border border-gray-300 dark:border-gray-600 rounded-md p-3 text-sm hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              New Chat
            </button>
          </div>
          <ul className="mt-2">
            {chatList.map((chat) => (
                <li 
                  key={chat.chatId} 
                  onClick={() => handleChatSelect(chat.chatId)}
                  className="px-3 py-2 mx-2 rounded-md cursor-pointer hover:bg-gray-100 dark:hover:bg-gray-800"
                >
                  <div className="flex items-center">
                    <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
                    </svg>
                    <span className="text-sm truncate">{chat.name}</span>
                  </div>
                </li>
              ))
             }
          </ul>
        </div>
      </div>

      {/* Main welcome area with centered input */}
      <div className="flex-1 flex flex-col items-center justify-center">
        <div className="max-w-2xl w-full px-6">
          <div className="text-center mb-8">
            <h1 className="text-4xl font-bold mb-6">Welcome to SortedChat</h1>
            <p className="text-xl text-gray-600 dark:text-gray-300">
              How can I help you today?
            </p>
          </div>
          
          <form onSubmit={handleSubmit} className="w-full">
            <div className="flex items-end">
              <div className="flex-1 relative">
                <textarea
                  value={inputValue}
                  onChange={(e) => setInputValue(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder="Type a message to start a new chat..."
                  className="w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-800 rounded-lg px-4 py-3 pr-10 resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
                  rows={2}
                  autoFocus
                  disabled={isLoading}
                />
              </div>
              <button
                type="submit"
                disabled={!inputValue.trim() || isLoading}
                className={`ml-2 p-3 rounded-full ${
                  inputValue.trim() && !isLoading
                    ? 'bg-blue-500 text-white hover:bg-blue-600' 
                    : 'bg-gray-300 dark:bg-gray-700 text-gray-500 dark:text-gray-400 cursor-not-allowed'
                }`}
              >
                {isLoading ? (
                  <div className="h-5 w-5 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                ) : (
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" viewBox="0 0 20 20" fill="currentColor">
                    <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-8.707l-3-3a1 1 0 00-1.414 1.414L10.586 9H7a1 1 0 100 2h3.586l-1.293 1.293a1 1 0 101.414 1.414l3-3a1 1 0 000-1.414z" clipRule="evenodd" />
                  </svg>
                )}
              </button>
            </div>
            {error && <p className="mt-2 text-red-500 text-sm">{error}</p>}
          </form>
          
          <div className="mt-12 grid grid-cols-1 md:grid-cols-2 gap-4">
            <div className="bg-gray-50 dark:bg-gray-800 p-4 rounded-lg">
              <h3 className="font-semibold mb-2">Examples</h3>
              <ul className="space-y-2 text-sm">
                <li 
                  className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded cursor-pointer"
                  onClick={() => handleExampleClick("Explain quantum computing in simple terms")}
                >
                  "Explain quantum computing in simple terms"
                </li>
                <li 
                  className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded cursor-pointer"
                  onClick={() => handleExampleClick("How do I make a HTTP request in JavaScript?")}
                >
                  "How do I make a HTTP request in JavaScript?"
                </li>
                <li 
                  className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded cursor-pointer"
                  onClick={() => handleExampleClick("Write a poem about programming")}
                >
                  "Write a poem about programming"
                </li>
              </ul>
            </div>
            <div className="bg-gray-50 dark:bg-gray-800 p-4 rounded-lg">
              <h3 className="font-semibold mb-2">Capabilities</h3>
              <ul className="space-y-2 text-sm">
                <li className="p-2">Remembers what was said earlier in the conversation</li>
                <li className="p-2">Allows you to provide follow-up corrections</li>
                <li className="p-2">Can assist with creative tasks and problem-solving</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}