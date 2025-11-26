import { toast } from "sonner";
import {
  ChatInfo,
  ChatMessage,
  ChatRequest,
  ChatResponse,
  ChatSearchRequest,
  CreateChatRequest,
  CreateProjectRequest,
  GetChatListRequest,
  GetHistoryRequest,
  ListModelsRequest,
  ModelListInfo,
  SearchResult,
  SortedChatClient,
  Project,
  GetProjectsRequest,
  ListDocumentsRequest,
  Document,
  GenerateEmbeddingRequest,
  GenerateChatNameRequest,
  BranchAChatRequest,
  ListChatBranchRequest,
  RAGDocumentReference as DocumentReference,
  ProjectContext, // Alias for backward compatibility
  RAGDocumentReferenceRequest,
  RAGDocumentReference,
  DeleteDocumentRequest,
  ResponseSummary,
  DeleteChatRequest,
  DeleteChatRequestOperation,
  RestoreChatRequest,
  RenameItemRequest,
  RenameItemRequestItemType,
  ChatProgress,
  MessageContent,
  ImageUrl,
} from "../../proto/chatservice";
import { atom, onMount } from "nanostores";
import { createAuthenticatedClientOptions } from "../lib/auth";
import { getUIConfig } from "../lib/config";
import type { ClientReadableStream } from "grpc-web";

let _chatClient: SortedChatClient | undefined;

export function getChatClient(): SortedChatClient {
  if (!_chatClient) {
    const config = getUIConfig();
    if (!config) {
      throw new Error("UI config not loaded, cannot initialize chat client.");
    }
    _chatClient = new SortedChatClient(
      config.API_URL,
      {},
      createAuthenticatedClientOptions()
    );
  }
  return _chatClient;
}

// --- stores ---
export const $chatList = atom<ChatInfo[]>([]);

export const $currentChatId = atom<string>("");

export const $currentChatMessages = atom<{
  data: ChatMessage[] | undefined;
  loading: boolean;
  error: string | null;
}>({
  data: undefined,
  loading: false,
  error: null,
});

// Add new stores for document references
export const $currentDocumentReferences = atom<DocumentReference[]>([]);
export const $showDocumentReferences = atom<boolean>(false);

// Add RAG enabled store
export const $ragEnabled = atom<boolean>(true);

// Store for detailed RAG document references
export const $ragDocumentDetails = atom<{
  data: RAGDocumentReference | null;
  loading: boolean;
  error: string | null;
}>({
  data: null,
  loading: false,
  error: null,
});

export const fetchChatMessages = async (chatId: string) => {
  if (!chatId) return;

  $currentChatMessages.set({
    data: undefined,
    loading: true,
    error: null,
  });

  try {
    const res = await getChatClient().GetHistory(
      GetHistoryRequest.fromObject({ chatId }),
      {}
    );

    $currentChatMessages.set({
      data: res.history || [],
      loading: false,
      error: null,
    });

    // Extract document references from chat history
    const allReferences: DocumentReference[] = [];
    if (res.history) {
      res.history.forEach(message => {
        if (message.references && message.references.length > 0) {
          allReferences.push(...message.references);
        }
      });
    }

    if (res.chat_metadata) {
      $chatMetadata.set(res.chat_metadata);
    } else {
      $chatMetadata.set(null);
    }

    
    // Set document references if any exist
    if (allReferences.length > 0) {
      $currentDocumentReferences.set(allReferences);
      $showDocumentReferences.set(true);
    } else {
      $currentDocumentReferences.set([]);
      $showDocumentReferences.set(false);
    }

  } catch (error) {
    console.error("Failed to fetch chat messages:", error);
    $currentChatMessages.set({
      data: undefined,
      loading: false,
      error: (error as string) || "Failed to fetch messages",
    });
  }
};

export const $currentChatMessage = atom<string>("");
export const $streamingMessage = atom<string>("");
//key here is message id and value is response summary, only used till we do not reload the page
//store new chat summaries after page load
export const $responseSummaries = atom<Record<string, ResponseSummary>>({});

export const $currentUserMessageId = atom<string>("");
export const $currentAssistantMessageId = atom<string | null>(null);

