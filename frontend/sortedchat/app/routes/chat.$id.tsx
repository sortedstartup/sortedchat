import { useState, useEffect, useRef } from "react";
import { useParams, useNavigate, Link } from "react-router";
import logoDark from "../welcome/logo-dark.svg";
import logoLight from "../welcome/logo-light.svg";
import {
  $chatList,
  $currentChatId,
  $currentChatMessage,
  $currentChatMessages,
  $streamingMessage,
  createNewChat,
  doChat,
} from "~/store/chat";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { useStore } from "@nanostores/react";

export default function Chat() {
  const { id } = useParams();
  const navigate = useNavigate();
  const chatId = id;

  const chatList = useStore($chatList);

  const { data, loading, error: fetchError } = useStore($currentChatMessages);

  const streamingMessage = useStore($streamingMessage);
  const currentChatMessage = useStore($currentChatMessage);

  const [inputValue, setInputValue] = useState("");
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const messagesEndRef = useRef<HTMLDivElement>(null);

  // Auto-scroll function
  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: "smooth" });
  };

  useEffect(() => {
    $currentChatId.set(chatId || "");
  }, [chatId]);

  useEffect(() => {
    scrollToBottom();
  }, [data, streamingMessage, currentChatMessage]);

  const handleSend = () => {
    console.log(inputValue);
    if (inputValue.trim()) {
      doChat(inputValue);
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

  const handleChatSelect = (selectedChatId: string) => {
    $currentChatId.set(selectedChatId);
    navigate(`/chat/${selectedChatId}`);
  };

  const handleNewChat = async () => {
    try {
      const chatId = await createNewChat();
      if (chatId) {
        navigate(`/chat/${chatId}`);
      } else {
        console.error("No chatId returned from server");
      }
    } catch (err) {
      console.error("Failed to create new chat:", err);
    }
  };

  if (error) {
    return (
      <div className="flex h-screen items-center justify-center bg-white dark:bg-gray-900">
        <div className="text-center p-6">
          <h1 className="text-2xl font-bold mb-4 text-red-500">Error</h1>
          <p className="mb-6">{error}</p>
          <Link
            to="/"
            className="bg-blue-500 text-white px-4 py-2 rounded-md hover:bg-blue-600 transition-colors"
          >
            Return to Home
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="flex h-screen bg-white dark:bg-gray-900">
      {/* Left sidebar - Chat list */}
      <div className="w-64 border-r border-gray-200 dark:border-gray-700 flex flex-col">
        <div className="p-4 border-b border-gray-200 dark:border-gray-700">
          <div className="w-full p-2">
            <Link to="/">
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
            </Link>
          </div>
        </div>
        <div className="flex-1 overflow-y-auto">
          <div className="p-3">
            <button
              onClick={handleNewChat}
              className="w-full flex items-center justify-center gap-2 border border-gray-300 dark:border-gray-600 rounded-md p-3 text-sm hover:bg-gray-100 dark:hover:bg-gray-800 transition-colors"
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-4 w-4"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 4v16m8-8H4"
                />
              </svg>
              New Chat
            </button>
          </div>
          <ul className="mt-2">
            {chatList.map((chat) => (
              <li
                key={chat.chatId}
                onClick={() => handleChatSelect(chat.chatId)}
                className={`px-3 py-2 mx-2 rounded-md cursor-pointer ${
                  chat.chatId === "" + chatId
                    ? "bg-gray-200 dark:bg-gray-700"
                    : "hover:bg-gray-100 dark:hover:bg-gray-800"
                }`}
              >
                <div className="flex items-center">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-4 w-4 mr-2"
                    fill="none"
                    viewBox="0 0 24 24"
                    stroke="currentColor"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"
                    />
                  </svg>
                  <span className="text-sm truncate">{chat.name}</span>
                </div>
              </li>
            ))}
          </ul>
        </div>
      </div>

      {/* Main chat area */}
      <div className="flex-1 flex flex-col">
        {/* Chat header */}
        <div className="border-b border-gray-200 dark:border-gray-700 p-4">
          <h1 className="text-lg font-semibold">
            {chatList.find((chat) => chat.chatId === chatId)?.name ||
              `Chat ${chatId}`}
          </h1>
        </div>

        {/* Chat messages */}
        <div className="flex-1 overflow-y-auto p-4">
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
              {/* Chat history */}
              {data?.map((message, index) => (
                <div
                  key={index}
                  className={`mb-4 flex ${
                    message.role === "user" ? "justify-end" : "justify-start"
                  }`}
                >
                  <div
                    className={`max-w-[80%] p-3 rounded-lg ${
                      message.role === "user"
                        ? "bg-blue-500 text-white rounded-br-none"
                        : "bg-gray-200 dark:bg-gray-700 rounded-bl-none"
                    }`}
                  >
                    <ReactMarkdown remarkPlugins={[remarkGfm]}>
                      {message.content}
                    </ReactMarkdown>
                  </div>
                </div>
              ))}

              {/* Current message user is typing (only show if there's content) */}
              {currentChatMessage && currentChatMessage.trim() && (
                <div className="mb-4 flex justify-end">
                  <div className="max-w-[80%] p-3 rounded-lg bg-blue-500 text-white rounded-br-none">
                    <ReactMarkdown remarkPlugins={[remarkGfm]}>
                      {currentChatMessage}
                    </ReactMarkdown>
                  </div>
                </div>
              )}

              {/* Streaming message (only show if there's content) */}
              {streamingMessage && streamingMessage.trim() && (
                <div className="mb-4 flex justify-start">
                  <div className="max-w-[80%] p-3 rounded-lg bg-gray-200 dark:bg-gray-700 rounded-bl-none">
                    <ReactMarkdown remarkPlugins={[remarkGfm]}>
                      {streamingMessage}
                    </ReactMarkdown>
                  </div>
                </div>
              )}
              <div ref={messagesEndRef} />
            </>
          )}
        </div>

        {/* Input area */}
        <div className="border-t border-gray-200 dark:border-gray-700 p-4">
          <div className="flex items-end">
            <div className="flex-1 relative">
              <textarea
                value={inputValue}
                onChange={(e) => setInputValue(e.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Type a message..."
                className="w-full border border-gray-300 dark:border-gray-600 dark:bg-gray-800 rounded-lg px-4 py-2 pr-10 resize-none focus:outline-none focus:ring-2 focus:ring-blue-500"
                rows={1}
              />
            </div>
            <button
              onClick={handleSend}
              disabled={!inputValue.trim() || isLoading}
              className={`ml-2 p-2 rounded-full ${
                inputValue.trim() && !isLoading
                  ? "bg-blue-500 text-white hover:bg-blue-600"
                  : "bg-gray-300 dark:bg-gray-700 text-gray-500 dark:text-gray-400 cursor-not-allowed"
              }`}
            >
              <svg
                xmlns="http://www.w3.org/2000/svg"
                className="h-5 w-5"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path
                  fillRule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-8.707l-3-3a1 1 0 00-1.414 1.414L10.586 9H7a1 1 0 100 2h3.586l-1.293 1.293a1 1 0 101.414 1.414l3-3a1 1 0 000-1.414z"
                  clipRule="evenodd"
                />
              </svg>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
