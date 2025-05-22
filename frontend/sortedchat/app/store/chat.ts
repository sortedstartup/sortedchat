import { nanoquery } from "@nanostores/query";
import { ChatInfo, ChatMessage, ChatRequest, ChatResponse, CreateChatRequest, GetChatListRequest, GetHistoryRequest, SortedChatClient } from "../../proto/chatservice"
import { atom, onMount } from 'nanostores'

var chat = new SortedChatClient(import.meta.env.VITE_API_URL)

export const [createFetcherStore, createMutatorStore] = nanoquery({
  fetcher: (...keys) => chat.GetHistory(GetHistoryRequest.fromObject({
    chatId:keys.join('')
  }),{}).then(r=>r.history)
});

// --- stores ---
export const $chatList = atom<ChatInfo[]>([])

export const $currentChatId = atom<string>("")
export const $currentChatMessages = createFetcherStore<ChatMessage[]>([$currentChatId]);

export const $currentChatMessage = atom<string>("")
export const $streamingMessage=atom<string>("")

// --- stores ----

// --- state management ---
export const createNewChat = () => {

  // todo: this should always be generated on the server
  const uuid = uuidv4()
  const promise = chat.CreateChat(CreateChatRequest.fromObject({
    chatId: uuid,
    name: uuid
  }),{}).then(r=>{
    //todo: need to do better chat history management
    // we use to mutate the cache, no need to fetch whole history
    getChatList()
  })

  return {uuid: uuid, promise: promise }
}
export const doChat=(msg:string)=>{

  $currentChatMessage.set(msg)
  $streamingMessage.set("")

  // grpc call
  const stream = chat.Chat(ChatRequest.fromObject({
                  text: msg,
                  chatId: $currentChatId.get()
                }), {});

  stream.on("data", (res: ChatResponse) => {
    $streamingMessage.set($streamingMessage.get()+res.text)
  });

  stream.on("end", () => {    
  });

  stream.on("error", (err) => {
  });
}

// when current chat id changes
// when a user clicks on a different chat
$currentChatId.listen((newValue, oldValue)=>{
  $streamingMessage.set("")
  $currentChatMessage.set("")
})

const getChatList = () => {
    chat.GetChatList(GetChatListRequest.fromObject({}), {})
  .then(value=>{
    $chatList.set(value.chats)
  })
}

// load chat history of first use
onMount($chatList, () => {
  getChatList()
  return () => {
    // Disabled mode
  }
})


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


function uuidv4() {
  const bytes = crypto.getRandomValues(new Uint8Array(16));
  // Set RFC-4122 version & variant bits
  bytes[6] = (bytes[6] & 0x0f) | 0x40;
  bytes[8] = (bytes[8] & 0x3f) | 0x80;
  return [...bytes].map((b,i)=>
    (b.toString(16).padStart(2,'0') + ([4,6,8,10].includes(i)?'-':'')))
    .join('');
}