const addMessageToHistory = (message: ChatMessage) => {
  const currentState = $currentChatMessages.get();
  if (currentState.data) {
    // const messageCopy = structuredClone(message);  // will check this later
    $currentChatMessages.set({
      ...currentState,
      data: [...currentState.data, message],
    });
  }
};

// --- state management ---
export const createNewChat = async (projectId?: string) => {
  const requestObj: {name: string,project_id?: string} = {
    name: "",
  };
  if (projectId) {
    requestObj.project_id = projectId;
  }
  
  const response = await getChatClient().CreateChat(
    CreateChatRequest.fromObject(requestObj),
    {}
  );
  getChatList(projectId, false);
  return response.chat_id;
};

export const $projectChatList = atom<ChatInfo[]>([]);
export const $trashChatList = atom<ChatInfo[]>([]);

export const getChatList = (projectId?: string, softDeleted?: boolean) => {

  const requestObj = GetChatListRequest.fromObject({ project_id: projectId, soft_deleted: softDeleted });

  getChatClient().GetChatList(requestObj, {}).then((value: { chats: ChatInfo[] }) => {
    if (softDeleted) {
      $trashChatList.set(value.chats);  
    } else {
      (projectId ? $projectChatList : $chatList).set(value.chats);
    }
  });
};


const isFirstMessageInChat = (): boolean => {
  const currentState = $currentChatMessages.get();
  return !currentState.data || currentState.data.length === 0;
};

export const $chatMetadata = atom<ChatInfo | null>(null);
export const $chatProgress = atom<ChatProgress | null>(null);

export let stream: ClientReadableStream<ChatResponse> | null = null;
export let $isStreaming = atom<boolean>(false);

// Helper function to convert File to base64
async function imageToBase64(file: File): Promise<string> {
  const allowedTypes = ['image/jpeg', 'image/png', 'image/gif', 'image/webp'];
  if (!allowedTypes.includes(file.type)) {
    throw new Error(`Unsupported image type: ${file.type}. Allowed types: ${allowedTypes.join(', ')}`);
  }
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const base64 = reader.result as string;
      resolve(base64);  // Already in format: data:image/jpeg;base64,/9j/4AAQ...
    };
    reader.onerror = reject;
    reader.readAsDataURL(file);
  });
}

