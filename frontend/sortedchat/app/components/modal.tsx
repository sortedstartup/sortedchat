import React, { useState, useEffect } from "react";
import { useStore } from "@nanostores/react";
import { $searchText, $searchResults } from "../store/chat";
import { useNavigate } from "react-router";

interface SearchModalProps {
  isOpen: boolean;
  onClose: () => void;
}

const SearchModal: React.FC<SearchModalProps> = ({ isOpen, onClose }) => {
  const [localSearchText, setLocalSearchText] = useState("");
  const searchResults = useStore($searchResults);
   const navigate = useNavigate();

  // Debounced effect to update $searchText when user stops typing
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (localSearchText.trim()) {
        $searchText.set(localSearchText.trim());
      }
    }, 500); // 500ms delay after user stops typing

    return () => clearTimeout(timeoutId);
  }, [localSearchText]);

  // Reset local state when modal opens
  useEffect(() => {
    if (isOpen) {
      setLocalSearchText("");
    }
  }, [isOpen]);

  const handleClose = () => {
    onClose();
    $searchText.set("");
    $searchResults.set([]);
  };

  // Close modal on Escape key press
  useEffect(() => {
    const handleEscapeKey = (event: KeyboardEvent) => {
      if (event.key === "Escape") {
        handleClose();
      }
    };

    if (isOpen) {
      document.addEventListener("keydown", handleEscapeKey);
    }

    return () => {
      document.removeEventListener("keydown", handleEscapeKey);
    };
  }, [isOpen, onClose]);

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-lg w-3/4 max-w-2xl max-h-[80vh] overflow-hidden flex flex-col relative">
        {/* Close X button */}
        <button
          onClick={handleClose}
          className="absolute top-4 right-4 text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 text-xl font-bold transition-colors"
        >
          ×
        </button>

        <input
          type="text"
          placeholder="Search conversations..."
          className="w-full p-3 border-b border-gray-300 dark:border-gray-600 bg-transparent focus:outline-none focus:border-blue-500 text-gray-900 dark:text-white"
          value={localSearchText}
          onChange={(e) => setLocalSearchText(e.target.value)}
          autoFocus
        />

        {/* Search Results */}
        <div className="flex-1 overflow-y-auto mt-4 space-y-2">
          {searchResults.length > 0 ? (
            searchResults.map((result, index) => (
              <div
                key={index}
                className="p-3 border border-gray-200 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer"
                onClick={() => {
                  // Handle result selection here if needed
                  navigate(`/chat/${result.chat_id}`);
                  handleClose()
                }}
              >
                <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">
                  Chat: {result.chat_name || "Unnamed Chat"}
                </div>
                <div className="text-gray-900 dark:text-white line-clamp-2">
                  {result.matched_text}
                </div>
              </div>
            ))
          ) : localSearchText.trim() ? (
            <div className="text-center text-gray-500 dark:text-gray-400 py-8">
              No results found for "{localSearchText}"
            </div>
          ) : (
            <div className="text-center text-gray-500 dark:text-gray-400 py-8">
              Start typing to search your conversations...
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default SearchModal;
