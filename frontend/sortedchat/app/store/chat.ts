import { nanoquery } from "@nanostores/query";
import {
  ChatInfo,
  ChatMessage,
  ChatRequest,
  ChatResponse,
  CreateChatRequest,
  GetChatListRequest,
  GetHistoryRequest,
  SortedChatClient,
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
    console.log(chatId, ":", res);
    
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
      error: error as string || "Failed to fetch messages",
    });
  }
};

// Auto-fetch when chat ID changes
$currentChatId.listen((newChatId) => {
  console.log("Chat ID changed to:", newChatId);
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

// --- stores ----

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

  const stream = chat.Chat(
    ChatRequest.fromObject({
      text: msg,
      chatId: $currentChatId.get(),
    }),
    {}
  );

  stream.on("data", (res: ChatResponse) => {
    $streamingMessage.set($streamingMessage.get() + res.text);
  });

  stream.on("end", () => {});

  stream.on("error", (err) => {});
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
