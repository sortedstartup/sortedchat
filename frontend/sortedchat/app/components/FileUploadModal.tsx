import React from "react";
import FileUploader from "./upload";

type UploadModalProps = {
  isOpen: boolean;
  onClose: () => void;
};

const UploadModal: React.FC<UploadModalProps> = ({ isOpen, onClose }) => {
  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
      <div className="bg-white dark:bg-gray-900 p-6 rounded-lg w-[90%] max-w-xl shadow-lg">
        <div className="flex justify-between items-center mb-4">
          <h2 className="text-lg font-semibold">Upload Files or Folder</h2>
          <button
            onClick={onClose}
            className="text-gray-600 hover:text-gray-900 dark:text-gray-300 dark:hover:text-white"
          >
            ×
          </button>
        </div>

        <FileUploader
          uploadUrl="http://localhost:8080/upload"
          onFileUpload={(file) => console.log("Uploaded:", file)}
          onCompleteUpload={(files) => console.log("All files uploaded:", files)}
        />
      </div>
    </div>
  );
};

export default UploadModal;