export const doChat = async (msg: string, projectId: string | undefined, images?: File[], imageDetail: string = "auto") => {
  $currentChatMessage.set(msg);
  $streamingMessage.set("");
  $currentUserMessageId.set("");
  $currentAssistantMessageId.set(null);

  const isFirstMessage = isFirstMessageInChat();
  const isNewlyBranched = $isNewlyBranched.get();

  let assistantResponse = "";
  let messageId = "";
  let currentChatReferences: any[] = []; // Track references for this specific chat

  if (isFirstMessage || isNewlyBranched) {
      generateChatName(msg);
      if (isNewlyBranched) {
        $isNewlyBranched.set(false);
      }
    }

  // Get RAG enabled state - use stored value for project chats, false for regular chats
  const ragEnabled = projectId ? $ragEnabled.get() : false;
  
  // Clear document references if RAG is disabled
  if (!ragEnabled) {
    $currentDocumentReferences.set([]);
    $showDocumentReferences.set(false);
  }

  // Build multi-modal content
  const contents: MessageContent[] = [];
  
  // Add text content if provided
  if (msg.trim()) {
    contents.push(
      MessageContent.fromObject({
        type: "text",
        text: msg
      })
    );
  }
  
  // Add image contents if provided
  if (images && images.length > 0) {
    // Check model capabilities before processing images

    const MAX_IMAGES = 10; // Align with backend limit
    if (images.length > MAX_IMAGES) {
      toast.error(`Maximum ${MAX_IMAGES} images allowed per message.`);
      return;
    }
    
    // Enforce total size limit (e.g., 20MB total)
    const MAX_TOTAL_SIZE = 20 * 1024 * 1024; // 20MB
    const totalSize = images.reduce((sum, img) => sum + img.size, 0);
    if (totalSize > MAX_TOTAL_SIZE) {
      toast.error(`Total image size must not exceed ${MAX_TOTAL_SIZE / (1024 * 1024)}MB.`);
      return;
    }


    const selectedModel = $selectedModel.get();
    const modelInfo = $availableModels.get().find(m => m.id === selectedModel);
    
    if (!modelInfo?.capabilities?.image?.input) {
      toast.error("Selected model does not support image input. Please choose a vision-capable model.");
      return;
    }
    
    for (const image of images) {
      try {
        const base64 = await imageToBase64(image);
        contents.push(
          MessageContent.fromObject({
            type: "image_url",
            image_url: ImageUrl.fromObject({
              url: base64,
              detail: imageDetail
            })
          })
        );
      } catch (error) {
        console.error("Failed to encode image:", error);
        toast.error("Failed to process image. Please try again.");
        return;
      }
    }
  }

  // grpc call
   stream = getChatClient().Chat(
    ChatRequest.fromObject({
      text: msg, // Keep for backward compatibility
      contents: contents, // New multi-modal content
      chatId: $currentChatId.get(),
      model: $selectedModel.get(),
      project_context: ProjectContext.fromObject({
        project_id: projectId || "",
        rag_enabled: ragEnabled,
      }),
    }),
    {}
  );

  stream.on("data", (res: ChatResponse) => {
    if (res.has_text) {
      assistantResponse += res.text;
      $streamingMessage.set(assistantResponse);
      $isStreaming.set(true);
    } else if (res.has_request_message_id) {
      $currentUserMessageId.set(res.request_message_id); //(user) message id is set in the store
    } else if (res.has_summary) {
      messageId = res.summary.message_id;
      const currentSummaries = $responseSummaries.get();
      $responseSummaries.set({
        ...currentSummaries,
        [res.summary.message_id]: res.summary,
      });
      $currentAssistantMessageId.set(messageId);


    } else if (res.has_document_reference && ragEnabled) {
      // Only process document references if RAG is enabled
      const docRefList = res.document_reference;
      
      if (docRefList.summary) {
        for (const summary of docRefList.summary) {
          
          const docRef = {
            doc_id: summary.doc_id,
            file_name: summary.file_name,
            Chunks: Array(summary.chunkCount).fill({}) // Create array with chunk count
          };
          
          currentChatReferences.push(docRef);
        }
      }
      
      // Update the store for real-time display
      $currentDocumentReferences.set([...currentChatReferences]);
      $showDocumentReferences.set(true);
      
    }else if (res.has_chat_metadata) {
      $chatMetadata.set(res.chat_metadata);
    } else if (res.has_progress) {
      $chatProgress.set(res.progress);
    }
  } );

  stream.on("end", () => {
    const userMessage = ChatMessage.fromObject({
      role: "user",
      content: msg,
      contents: contents.length > 0 ? contents : undefined, // Add multi-modal content
      rag_enabled: ragEnabled, // Set the rag_enabled field based on the current state
    });
    
    const assistantMessage = ChatMessage.fromObject({
      role: "assistant",
      content: assistantResponse,
      message_id: messageId,
      references: ragEnabled ? currentChatReferences : [], // Only add references if RAG is enabled
      rag_enabled: ragEnabled, // Set the rag_enabled field based on the current state
    });

    addMessageToHistory(userMessage);
    addMessageToHistory(assistantMessage);

    $isStreaming.set(false);

    $streamingMessage.set("");
    $currentChatMessage.set("");

    $chatProgress.set(null);

    
    // Clear document references if RAG is disabled
    if (!ragEnabled) {
      $currentDocumentReferences.set([]);
      $showDocumentReferences.set(false);
    }
    
    // Reset RAG to enabled for project chats after message completion
    if (projectId) {
      $ragEnabled.set(true);
    }
    
  });

  stream.on("error", (err: Error) => {
    console.error("Stream error:", err);
    $streamingMessage.set("");
    toast.error("An error occurred while receiving the response. Please try again.");  
    $chatProgress.set(null);    
    $isStreaming.set(false);
    
    // Reset RAG to enabled for project chats even on error
    if (projectId) {
      $ragEnabled.set(true);
    }
  });
};

