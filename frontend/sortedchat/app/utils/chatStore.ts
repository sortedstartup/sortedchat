
import { GetChatListRequest, SortedChatClient } from "../../proto/chatservice";
import { CreateChatRequest } from "../../proto/chatservice";
// import { grpc } from "@improbable-eng/grpc-web";

export interface Message {
  id: number;
  sender: string;
  content: string;
}

export interface ChatInfo {
  name: string;
  messages: Message[];
}

export interface ChatDataStore {
  [key: number]: ChatInfo;
}

class ChatStore {
  private static instance: ChatStore;
  private chatData: ChatDataStore = {};
  private client = new SortedChatClient(import.meta.env.VITE_API_URL);


  private constructor() { }

  public static getInstance(): ChatStore {
    if (!ChatStore.instance) {
      ChatStore.instance = new ChatStore();
    }
    return ChatStore.instance;
  }

  public getChatData(): ChatDataStore {
    return this.chatData;
  }

  public getChat(id: number): ChatInfo | undefined {
    return this.chatData[id];
  }

  public async createChat(id: number, name: string): Promise<string> {
    const request = new CreateChatRequest({ chatId: id.toString(), name:name });

    try {
      const response = await this.client.CreateChat(request, {});
      return response.message;
    } catch (err: any) {
      console.error("CreateChat RPC failed:", err.message);
      throw err;
    }
  }

  public addMessage(chatId: number, message: Message): void {
    if (!this.chatData[chatId]) {
      this.chatData[chatId] = {
        name: `Chat ${chatId}`,
        messages: []
      };
    }

    this.chatData[chatId].messages.push(message);
  }

  public async getAllChats(): Promise<Array<{ id: number; name: string }>> {
    const request = new GetChatListRequest();

    try {
      const response = await this.client.GetChatList(request, {});
      const chats = response.chats;
      console.log(chats)

      return chats.map(chat => ({
        id: parseInt(chat.chatId),
        name: chat.name
      }));
    } catch (err: any) {
      console.error("Failed to fetch chat list:", err.message);
      throw err;
    }
  }

}

export const chatStore = ChatStore.getInstance();
