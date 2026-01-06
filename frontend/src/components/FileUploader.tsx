import { useStore } from "@nanostores/react";
import React, { useState, useRef, useEffect } from "react";
import { $currentProjectId } from "@/store/chat";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { File, Folder, RotateCcw } from "lucide-react";
import { getJWTToken } from "@/lib/auth";
import { getUIConfig } from "@/lib/config";
import * as pdfjsLib from "pdfjs-dist";
import PdfWorker from "pdfjs-dist/build/pdf.worker?worker";

if (typeof window !== "undefined" && !pdfjsLib.GlobalWorkerOptions.workerPort) {
  pdfjsLib.GlobalWorkerOptions.workerPort = new PdfWorker();
}

export type FileItem = {
  id: string;
  file: File;
  path: string;
  status: "success" | "failed" | "uploading";
  error?: string;
};

type FileUploaderProps = {
  uploadUrl: string;
  projectId?: string;  // For project uploads
  agentId?: string;    // For agent uploads
  onFileUpload?: (file: FileItem) => void;
  onCompleteUpload?: (allFiles: FileItem[]) => void;
};

export const FileUploader: React.FC<FileUploaderProps> = ({
  uploadUrl,
  projectId,
  agentId,
  onFileUpload,
  onCompleteUpload,
}) => {
  const [fileList, setFileList] = useState<FileItem[]>([]);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const folderInputRef = useRef<HTMLInputElement | null>(null);
  const currentProjectId = useStore($currentProjectId);
  const [fullUploadUrl, setFullUploadUrl] = useState<string>("");

  // Get full API URL from config
  useEffect(() => {
    const config = getUIConfig();
    if (config) {
      setFullUploadUrl(config.API_UPLOAD_URL + uploadUrl);
    }
  }, [uploadUrl]);

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
    
    // Add either project_id or agent_id based on what's provided
    if (agentId) {
      formData.append("agent_id", agentId);
      formData.append("file_path", fileItem.path); // Send full path for agent uploads
    } else {
      const targetProjectId = projectId || currentProjectId.toString();
      formData.append("project_id", targetProjectId);
    }

    // Prepare headers with JWT token
    const headers: Record<string, string> = {};
    const token = getJWTToken();
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
      console.debug('Added JWT token to file upload request');
    } else {
      console.debug('No JWT token available for file upload request');
    }

    try {
      const res = await fetch(fullUploadUrl, {
        method: "POST",
        headers,
        body: formData,
      });

      if (res.ok) {
        const updated: FileItem = { ...fileItem, status: "success" };
        updateStatus(fileItem.id, "success");
        onFileUpload?.(updated);
        return updated;
      } else {
        const errorText = await res.text();
        const errorMsg = `Upload failed: ${res.status} ${errorText}`;
        console.error("Upload failed:", errorMsg);
        const updated: FileItem = {
          ...fileItem,
          status: "failed",
          error: errorMsg,
        };
        updateStatus(fileItem.id, "failed", errorMsg);
        onFileUpload?.(updated);
        return updated;
      }
    } catch (error) {
      const errorMsg = `Upload failed: ${error}`;
      console.error("Upload error:", error);
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

  const processFileForUpload = async (file: File, path: string): Promise<FileItem> => {
    const baseFileItem: FileItem = {
      id: crypto.randomUUID(),
      file: file,
      path: path,
      status: "uploading",
    };

    if (file.type === "text/plain") {
      return baseFileItem;
    }

    if (file.type === "application/pdf" || /\.pdf$/i.test(file.name)) {
      try {
        const arrayBuffer = await file.arrayBuffer();
        const loadingTask = pdfjsLib.getDocument({ data: arrayBuffer });
        const pdf = await loadingTask.promise;

        const chunks: string[] = [];
        for (let i = 1; i <= pdf.numPages; i++) {
          const page = await pdf.getPage(i);
          const content = await page.getTextContent();
          const pageText = (content.items as any[])
            .map((item) => (typeof item === "object" && "str" in item ? (item as any).str : ""))
            .join(" ");
          chunks.push(pageText, "\n\n");
        }

        const textBlob = new Blob([chunks.join("")], { type: "text/plain" });
        const fileName = file.name.replace(/\.pdf$/i, ".txt");
        
        const textFile = Object.assign(textBlob, {
          name: fileName,
        }) as File;

        return {
          ...baseFileItem,
          file: textFile,
          path: path.replace(/\.pdf$/i, ".txt"),
        };
      } catch (error) {
        console.error("PDF extraction failed:", error);
        return {
          ...baseFileItem,
          status: "failed",
          error: "PDF text extraction failed",
        };
      }
    }
    return baseFileItem;
  };

  const addFiles = async (files: FileList) => {
    const newItems = await Promise.all(
      Array.from(files).map(f => {
        const path = (f as any).webkitRelativePath || f.name;
        return processFileForUpload(f, path);
      })
    );

    const updatedList = [...fileList, ...newItems];
    setFileList(updatedList);

    Promise.all(newItems.map(uploadFile)).then((uploadedFiles) => {
      // Only pass successfully uploaded files to callback
      const successfulFiles = uploadedFiles.filter(f => f.status === "success");
      onCompleteUpload?.(successfulFiles);
    });
  };

  const handleFileChange = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files) return;
    await addFiles(e.target.files);
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