// Add helper functions to control document references visibility
export const hideDocumentReferences = () => {
  $showDocumentReferences.set(false);
};

export const showDocumentReferencesPanel = () => {
  $showDocumentReferences.set(true);
};

// Add function to toggle RAG enabled state
export const toggleRagEnabled = () => {
  const currentState = $ragEnabled.get();
  $ragEnabled.set(!currentState);
  
  // Clear document references if RAG is being disabled
  if (currentState) {
    $currentDocumentReferences.set([]);
    $showDocumentReferences.set(false);
  }
};

// Add function to set RAG enabled state for project chats
export const setRagEnabledForProject = (enabled: boolean) => {
  $ragEnabled.set(enabled);
  
  // Clear document references if RAG is being disabled
  if (!enabled) {
    $currentDocumentReferences.set([]);
    $showDocumentReferences.set(false);
  }
};

export const $chatName = atom<string>("");
export const generateChatName = async (msg: string) => {
  try{
    // grpc call
    const response = await getChatClient().GenerateChatName(
      GenerateChatNameRequest.fromObject({
        message: msg,
        chat_id: $currentChatId.get(),
        model: $selectedModel.get()
      }),
      {}
    );
    
    $chatName.set(response.chat_name)
  }
  catch(error) {
    console.error("Can't get the chat name", error)
  }
};

$chatName.listen(() => {
  const currentProjectId = $currentProjectId.get();
  getChatList(); 
  if (currentProjectId) {
    getChatList(currentProjectId);
  }
});

// load chat history of first use
onMount($chatList, () => {
  getChatList();

  return () => {
    // Disabled mode
  };
});

export const $availableModels = atom<ModelListInfo[]>([]);
export const $selectedModel = atom<string>("gpt-5-nano");

export const fetchAvailableModels = async () => {
  try {
    const response = await getChatClient().ListModel(ListModelsRequest.fromObject({}), {});
    $availableModels.set(response.models);
  } catch (err) {
    console.error("Failed to fetch models:", err);
  }
};

onMount($availableModels, () => {
  fetchAvailableModels();
});

// -- search --
export const $searchResults = atom<SearchResult[]>([]);
export const $searchText = atom<string>("");

$searchText.listen((newValue, oldValue) => {
   if (newValue !== oldValue) {
    if (newValue === "") {
      $searchResults.set([]);
    } else {
      getSearchResults();
    }
  }
});

export const getSearchResults = async () => {
  try {
    const response = await getChatClient().SearchChat(
      ChatSearchRequest.fromObject({
        query: $searchText.get(),
      }),
      {}
    );
    $searchResults.set(response.results);
  } catch (err) {
    console.error("failed", err);
  }
};

// -- Project --
export const $currentProject = atom<string>("");
export const $projectList = atom<Project[]>([]);
export const $currentProjectId = atom<string>("");

export const createProject = async (
  name: string,
  description: string,
) => {
  try {
    const response = await getChatClient().CreateProject(
      CreateProjectRequest.fromObject({
        name: name,
        description: description,
        additional_data: "", // TODO: looks like not needed
      }),
      {}
    );
    $currentProjectId.set(response.project_id);
    toast.success(response.message);
    await getProjectList();
    return response.project_id;
  } catch (error) {
    console.error("failed", error);
    toast.error("Failed to create project: " + (error as Error).message);
    throw error;
  }
};

export const getProjectList = async () => {
  try {
    const response = await getChatClient().GetProjects(
      GetProjectsRequest.fromObject({}),
      {}
    );
    $projectList.set(response.projects || []);
    const currentId = $currentProjectId.get();
    if (currentId) {
      const foundProject = response.projects.find((p: Project) => p.id === currentId);
      if (foundProject) {
        $currentProject.set(foundProject.name);
      }
    }
  } catch (err) {
    console.error(err);
  }
};

onMount($projectList, () => {
  getProjectList();

  return () => {
    // Disabled mode
  };
});

export const $documents = atom<Document[]>([]);

export async function fetchDocuments(projectId: string) {
  try {
    const res = await getChatClient().ListDocuments(
      ListDocumentsRequest.fromObject({ project_id: projectId }),
      {}
    );

    $documents.set(res.documents);
  } catch (err) {
    console.error("Failed to fetch documents:", err);
    $documents.set([]);
  }
}

