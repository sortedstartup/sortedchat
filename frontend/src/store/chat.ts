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
  DocumentReference, // Already imported
} from "../../proto/chatservice";
import { atom, onMount } from "nanostores";

var chat = new SortedChatClient(import.meta.env.VITE_API_URL);

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

export const fetchChatMessages = async (chatId: string) => {
  if (!chatId) return;

  $currentChatMessages.set({
    data: undefined,
    loading: true,
    error: null,
  });

  try {
    const res = await chat.GetHistory(
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

// Auto-fetch when chat ID changes
$currentChatId.listen((newChatId) => {
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

export const $currentChatMessage = atom<string>("");
export const $streamingMessage = atom<string>("");

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
  
  const response = await chat.CreateChat(
    CreateChatRequest.fromObject(requestObj),
    {}
  );
  getChatList(projectId);
  return response.chat_id;
};

export const $projectChatList = atom<ChatInfo[]>([]);   
export const getChatList = (projectId?: string) => {
  const requestObj: GetChatListRequest = projectId
    ? GetChatListRequest.fromObject({ project_id: projectId })
    : new GetChatListRequest();

  chat.GetChatList(requestObj, {}).then((value: { chats: ChatInfo[] }) => {
    (projectId ? $projectChatList : $chatList).set(value.chats);
  });
};

const isFirstMessageInChat = (): boolean => {
  const currentState = $currentChatMessages.get();
  return !currentState.data || currentState.data.length === 0;
};

export const doChat = (msg: string,projectId: string | undefined) => {
  $currentChatMessage.set(msg);
  $streamingMessage.set("");
  // Don't reset document references here as they accumulate during the chat

  const isFirstMessage = isFirstMessageInChat();
  const isNewlyBranched = $isNewlyBranched.get();

  let assistantResponse = "";
  let messageId = "";
  let currentChatReferences: DocumentReference[] = []; // Track references for this specific chat

  if (isFirstMessage || isNewlyBranched) {
      generateChatName(msg);
      if (isNewlyBranched) {
        $isNewlyBranched.set(false);
      }
    }

  // grpc call
  const stream = chat.Chat(
    ChatRequest.fromObject({
      text: msg,
      chatId: $currentChatId.get(),
      model: $selectedModel.get(),
      project_id: projectId || "",
    }),
    {}
  );

  stream.on("data", (res: ChatResponse) => {
    if (res.has_text) {
      assistantResponse += res.text;
      $streamingMessage.set(assistantResponse);
    } else if (res.has_summary) {
      messageId = res.summary.message_id;
      console.log('Received message ID:', messageId);
    } else if (res.has_document_reference) {
      // Handle document reference
      const docRef = res.document_reference;
      console.log('Received document reference:', docRef.file_name, `Bytes ${docRef.start_byte}-${docRef.end_byte}`);
      
      // Add each chunk reference (don't avoid duplicates since we want all chunks)
      currentChatReferences.push(docRef);
      
      // Update the store for real-time display
      $currentDocumentReferences.set([...currentChatReferences]);
      $showDocumentReferences.set(true);
      
      // Only show toast for the first chunk of each document
      const existingDocsIds = currentChatReferences.slice(0, -1).map(ref => ref.docs_id);
      if (!existingDocsIds.includes(docRef.docs_id)) {
        toast.info(`Found relevant document: ${docRef.file_name}`);
      }
    }
  });

  stream.on("end", () => {
    const userMessage = ChatMessage.fromObject({
      role: "user",
      content: msg,
    });
    
    const assistantMessage = ChatMessage.fromObject({
      role: "assistant",
      content: assistantResponse,
      message_id: messageId,
      references: currentChatReferences, // Add references to the assistant message
    });

    addMessageToHistory(userMessage);
    addMessageToHistory(assistantMessage);

    $streamingMessage.set("");
    $currentChatMessage.set("");
    
    // Log all document references for this chat
    if (currentChatReferences.length > 0) {
      console.log('Document references for this chat:', currentChatReferences);
    }
  });

  stream.on("error", (err: Error) => {
    console.error("Stream error:", err);
    $streamingMessage.set("");
    $currentChatMessage.set("");
  });
};

// Add helper functions to control document references visibility
export const hideDocumentReferences = () => {
  $showDocumentReferences.set(false);
};

export const showDocumentReferencesPanel = () => {
  $showDocumentReferences.set(true);
};

// Clear document references when changing chats
$currentChatId.listen((_newValue, _oldValue) => {
  $streamingMessage.set("");
  $currentChatMessage.set("");
  // Don't clear document references here as they'll be set by fetchChatMessages
});

// ... rest of your existing code remains the same ...

export const $chatName = atom<string>("");
export const generateChatName = async (msg: string) => {
  try{
    // grpc call
    const response = await chat.GenerateChatName(
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
export const $selectedModel = atom<string>("gpt-4.1");

export const fetchAvailableModels = async () => {
  try {
    const response = await chat.ListModel(ListModelsRequest.fromObject({}), {});
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
    const response = await chat.SearchChat(
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
    const response = await chat.CreateProject(
      CreateProjectRequest.fromObject({
        name: name,
        description: description,
        additional_data: "", // TODO: looks like not needed
      }),
      {}
    );
    $currentProjectId.set(response.project_id);
    await getProjectList();
  } catch (error) {
    console.error("failed", error);
  }
};

export const getProjectList = async () => {
  try {
    const response = await chat.GetProjects(
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
    const res = await chat.ListDocuments(
      ListDocumentsRequest.fromObject({ project_id: projectId }),
      {}
    );

    $documents.set(res.documents);
  } catch (err) {
    console.error("Failed to fetch documents:", err);
    $documents.set([]);
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
    getChatList(newProjectId);
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
    const response = await chat.SubmitGenerateEmbeddingsJob(
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

    const res = await chat.BranchAChat(BranchAChatRequest.fromObject({
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
    const res = await chat.ListChatBranch(ListChatBranchRequest.fromObject({
      chat_id: chatId,
    }),{});
    console.log('response from branch chat list', res.branch_chat_list)
    $listChatBranch.set(res.branch_chat_list);
  } catch (error) {
    console.error('Failed to fetch branch chat list:', error);
    $listChatBranch.set([]);
  }
}

$currentChatId.listen((newChatId) => {
  if (newChatId) {
    ListChatBranch(newChatId);
  } else {
    $listChatBranch.set([]);
  }
});