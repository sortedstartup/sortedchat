import { useState, useEffect } from "react";
import { useNavigate } from "react-router";
import logoDark from "./logo-dark.svg";
import logoLight from "./logo-light.svg";
import { chatStore } from "../utils/chatStore";
import { doChat } from "~/store/chat";

export function Welcome() {
  const navigate = useNavigate();
  const [inputValue, setInputValue] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  // const [chats, setChats] = useState(chatStore.getAllChats().map(chat => ({
  //   ...chat,
  //   selected: false,
  // })));

  const [chats, setChats] = useState<Array<{ id: number; name: string; selected: boolean }>>([]);


  // Refresh chats from store when component mounts
  useEffect(() => {
    chatStore.getAllChats().then(fetchedChats => {
        setChats(
          fetchedChats.map(chat => ({
            ...chat,
            selected: Number(chat.id) === 1,
          }))
        );
      }).catch(err => {
        console.error("Failed to load chats:", err);
      });
  }, []);

  const handleChatSelect = (chatId: number) => {
    navigate(`/chat/${chatId}`);
  };

  const handleNewChat = () => {
    const newChatId = Math.max(...chats.map(chat => chat.id), 0) + 1;
    navigate(`/chat/${newChatId}`);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (inputValue.trim()) {
      try {
        setIsLoading(true);
        setError(null);
        
        // Create a new chat with this initial message
        const newChatId = Math.max(...chats.map(chat => chat.id), 0) + 1;
        
        // Get a short name for the chat
        const chatName = inputValue.length > 20 ? inputValue.substring(0, 20) + "..." : inputValue;
        
        // Save the user's input
        const userMessage = inputValue;
        
        // Initialize the chat with just the user's message for now
        chatStore.createChat(newChatId, chatName, userMessage);
        
        // Call the server for a response
        try {
          const serverResponse = await doChat(userMessage);
          
          // Add the server's response to the chat
          chatStore.addMessage(newChatId, {
            id: 2,
            sender: "ai",
            content: serverResponse || "Sorry, I couldn't process your request."
          });
        } catch (err) {
          console.error("Error getting response from server:", err);
          
          // Add an error message to the chat
          chatStore.addMessage(newChatId, {
            id: 2,
            sender: "ai",
            content: "Sorry, I encountered an error processing your request."
          });
        }
        
        // Navigate to the new chat
        navigate(`/chat/${newChatId}`);
      } catch (err) {
        console.error("Error creating chat:", err);
        setError("Failed to create chat. Please try again.");
      } finally {
        setIsLoading(false);
      }
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSubmit(e as unknown as React.FormEvent);
    }
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
              className="w-full flex items-center justify-center gap-2 border border-gray-300 dark:border-gray-600 rounded-md p-3 text-sm hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
            >
              <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              New Chat
            </button>
          </div>
          <ul className="mt-2">
            {chats.map((chat) => (
              <li 
                key={chat.id} 
                onClick={() => handleChatSelect(chat.id)}
                className={`px-3 py-2 mx-2 rounded-md cursor-pointer ${chat.selected ? "bg-gray-200 dark:bg-gray-700" : "hover:bg-gray-100 dark:hover:bg-gray-800"}`}
              >
                <div className="flex items-center">
                  <svg xmlns="http://www.w3.org/2000/svg" className="h-4 w-4 mr-2" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z" />
                  </svg>
                  <span className="text-sm truncate">{chat.name}</span>
                </div>
              </li>
            ))}
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
                <li className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded cursor-pointer">
                  "Explain quantum computing in simple terms"
                </li>
                <li className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded cursor-pointer">
                  "How do I make a HTTP request in JavaScript?"
                </li>
                <li className="p-2 hover:bg-gray-100 dark:hover:bg-gray-700 rounded cursor-pointer">
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

const resources = [
  {
    href: "https://reactrouter.com/docs",
    text: "React Router Docs",
    icon: (
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="24"
        height="20"
        viewBox="0 0 20 20"
        fill="none"
        className="stroke-gray-600 group-hover:stroke-current dark:stroke-gray-300"
      >
        <path
          d="M9.99981 10.0751V9.99992M17.4688 17.4688C15.889 19.0485 11.2645 16.9853 7.13958 12.8604C3.01467 8.73546 0.951405 4.11091 2.53116 2.53116C4.11091 0.951405 8.73546 3.01467 12.8604 7.13958C16.9853 11.2645 19.0485 15.889 17.4688 17.4688ZM2.53132 17.4688C0.951566 15.8891 3.01483 11.2645 7.13974 7.13963C11.2647 3.01471 15.8892 0.951453 17.469 2.53121C19.0487 4.11096 16.9854 8.73551 12.8605 12.8604C8.73562 16.9853 4.11107 19.0486 2.53132 17.4688Z"
          strokeWidth="1.5"
          strokeLinecap="round"
        />
      </svg>
    ),
  },
  {
    href: "https://rmx.as/discord",
    text: "Join Discord",
    icon: (
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="24"
        height="20"
        viewBox="0 0 24 20"
        fill="none"
        className="stroke-gray-600 group-hover:stroke-current dark:stroke-gray-300"
      >
        <path
          d="M15.0686 1.25995L14.5477 1.17423L14.2913 1.63578C14.1754 1.84439 14.0545 2.08275 13.9422 2.31963C12.6461 2.16488 11.3406 2.16505 10.0445 2.32014C9.92822 2.08178 9.80478 1.84975 9.67412 1.62413L9.41449 1.17584L8.90333 1.25995C7.33547 1.51794 5.80717 1.99419 4.37748 2.66939L4.19 2.75793L4.07461 2.93019C1.23864 7.16437 0.46302 11.3053 0.838165 15.3924L0.868838 15.7266L1.13844 15.9264C2.81818 17.1714 4.68053 18.1233 6.68582 18.719L7.18892 18.8684L7.50166 18.4469C7.96179 17.8268 8.36504 17.1824 8.709 16.4944L8.71099 16.4904C10.8645 17.0471 13.128 17.0485 15.2821 16.4947C15.6261 17.1826 16.0293 17.8269 16.4892 18.4469L16.805 18.8725L17.3116 18.717C19.3056 18.105 21.1876 17.1751 22.8559 15.9238L23.1224 15.724L23.1528 15.3923C23.5873 10.6524 22.3579 6.53306 19.8947 2.90714L19.7759 2.73227L19.5833 2.64518C18.1437 1.99439 16.6386 1.51826 15.0686 1.25995ZM16.6074 10.7755L16.6074 10.7756C16.5934 11.6409 16.0212 12.1444 15.4783 12.1444C14.9297 12.1444 14.3493 11.6173 14.3493 10.7877C14.3493 9.94885 14.9378 9.41192 15.4783 9.41192C16.0471 9.41192 16.6209 9.93851 16.6074 10.7755ZM8.49373 12.1444C7.94513 12.1444 7.36471 11.6173 7.36471 10.7877C7.36471 9.94885 7.95323 9.41192 8.49373 9.41192C9.06038 9.41192 9.63892 9.93712 9.6417 10.7815C9.62517 11.6239 9.05462 12.1444 8.49373 12.1444Z"
          strokeWidth="1.5"
        />
      </svg>
    ),
  },
];
