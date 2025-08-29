import { useStore } from "@nanostores/react";
import React, { useState, useRef } from "react";
import { $currentProjectId } from "@/store/chat";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { File, Folder, RotateCcw } from "lucide-react";
import { getJWTToken } from "@/lib/auth";
import * as pdfjsLib from "pdfjs-dist";
import PdfWorker from "pdfjs-dist/build/pdf.worker?worker";

pdfjsLib.GlobalWorkerOptions.workerPort = new PdfWorker();

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

    // Prepare headers with JWT token
    const headers: Record<string, string> = {};
    const token = getJWTToken();
    if (token) {
      headers["Authorization"] = `Bearer ${token}`;
      console.debug('Added JWT token to file upload request');
    } else {
      console.debug('No JWT token available for file upload request');
    }

      const res = await fetch(uploadUrl, {
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
      onCompleteUpload?.(uploadedFiles);
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