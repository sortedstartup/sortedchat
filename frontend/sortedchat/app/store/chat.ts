import { nanoquery } from "@nanostores/query";
import { ChatInfo, ChatMessage, ChatRequest, ChatResponse, GetChatListRequest, GetHistoryRequest, SortedChatClient } from "../../proto/chatservice"
import { atom, onMount } from 'nanostores'

var chat = new SortedChatClient(import.meta.env.VITE_API_URL)

const $chatList = atom<ChatInfo[]>([])
const $currentChatId = atom<string>("")
export const $currentChatMessage = atom<string>("")
export const $streamingMessage=atom<string>("")

export const doChatNano=(msg:string)=>{
  $currentChatMessage.set(msg)
  $streamingMessage.set("")
  doChat(msg,$currentChatId.get(),(chunk:string)=>{
    $streamingMessage.set($streamingMessage.get()+chunk)
  })
}

// when current chat id changes
// when a user clicks on a different chat
$currentChatId.listen((newValue, oldValue)=>{
  $streamingMessage.set("")
  $currentChatMessage.set("")
})

onMount($chatList, () => {

  chat.GetChatList(GetChatListRequest.fromObject({}), {})
  .then(value=>{
    $chatList.set(value.chats)
  })

  return () => {
    // Disabled mode
  }
})


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


export const [createFetcherStore, createMutatorStore] = nanoquery({
  fetcher: (...keys) => chat.GetHistory(GetHistoryRequest.fromObject({
    chatId:keys.join('')
  }),{}).then(r=>r.history)
});

const $currentChatMessages = createFetcherStore<ChatMessage[]>([$currentChatId]);

const $addMessage = createMutatorStore<ChatMessage>(async ({ data: message, revalidate, getCacheUpdater }) => {
    // You can either revalidate the author…
    // revalidate(`/api/users/${message}`);

    // …or you can optimistically update current cache.
    const [updateCache, post] = getCacheUpdater(`/api/post/${message}`);
    
    updateCache({ ...post, comments: [...post.comments, message] });

    // Even though `fetch` is called after calling `revalidate`, we will only
    // revalidate the keys after `fetch` resolves
    return fetch('…')
  })

export { doChat, $chatList, $currentChatMessages, $currentChatId }






