// package: sortedchat
// file: chatservice.proto

import * as jspb from "google-protobuf";

export class CreateChatRequest extends jspb.Message {
  getName(): string;
  setName(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateChatRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CreateChatRequest): CreateChatRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CreateChatRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateChatRequest;
  static deserializeBinaryFromReader(message: CreateChatRequest, reader: jspb.BinaryReader): CreateChatRequest;
}

export namespace CreateChatRequest {
  export type AsObject = {
    name: string,
  }
}

export class CreateChatResponse extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  getChatId(): string;
  setChatId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateChatResponse.AsObject;
  static toObject(includeInstance: boolean, msg: CreateChatResponse): CreateChatResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CreateChatResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateChatResponse;
  static deserializeBinaryFromReader(message: CreateChatResponse, reader: jspb.BinaryReader): CreateChatResponse;
}

export namespace CreateChatResponse {
  export type AsObject = {
    message: string,
    chatId: string,
  }
}

export class ChatRequest extends jspb.Message {
  getText(): string;
  setText(value: string): void;

  getChatid(): string;
  setChatid(value: string): void;

  getModel(): string;
  setModel(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ChatRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ChatRequest): ChatRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ChatRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ChatRequest;
  static deserializeBinaryFromReader(message: ChatRequest, reader: jspb.BinaryReader): ChatRequest;
}

export namespace ChatRequest {
  export type AsObject = {
    text: string,
    chatid: string,
    model: string,
  }
}

export class ChatResponse extends jspb.Message {
  getText(): string;
  setText(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ChatResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ChatResponse): ChatResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ChatResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ChatResponse;
  static deserializeBinaryFromReader(message: ChatResponse, reader: jspb.BinaryReader): ChatResponse;
}

export namespace ChatResponse {
  export type AsObject = {
    text: string,
  }
}

export class GetHistoryRequest extends jspb.Message {
  getChatid(): string;
  setChatid(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetHistoryRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetHistoryRequest): GetHistoryRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetHistoryRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetHistoryRequest;
  static deserializeBinaryFromReader(message: GetHistoryRequest, reader: jspb.BinaryReader): GetHistoryRequest;
}

export namespace GetHistoryRequest {
  export type AsObject = {
    chatid: string,
  }
}

export class GetHistoryResponse extends jspb.Message {
  clearHistoryList(): void;
  getHistoryList(): Array<ChatMessage>;
  setHistoryList(value: Array<ChatMessage>): void;
  addHistory(value?: ChatMessage, index?: number): ChatMessage;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetHistoryResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetHistoryResponse): GetHistoryResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetHistoryResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetHistoryResponse;
  static deserializeBinaryFromReader(message: GetHistoryResponse, reader: jspb.BinaryReader): GetHistoryResponse;
}

export namespace GetHistoryResponse {
  export type AsObject = {
    historyList: Array<ChatMessage.AsObject>,
  }
}

export class ChatMessage extends jspb.Message {
  getRole(): string;
  setRole(value: string): void;

  getContent(): string;
  setContent(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ChatMessage.AsObject;
  static toObject(includeInstance: boolean, msg: ChatMessage): ChatMessage.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ChatMessage, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ChatMessage;
  static deserializeBinaryFromReader(message: ChatMessage, reader: jspb.BinaryReader): ChatMessage;
}

export namespace ChatMessage {
  export type AsObject = {
    role: string,
    content: string,
  }
}

export class GetChatListRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetChatListRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetChatListRequest): GetChatListRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetChatListRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetChatListRequest;
  static deserializeBinaryFromReader(message: GetChatListRequest, reader: jspb.BinaryReader): GetChatListRequest;
}

export namespace GetChatListRequest {
  export type AsObject = {
  }
}

export class GetChatListResponse extends jspb.Message {
  clearChatsList(): void;
  getChatsList(): Array<ChatInfo>;
  setChatsList(value: Array<ChatInfo>): void;
  addChats(value?: ChatInfo, index?: number): ChatInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetChatListResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetChatListResponse): GetChatListResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetChatListResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetChatListResponse;
  static deserializeBinaryFromReader(message: GetChatListResponse, reader: jspb.BinaryReader): GetChatListResponse;
}

export namespace GetChatListResponse {
  export type AsObject = {
    chatsList: Array<ChatInfo.AsObject>,
  }
}

export class ChatInfo extends jspb.Message {
  getChatid(): string;
  setChatid(value: string): void;

  getName(): string;
  setName(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ChatInfo.AsObject;
  static toObject(includeInstance: boolean, msg: ChatInfo): ChatInfo.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ChatInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ChatInfo;
  static deserializeBinaryFromReader(message: ChatInfo, reader: jspb.BinaryReader): ChatInfo;
}

export namespace ChatInfo {
  export type AsObject = {
    chatid: string,
    name: string,
  }
}

export class ModelListInfo extends jspb.Message {
  getId(): string;
  setId(value: string): void;

  getLabel(): string;
  setLabel(value: string): void;

  getProvider(): string;
  setProvider(value: string): void;

  getUrl(): string;
  setUrl(value: string): void;

  getInputTokenCost(): number;
  setInputTokenCost(value: number): void;

  getOutputTokenCost(): number;
  setOutputTokenCost(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ModelListInfo.AsObject;
  static toObject(includeInstance: boolean, msg: ModelListInfo): ModelListInfo.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ModelListInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ModelListInfo;
  static deserializeBinaryFromReader(message: ModelListInfo, reader: jspb.BinaryReader): ModelListInfo;
}

export namespace ModelListInfo {
  export type AsObject = {
    id: string,
    label: string,
    provider: string,
    url: string,
    inputTokenCost: number,
    outputTokenCost: number,
  }
}

export class ListModelsRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ListModelsRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ListModelsRequest): ListModelsRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ListModelsRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ListModelsRequest;
  static deserializeBinaryFromReader(message: ListModelsRequest, reader: jspb.BinaryReader): ListModelsRequest;
}

export namespace ListModelsRequest {
  export type AsObject = {
  }
}

export class ListModelsResponse extends jspb.Message {
  clearModelsList(): void;
  getModelsList(): Array<ModelListInfo>;
  setModelsList(value: Array<ModelListInfo>): void;
  addModels(value?: ModelListInfo, index?: number): ModelListInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ListModelsResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ListModelsResponse): ListModelsResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ListModelsResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ListModelsResponse;
  static deserializeBinaryFromReader(message: ListModelsResponse, reader: jspb.BinaryReader): ListModelsResponse;
}

export namespace ListModelsResponse {
  export type AsObject = {
    modelsList: Array<ModelListInfo.AsObject>,
  }
}