export async function deleteDocument(projectId: string, docId: string) {
  try {
    const res = await getChatClient().DeleteDocument(
      DeleteDocumentRequest.fromObject({
        project_id: projectId,
        doc_id: docId,
      }),
      {}
    );
    
    toast.success(res.message);
    
    // Refresh the documents list
    await fetchDocuments(projectId);
    
    return res.message;
  } catch (error) {
    console.error("Failed to delete document:", error);
    toast.error("Failed to delete document: " + (error as Error).message);
    throw error;
  }
}

$currentProjectId.listen((projectId) => {
  if (typeof projectId === "string" && projectId != "") {
    fetchDocuments(projectId);
  }
});

$documents.listen((projectId) => {
  if (typeof projectId === "string" && projectId !== "") {
    fetchDocuments(projectId);
  }
});

$currentProjectId.listen((newProjectId) => {
  if (newProjectId) {
    getChatList(newProjectId, false);
  } else {
    $chatList.set([]);
  }
});

export const $isErrorDocs = atom<boolean>(false);
export const $isPolling = atom<boolean>(false);

$documents.listen((documents) => {
  const hasErrorDocs = documents.some(doc => doc.embedding_status === 2);
  $isErrorDocs.set(hasErrorDocs);

  if ($isPolling.get()) {
    const allSuccessful = documents.every(doc => doc.embedding_status === 3);
    if (allSuccessful) {
      $isPolling.set(false);
    }
  }
});

export const SubmitGenerateEmbeddingsJob = async (projectId: string): Promise<String> => {
  try {
    const response = await getChatClient().SubmitGenerateEmbeddingsJob(
      GenerateEmbeddingRequest.fromObject({
        project_id: projectId,
      }),
      {}
    );
    
    $isPolling.set(true);
    toast.success(response.message || "Embedding job submitted successfully");
    
    for (let i = 0; i < 8; i++) {
      setTimeout(() => {
        if ($isPolling.get()) {
          fetchDocuments(projectId);
        }
        if (i === 7) {
          $isPolling.set(false);
        }
      }, i * 3000); 
    }
    
    return response.message; 
  } catch (error) {
    console.error("Failed to submit embedding job:", error);
    toast.error("Failed to submit embedding job: " + (error as Error).message);
    $isPolling.set(false);
    return "failed to submit embedding job";
  }
}

export const $isNewlyBranched = atom<boolean>(false);

export async function BranchChat(branch_from_message_id: string) {
  try {
    const currentChatId = $currentChatId.get();

    if (!branch_from_message_id || branch_from_message_id.trim() === "") {
      toast.error("Invalid message ID for branching");
      return;
    }

    const res = await getChatClient().BranchAChat(BranchAChatRequest.fromObject({
      source_chat_id: currentChatId,
      branch_from_message_id: branch_from_message_id,
      branch_name: ""
    }), {});

    if (res.new_chat_id) {
      toast.success("Chat branched successfully!");
      console.log('Setting isNewlyBranched to true for chat:', res.new_chat_id);
      $isNewlyBranched.set(true);
      $currentChatId.set(res.new_chat_id);

    } else {
      toast.error("Failed to create branch: No chat ID returned");
    }
  } catch (error) {
    console.error('Failed to branch chat:', error);
    toast.error(`Failed to branch chat: ${(error as Error).message || 'Unknown error'}`);
  }
}

export const $listChatBranch = atom<ChatInfo[]>([]);

export async function ListChatBranch (chatId: string) {
  try {
    const res = await getChatClient().ListChatBranch(ListChatBranchRequest.fromObject({
      chat_id: chatId,
    }),{});
    $listChatBranch.set(res.branch_chat_list);
  } catch (error) {
    console.error('Failed to fetch branch chat list:', error);
    $listChatBranch.set([]);
  }
}

