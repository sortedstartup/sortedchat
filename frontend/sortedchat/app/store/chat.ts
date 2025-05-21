import { ChatInfo, ChatRequest, ChatResponse, GetChatListRequest, GetHistoryRequest, SortedChatClient } from "../../proto/chatservice"
import { atom, onMount } from 'nanostores'

var chat = new SortedChatClient(import.meta.env.VITE_API_URL)

function doChat(
  message: string,
  chatId: string,
  onMessage: (chunk: string) => void,
  onComplete?: () => void,
  onError?: (err: any) => void
) {  
  const req = ChatRequest.fromObject({ text: message, chatId: chatId });

  const stream = chat.Chat(req, {});

  stream.on("data", (res: ChatResponse) => {
    onMessage(res.text);
  });

  stream.on("end", () => {
    onComplete?.();
  });

  stream.on("error", (err) => {
    onError?.(err);
  });
}

const $chatList = atom<ChatInfo[]>([])

onMount($chatList, () => {

  chat.GetChatList(GetChatListRequest.fromObject({}), {})
  .then(value=>{
    $chatList.set(value.chats)
  })

  return () => {
    // Disabled mode
  }
})


export { doChat, $chatList }





