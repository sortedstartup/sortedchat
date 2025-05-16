import { ChatRequest, SortedChatClient } from "../../proto/chatservice"

var chat = new SortedChatClient(import.meta.env.VITE_API_URL)

async function doChat(msg: string) {
    const response = await chat.Chat(
        ChatRequest.fromObject({
            text: msg,
    }),
    {}
    )
    return response.text
}

export { doChat }





