import { ChatRequest, SortedChatClient } from "../../proto/chatservice"

var chat = new SortedChatClient("http://localhost:8080")

async function doChat(msg: string) {
    const response = await chat.Chat(
        ChatRequest.fromObject({
            text: msg,
    }),
    {}
    )
    return response.text
}







