import { useEffect, useState } from "react";
import {
  CornerDownLeft,
  FileText,
  Upload,
  Eye,
  MessageSquare,
  RefreshCw,
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
import { FileUploader } from "../components/FileUploader";
import { useStore } from "@nanostores/react";
import {
  $currentProject,
  $currentProjectId,
  $documents,
  fetchDocuments,
  createNewChat,
  $currentChatId,
  doChat,
  $projectChatList,
  SubmitGenerateEmbeddingsJob,
  $isErrorDocs,
  $isPolling,
} from "@/store/chat";
import { useNavigate, useParams } from "react-router-dom";
const API_UPLOAD_URL = import.meta.env.VITE_API_UPLOAD_URL;

export function Project() {
  const [message, setMessage] = useState("");
  const [isUploadDialogOpen, setIsUploadDialogOpen] = useState(false);
  const [isDocumentsDialogOpen, setIsDocumentsDialogOpen] = useState(false);
  const documents = useStore($documents);
  const projectName = useStore($currentProject);
  const currentProjectId = useStore($currentProjectId);
  const chatsList = useStore($projectChatList);
  const isErrorDocs = useStore($isErrorDocs);
    const isPolling = useStore($isPolling);

  const navigate = useNavigate();

  const { projectId } = useParams();

  useEffect(() => {
    if (projectId) {
      $currentProjectId.set(projectId);
      fetchDocuments(projectId);
    }
  }, [projectId]);

  const handleSendMessage = async () => {
    const newChatId = await createNewChat(projectId);
    $currentChatId.set(newChatId);
    navigate(`/project/${projectId}/chat/${newChatId}`);
    setTimeout(() => {
      doChat(message, projectId);
    }, 100);
    setMessage("");
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

  const handleRetryEmbedding = async () => {
    try {
      console.log("Retrying embedding for project:1", isErrorDocs);
      $isErrorDocs.set(false);

      await SubmitGenerateEmbeddingsJob(currentProjectId.toString());
    } catch (error) {
      console.error("Error retrying embedding:", error);
    }
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
        <div className="text-center max-w-md w-full">
          <div className="flex gap-3 justify-center mb-6">
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
              <DialogContent className="sm:max-w-2xl max-h-[70vh] overflow-hidden [&>button]:hidden">
                <DialogHeader className="flex flex-row items-center justify-between space-y-0 mr-2">
                  <DialogTitle>Project Documents</DialogTitle>
                  <div className="flex gap-2">
                      <Button
                        variant="outline"
                        size="sm"
                        onClick={() => {fetchDocuments(currentProjectId.toString())}}
                        disabled={isPolling}
                        className="gap-2"
                      >
                        <RefreshCw className={`h-4 w-4 ${isPolling ? 'animate-spin' : ''}`} />
                        {isPolling ? 'Refreshing...' : 'Refresh'}
                      </Button>
                      
                      {$isErrorDocs.get() ? (
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={handleRetryEmbedding}
                          className="gap-2"
                        >
                          <RefreshCw className="h-4 w-4" />
                          Regenerate
                        </Button>
                      ) : null}
                  </div>
                </DialogHeader>

                <div className="space-y-2 max-h-[50vh] overflow-auto">
                  {$documents.get().length > 0 ? (
                    $documents.get().map((doc: any, index: number) => (
                      <div
                        key={doc.id || index}
                        className="flex items-center justify-between p-3 border rounded-lg hover:bg-gray-50 transition-colors"
                      >
                        <div
                          className="flex items-center gap-3 flex-1 cursor-pointer"
                          onClick={() =>
                            window.open(
                              `${API_UPLOAD_URL}/documents/${doc.docs_id}`,
                              "_blank"
                            )
                          }
                        >
                          <FileText className="size-5 text-orange-500" />
                          <div className="flex flex-col items-start">
                            <span className="font-medium">{doc.file_name}</span>
                            <div className="flex items-center gap-2 mt-1">
                              <span className="text-xs text-gray-500">
                                {doc.embedding_status === 2
                                  ? "Indexing failed, Regenerate embeddings"
                                  : doc.embedding_status === 0
                                  ? "Currently in queue"
                                  : doc.embedding_status === 1
                                  ? "Embedding in progress"
                                  : ""}
                              </span>
                            </div>
                          </div>
                        </div>

                        <div className="flex items-center gap-2">
                          <Button variant="ghost" size="sm">
                            <Eye className="size-4" />
                          </Button>
                        </div>
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
                  uploadUrl={`${API_UPLOAD_URL}/upload`}
                  onFileUpload={(file) => console.log("Uploaded:", file)}
                  onCompleteUpload={handleUploadComplete}
                />
              </DialogContent>
            </Dialog>
          </div>

          <div className="w-full max-w-lg">
            <h3 className="text-lg font-semibold mb-3 text-gray-800">
              Project Chats
            </h3>
            <div className="space-y-2 max-h-64 overflow-auto">
              {chatsList.length > 0 ? (
                chatsList.map((chat: any) => (
                  <div
                    key={chat.chatId}
                    className="flex items-center gap-3 p-3 border rounded-lg hover:bg-gray-50 cursor-pointer transition-colors text-left"
                    onClick={() => {
                      navigate(`/project/${projectId}/chat/${chat.chatId}`);
                    }}
                  >
                    <MessageSquare className="size-5 text-orange-500 flex-shrink-0" />
                    <div className="flex-1 min-w-0">
                      <p className="font-medium text-gray-900 truncate">
                        {chat.name}
                      </p>
                    </div>
                  </div>
                ))
              ) : (
                <div className="text-center py-8 text-gray-500">
                  <MessageSquare className="size-12 mx-auto mb-3 text-gray-300" />
                  <p>No chats yet</p>
                </div>
              )}
            </div>
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
