import { ChatRequest, HelloRequest, HelloResponse, SortedChatClient } from "../../proto/chatservice"

var chat = new SortedChatClient(import.meta.env.VITE_API_URL)

function doChat(
  message: string,
  onMessage: (chunk: string) => void,
  onComplete?: () => void,
  onError?: (err: any) => void
) {  
  const req = HelloRequest.fromObject({ text: message });

  const stream = chat.LotsOfReplies(req, {});

  stream.on("data", (res: HelloResponse) => {
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





