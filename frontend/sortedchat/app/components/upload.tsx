import { useStore } from "@nanostores/react";
import React, { useState, useRef } from "react";
import { $currentProjectId } from "~/store/chat";

export type FileItem = {
  id: string;
  file: File;
  path: string;
  status: "success" | "failed";
  error?: string;
};

type FileUploaderProps = {
  uploadUrl: string;
  onFileUpload?: (file: FileItem) => void;
  onCompleteUpload?: (allFiles: FileItem[]) => void;
};

const FileUploader: React.FC<FileUploaderProps> = ({
  uploadUrl,
  onFileUpload,
  onCompleteUpload,
}) => {
  const [fileList, setFileList] = useState<FileItem[]>([]);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const folderInputRef = useRef<HTMLInputElement | null>(null);
  const currentProjectId = useStore($currentProjectId);

  const updateStatus = (
    id: string,
    status: "success" | "failed",
    error?: string
  ) => {
    setFileList((prev) =>
      prev.map((f) => (f.id === id ? { ...f, status, error } : f))
    );
  };

  const uploadFile = async (fileItem: FileItem): Promise<FileItem> => {
    const formData = new FormData();
    formData.append("file", fileItem.file, fileItem.path);
    formData.append("project_id",currentProjectId.toString())

    const res = await fetch(uploadUrl, {
      method: "POST",
      body: formData,
    });

    if (res.ok) {
      const updated: FileItem = { ...fileItem, status: "success" };
      updateStatus(fileItem.id, "success");
      onFileUpload?.(updated);
      return updated;
    } else {
      const errorMsg = `error`;
      const updated: FileItem = {
        ...fileItem,
        status: "failed",
        error: errorMsg,
      };
      updateStatus(fileItem.id, "failed", errorMsg);
      onFileUpload?.(updated);
      return updated;
    }
  };

  const addFiles = (files: FileList) => {
    const newItems: FileItem[] = [];
    for (let i = 0; i < files.length; i++) {
      const f = files[i];
      newItems.push({
        id: crypto.randomUUID(),
        file: f,
        path: (f as any).webkitRelativePath || f.name,
        status: "failed",
      });
    }

    const updatedList = [...fileList, ...newItems];
    setFileList(updatedList);

    Promise.all(newItems.map(uploadFile)).then(() => {
      onCompleteUpload?.([...updatedList]);
    });
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return;
    addFiles(e.target.files);
    e.target.value = "";
  };

  return (
    <div>
      <button
        onClick={() => fileInputRef.current?.click()}
        className="border-2 p-4 m-2"
      >
        Select Files
      </button>
      <button
        onClick={() => folderInputRef.current?.click()}
        className="border-2 p-4 m-2"
      >
        Select Folder
      </button>

      <input
        type="file"
        multiple
        ref={fileInputRef}
        style={{ display: "none" }}
        onChange={handleFileChange}
      />

      <input
        type="file"
        //@ts-ignore
        webkitdirectory=""
        ref={folderInputRef}
        style={{ display: "none" }}
        onChange={handleFileChange}
      />

      {fileList.length > 0 && (
        <ul>
          {fileList.map((f) => (
            <li key={f.id}>
              {f.path} - {f.status}
              {f.status === "failed" && (
                <>
                  <span> ({f.error}) </span>
                  <button onClick={() => uploadFile(f)}>Retry</button>
                </>
              )}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
};

export default FileUploader;
