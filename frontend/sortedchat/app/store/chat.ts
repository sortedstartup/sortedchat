import { HelloRequest } from "proto/chatservice_pb";
import { ChatRequest, SortedChatClient } from "../../proto/chatservice"

var chat = new SortedChatClient(import.meta.env.VITE_API_URL)

async function doChat(msg: string) {    
  const request = new HelloRequest();
  request.setText(msg);

  const stream = chat.lotsOfReplies(request, {});
}

export { doChat }





