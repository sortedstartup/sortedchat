import { useEffect, useState } from "react";
import {
  Paperclip,
  Mic,
  CornerDownLeft,
  FileText,
  Upload,
  Eye,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { ChatInput } from "@/components/ui/chat/chat-input";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { FileUploader } from "./FileUploader";
import { useStore } from "@nanostores/react";
import {
  $currentProject,
  $currentProjectId,
  $documents,
  fetchDocuments,
} from "@/store/chat";
import { useParams } from "react-router-dom";

export function Project() {
  const [message, setMessage] = useState("");
  const [isUploadDialogOpen, setIsUploadDialogOpen] = useState(false);
  const [isDocumentsDialogOpen, setIsDocumentsDialogOpen] = useState(false);
  const documents = useStore($documents);
  const projectName = useStore($currentProject);
  const currentProjectId = useStore($currentProjectId);
  const { projectId } = useParams();

  useEffect(() => {
    if (projectId) {
      $currentProjectId.set(projectId);
      fetchDocuments(projectId);
    }
  }, [projectId]);

  const handleSendMessage = () => {
    if (message.trim()) {
      console.log("Sending message:", message);
      setMessage("");
    }
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  const handleUploadComplete = (files: any[]) => {
    console.log("All files uploaded:", files);
  };

  const handleUploadDialogClose = (open: boolean) => {
    setIsUploadDialogOpen(open);

    if (!open) {
      fetchDocuments(currentProjectId.toString());
    }
  };

  const handleDocumentsDialogClose = (open: boolean) => {
    setIsDocumentsDialogOpen(open);
  };

  return (
    <div className="flex flex-col h-full mx-4 max-h-full">
      <div className="p-4 border-b border-gray-200 flex-shrink-0">
        <div className="flex items-center gap-3">
          <div className="w-8 h-8 bg-gray-100 rounded flex items-center justify-center">
            <FileText className="size-5 text-orange-500" />
          </div>
          <h1 className="text-xl font-bold">{projectName}</h1>
        </div>
      </div>

      <div className="flex-1 flex items-center justify-center p-6 overflow-hidden">
        <div className="text-center max-w-md">
          <div className="flex gap-3 justify-center">
            <Dialog
              open={isDocumentsDialogOpen}
              onOpenChange={handleDocumentsDialogClose}
            >
              <DialogTrigger asChild>
                <Button variant="outline" size="sm">
                  <FileText className="size-4 mr-2" />
                  Project Documents
                  {documents.length > 0 && (
                    <span className="ml-2 bg-orange-100 text-orange-800 text-xs px-2 py-1 rounded-full">
                      {documents.length}
                    </span>
                  )}
                </Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-2xl max-h-[70vh] overflow-hidden">
                <DialogHeader>
                  <DialogTitle>Project Documents</DialogTitle>
                </DialogHeader>

                <div className="space-y-2 max-h-[50vh] overflow-auto">
                  {documents.length > 0 ? (
                    documents.map((doc: any, index: number) => (
                      <div
                        key={doc.id || index}
                        className="flex items-center justify-between p-3 border rounded-lg hover:bg-gray-50 cursor-pointer transition-colors"
                        onClick={() =>
                          window.open(
                            `http://localhost:8080/documents/${doc.docs_id}`,
                            "_blank",
                            "noopener,noreferrer"
                          )
                        }
                      >
                        <div className="flex items-center gap-3">
                          <FileText className="size-5 text-orange-500" />
                          <div className="flex flex-col items-start">
                            <span className="font-medium">{doc.file_name}</span>
                          </div>
                        </div>
                        <Button variant="ghost" size="sm">
                          <Eye className="size-4" />
                        </Button>
                      </div>
                    ))
                  ) : (
                    <div className="text-center py-8 text-gray-500">
                      <FileText className="size-12 mx-auto mb-3 text-gray-300" />
                      <p>No documents uploaded yet</p>
                      <p className="text-sm">Upload files to see them here</p>
                    </div>
                  )}
                </div>
              </DialogContent>
            </Dialog>

            <Dialog
              open={isUploadDialogOpen}
              onOpenChange={handleUploadDialogClose}
            >
              <DialogTrigger asChild>
                <Button variant="outline" size="sm">
                  <Upload className="size-4 mr-2" />
                  Upload Documents
                </Button>
              </DialogTrigger>
              <DialogContent className="sm:max-w-md">
                <DialogHeader>
                  <DialogTitle>Upload Files or Folder</DialogTitle>
                </DialogHeader>
                <FileUploader
                  uploadUrl="http://localhost:8080/upload"
                  onFileUpload={(file) => console.log("Uploaded:", file)}
                  onCompleteUpload={handleUploadComplete}
                />
              </DialogContent>
            </Dialog>
          </div>
        </div>
      </div>

      <div className="bg-white p-4 border-t border-gray-200 flex-shrink-0">
        <div className="relative rounded-lg border border-gray-200 bg-gray-50 focus-within:ring-1 focus-within:ring-orange-500 p-1">
          <ChatInput
            placeholder="Type your message here..."
            className="min-h-12 border-0 p-3 shadow-none focus-visible:ring-0 bg-gray-50 text-black resize-none"
            value={message}
            onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
              setMessage(e.target.value)
            }
            onKeyDown={handleKeyDown}
          />
          <div className="flex items-center p-3 pt-0">
            <Button
              size="sm"
              className="ml-auto gap-1.5"
              onClick={handleSendMessage}
              type="button"
            >
              Send Message
              <CornerDownLeft className="size-4" />
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