$currentChatId.listen((newChatId) => {
  $streamingMessage.set("");
  $currentChatMessage.set("");
  $responseSummaries.set({});
  $currentUserMessageId.set("");
  $currentAssistantMessageId.set(null);
  $chatMetadata.set(null);

  // fetch branch chat list
  if (newChatId) {
    ListChatBranch(newChatId);
  } else {
    $listChatBranch.set([]);
  }

  // fetch chat messages
  if (newChatId) {
    fetchChatMessages(newChatId);
  } else {
    $currentChatMessages.set({
      data: undefined,
      loading: false,
      error: null,
    });
    // Clear document references when no chat is selected
    $currentDocumentReferences.set([]);
    $showDocumentReferences.set(false);
  }
});

// Function to fetch detailed RAG document references for a message
export const fetchRAGDocumentReference = async (messageId: string, projectId: string, docId?: string) => {
  if (!messageId) {
    console.error("Message ID is required to fetch RAG document references");
    return;
  }

  $ragDocumentDetails.set({
    data: null,
    loading: true,
    error: null,
  });

  try {
    const request = RAGDocumentReferenceRequest.fromObject({
      message_id: messageId,
      project_id: projectId,
      docId: docId || "", // Optional filter by specific document
    });

    const response = await getChatClient().GetRAGDocumentReference(request, {});
    
    $ragDocumentDetails.set({
      data: response.reference || null,
      loading: false,
      error: null,
    });

    return response.reference;
  } catch (error) {
    console.error('Failed to fetch RAG document reference:', error);
    const errorMessage = (error as Error).message || 'Failed to fetch document reference';
    
    $ragDocumentDetails.set({
      data: null,
      loading: false,
      error: errorMessage,
    });

    toast.error(`Failed to fetch document details: ${errorMessage}`);
    throw error;
  }
};

export const DeleteChat = async (chatId: string, operation: DeleteChatRequestOperation) => {
  try {
    const res = await getChatClient().DeleteChat(DeleteChatRequest.fromObject({ chat_id: chatId, operation: operation }), {});
    toast.success(res.message);


    if (operation === DeleteChatRequestOperation.SOFT_DELETE) {
      getChatList(undefined, false);
    }
    else {
      getChatList(undefined, true);
    }

    
  } catch (error) {
    console.error('Failed to Delete chat:', error);
    toast.error(`Failed to Delete chat: ${(error as Error).message || 'Unknown error'}`);
  }
}

export const RestoreChat = async (chatId: string) => {
  try {
    const res = await getChatClient().RestoreChat(RestoreChatRequest.fromObject({ chat_id: chatId }), {});
    toast.success(res.message);
    getChatList(undefined, true);
  } catch (error) {
    console.error('Failed to Restore chat:', error);
    toast.error(`Failed to Restore chat: ${(error as Error).message || 'Unknown error'}`);
  }
}

export const RenameItem = async (itemId: string, name: string, itemType: RenameItemRequestItemType) => {
  try {
    const res = await getChatClient().RenameItem(RenameItemRequest.fromObject({ item_id: itemId, name: name, item_type: itemType }), {});

    
    toast.success(res.message);

    if (itemType === RenameItemRequestItemType.CHAT) {
      const chatList = $chatList.get();
      chatList.forEach((chatItem: ChatInfo) => {
        if (chatItem.chatId === itemId) {
          chatItem.name = name;
        }
      });
      $chatList.set(chatList);
    } else if (itemType === RenameItemRequestItemType.PROJECT) {
      const projectList = $projectList.get();
      projectList.forEach((project: Project) => {
        if (project.id === itemId) {
          project.name = name;
        }
      });
      $projectList.set(projectList);
      
      // Update current project name if it's the one being renamed
      const currentProjectId = $currentProjectId.get();
      if (currentProjectId === itemId) {
        $currentProject.set(name);
      }
    }
    
  } catch (error) {
    console.error('Failed to rename item:', error);
    toast.error(`Failed to rename item: ${(error as Error).message || 'Unknown error'}`);
  }
}

// Keep the old RenameChat function for backward compatibility
export const RenameChat = async (chatId: string, name: string) => {
  return RenameItem(chatId, name, RenameItemRequestItemType.CHAT);
}

export const RenameProject = async (projectId: string, name: string) => {
  return RenameItem(projectId, name, RenameItemRequestItemType.PROJECT);
}