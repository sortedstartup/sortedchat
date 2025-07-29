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
export const createNewChat = async () => {
  const response = await chat.CreateChat(
    CreateChatRequest.fromObject({
      name: "New Chat",
    }),
    {}
  );
  getChatList();
  return response.chat_id;
};

export const doChat = (msg: string) => {
  $currentChatMessage.set(msg);
  $streamingMessage.set("");

  let assistantResponse = "";

  // grpc call
  const stream = chat.Chat(
    ChatRequest.fromObject({
      text: msg,
      chatId: $currentChatId.get(),
      model: $selectedModel.get(),
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

    $streamingMessage.set("");
    $currentChatMessage.set("");
  });

  stream.on("error", (err) => {
    console.error("Stream error:", err);
    $streamingMessage.set("");
    $currentChatMessage.set("");
  });
};

$currentChatId.listen((newValue, oldValue) => {
  $streamingMessage.set("");
  $currentChatMessage.set("");
});

const getChatList = () => {
  chat.GetChatList(GetChatListRequest.fromObject({}), {}).then((value) => {
    $chatList.set(value.chats);
  });
};

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
  if (newValue !== oldValue && newValue !== "") {
    getSearchResults();
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
export const $currentProjectId = atom<String>("");

export const createProject = async (
  description: string,
  additionalData: string
) => {
  try {
    const response = await chat.CreateProject(
      CreateProjectRequest.fromObject({
        name: $currentProject.get(),
        description: description,
        additional_data: additionalData,
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

    console.log(res.documents);

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
