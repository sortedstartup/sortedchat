import React from "react";

interface SearchModalProps {
  isOpen: boolean;
  onClose: () => void;
}

const SearchModal: React.FC<SearchModalProps> = ({ isOpen, onClose }) => {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-transparent flex items-center justify-center z-50">
      <div className="bg-white dark:bg-gray-800 p-6 rounded-lg shadow-lg w-1/2 max-w-md">
        <input
          type="text"
          placeholder="Search..."
          className="w-full p-2 border-b border-gray-300  "
        />

        <div>
          <button
            className="bg-gray-200 hover:bg-gray-300 text-gray-800 px-4 py-2 rounded mr-2"
            onClick={onClose}
          >
            Cancel
          </button>
          
        </div>
      </div>
    </div>
  );
};

export default SearchModal;
