// Define the types for our chat data
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

// Create a singleton store that can be imported across components
class ChatStore {
  private static instance: ChatStore;
  private chatData: ChatDataStore = {
    1: {
      name: "Chat about AI",
      messages: [
        { id: 1, sender: "ai", content: "Hello! How can I help you today?" },
        { id: 2, sender: "user", content: "I'm looking to learn more about AI." },
        { id: 3, sender: "ai", content: "Great! What specific aspect of AI are you interested in?" },
      ],
    },
    2: {
      name: "Project planning",
      messages: [
        { id: 1, sender: "ai", content: "Let's plan your project. What are you working on?" },
        { id: 2, sender: "user", content: "I'm building a new e-commerce website." },
        { id: 3, sender: "ai", content: "Excellent! What features do you need for your e-commerce site?" },
      ],
    },
    3: {
      name: "Brainstorming session",
      messages: [
        { id: 1, sender: "ai", content: "Ready to brainstorm! What topic are we exploring today?" },
        { id: 2, sender: "user", content: "I need ideas for a marketing campaign." },
        { id: 3, sender: "ai", content: "Let's generate some marketing campaign ideas. What's your target audience?" },
      ],
    },
  };

  private constructor() {}

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

  public createChat(id: number, name: string, initialMessage: string): ChatInfo {
    const newChat: ChatInfo = {
      name,
      messages: [
        { id: 1, sender: "user", content: initialMessage }
      ]
    };
    
    this.chatData[id] = newChat;
    return newChat;
  }

  public addMessage(chatId: number, message: Message): void {
    if (!this.chatData[chatId]) {
      this.createChat(chatId, `Chat ${chatId}`, "");
    }
    
    this.chatData[chatId].messages.push(message);
  }

  public getAllChats(): Array<{id: number, name: string}> {
    return Object.entries(this.chatData).map(([id, data]) => ({
      id: Number(id),
      name: data.name
    }));
  }
}

// Export a singleton instance
export const chatStore = ChatStore.getInstance(); 