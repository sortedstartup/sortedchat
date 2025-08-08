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
  ChatNameRequest,
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

  const isFirstMessage = isFirstMessageInChat();

  let assistantResponse = "";

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
    assistantResponse += res.text;
    $streamingMessage.set(assistantResponse);
  });

  stream.on("end", () => {
    const userMessage = ChatMessage.fromObject({
      role: "user",
      content: msg,
    });
    const assistantMessage = ChatMessage.fromObject({
      role: "assistant",
      content: assistantResponse,
    });

    addMessageToHistory(userMessage);
    addMessageToHistory(assistantMessage);

    if (isFirstMessage) {
      chatName(msg);
    }

    $streamingMessage.set("");
    $currentChatMessage.set("");
  });

  stream.on("error", (err: Error) => {
    console.error("Stream error:", err);
    $streamingMessage.set("");
    $currentChatMessage.set("");
  });
};
export const $chatName = atom<string>("");
export const chatName = async (msg: string) => {
  try{
    // grpc call
    const response = await chat.GetChatName(
      ChatNameRequest.fromObject({
        message: msg,
        chat_id: $currentChatId.get()
      }),
      {}
    );
    
    $chatName.set(response.message)
  }
  catch(error) {
    console.error("Can't get the chat name", error)
  }
};

$chatName.listen(() => {
  getChatList();
});

$currentChatId.listen((_newValue, _oldValue) => {
  $streamingMessage.set("");
  $currentChatMessage.set("");
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
export const $searchText = atom<string>("elon");

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
// -- search --
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