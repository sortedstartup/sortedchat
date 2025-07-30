import { useStore } from "@nanostores/react";
import React, { useState, useRef } from "react";
import { $currentProjectId } from "@/store/chat";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { File, Folder, RotateCcw } from "lucide-react";

export type FileItem = {
  id: string;
  file: File;
  path: string;
  status: "success" | "failed" | "uploading";
  error?: string;
};

type FileUploaderProps = {
  uploadUrl: string;
  onFileUpload?: (file: FileItem) => void;
  onCompleteUpload?: (allFiles: FileItem[]) => void;
};

export const FileUploader: React.FC<FileUploaderProps> = ({
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
    status: "success" | "failed" | "uploading",
    error?: string
  ) => {
    setFileList((prev) =>
      prev.map((f) => (f.id === id ? { ...f, status, error } : f))
    );
  };

  const uploadFile = async (fileItem: FileItem): Promise<FileItem> => {
    updateStatus(fileItem.id, "uploading");
    
    const formData = new FormData();
    formData.append("file", fileItem.file, fileItem.path);
    formData.append("project_id", currentProjectId.toString());

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
        const errorMsg = `Upload failed`;
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
        status: "uploading",
      });
    }

    const updatedList = [...fileList, ...newItems];
    setFileList(updatedList);

    Promise.all(newItems.map(uploadFile)).then((uploadedFiles) => {
      onCompleteUpload?.(uploadedFiles);
    });
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return;
    addFiles(e.target.files);
    e.target.value = "";
  };

  const getStatusBadge = (status: FileItem["status"]) => {
    if (status === "success") return <Badge variant="secondary">Success</Badge>;
    if (status === "failed") return <Badge variant="destructive">Failed</Badge>;
    if (status === "uploading") return <Badge variant="outline">Uploading</Badge>;
    return null;
  };

  return (
    <div className="space-y-4">
      <div className="flex gap-3">
        <Button
          onClick={() => fileInputRef.current?.click()}
          variant="outline"
          className="flex-1"
        >
          <File className="h-4 w-4 mr-2" />
          Select Files
        </Button>
        <Button
          onClick={() => folderInputRef.current?.click()}
          variant="outline"
          className="flex-1"
        >
          <Folder className="h-4 w-4 mr-2" />
          Select Folder
        </Button>
      </div>

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
        <Card>
          <CardContent className="p-4">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-sm font-medium">Upload Progress</h3>
            </div>

            <div className="space-y-2 max-h-48 overflow-y-auto">
              {fileList.map((f) => (
                <div
                  key={f.id}
                  className="flex items-center justify-between p-2 rounded-md border bg-gray-50/50"
                >
                  <div className="flex items-center gap-2 flex-1 min-w-0">
                    <span className="text-sm truncate" title={f.path}>
                      {f.path}
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    {getStatusBadge(f.status)}
                    {f.status === "failed" && (
                      <Button
                        onClick={() => uploadFile(f)}
                        size="sm"
                        variant="ghost"
                        className="h-6 px-2"
                        title="Retry upload"
                      >
                        <RotateCcw className="h-3 w-3" />
                      </Button>
                    )}
                  </div>
                </div>
              ))}
            </div>
          </CardContent>
        </Card>
      )}
    </div>
  );
};