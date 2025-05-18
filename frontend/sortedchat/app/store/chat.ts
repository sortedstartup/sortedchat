import { ChatRequest, ChatResponse, SortedChatClient } from "../../proto/chatservice"

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

export { doChat }





