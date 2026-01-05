// package: sortedchat
// file: chatservice.proto

import * as jspb from "google-protobuf";
import * as google_protobuf_struct_pb from "google-protobuf/google/protobuf/struct_pb";

export class RenameItemRequest extends jspb.Message {
  getItemId(): string;
  setItemId(value: string): void;

  getName(): string;
  setName(value: string): void;

  getItemType(): RenameItemRequest.ItemTypeMap[keyof RenameItemRequest.ItemTypeMap];
  setItemType(value: RenameItemRequest.ItemTypeMap[keyof RenameItemRequest.ItemTypeMap]): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RenameItemRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RenameItemRequest): RenameItemRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RenameItemRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RenameItemRequest;
  static deserializeBinaryFromReader(message: RenameItemRequest, reader: jspb.BinaryReader): RenameItemRequest;
}

export namespace RenameItemRequest {
  export type AsObject = {
    itemId: string,
    name: string,
    itemType: RenameItemRequest.ItemTypeMap[keyof RenameItemRequest.ItemTypeMap],
  }

  export interface ItemTypeMap {
    CHAT: 0;
    PROJECT: 1;
  }

  export const ItemType: ItemTypeMap;
}

export class RenameItemResponse extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RenameItemResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RenameItemResponse): RenameItemResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RenameItemResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RenameItemResponse;
  static deserializeBinaryFromReader(message: RenameItemResponse, reader: jspb.BinaryReader): RenameItemResponse;
}

export namespace RenameItemResponse {
  export type AsObject = {
    message: string,
  }
}

export class RestoreChatRequest extends jspb.Message {
  getChatId(): string;
  setChatId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RestoreChatRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RestoreChatRequest): RestoreChatRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RestoreChatRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RestoreChatRequest;
  static deserializeBinaryFromReader(message: RestoreChatRequest, reader: jspb.BinaryReader): RestoreChatRequest;
}

export namespace RestoreChatRequest {
  export type AsObject = {
    chatId: string,
  }
}

export class RestoreChatResponse extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RestoreChatResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RestoreChatResponse): RestoreChatResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RestoreChatResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RestoreChatResponse;
  static deserializeBinaryFromReader(message: RestoreChatResponse, reader: jspb.BinaryReader): RestoreChatResponse;
}

export namespace RestoreChatResponse {
  export type AsObject = {
    message: string,
  }
}

export class DeleteChatRequest extends jspb.Message {
  getChatId(): string;
  setChatId(value: string): void;

  getOperation(): DeleteChatRequest.OperationMap[keyof DeleteChatRequest.OperationMap];
  setOperation(value: DeleteChatRequest.OperationMap[keyof DeleteChatRequest.OperationMap]): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DeleteChatRequest.AsObject;
  static toObject(includeInstance: boolean, msg: DeleteChatRequest): DeleteChatRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DeleteChatRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DeleteChatRequest;
  static deserializeBinaryFromReader(message: DeleteChatRequest, reader: jspb.BinaryReader): DeleteChatRequest;
}

export namespace DeleteChatRequest {
  export type AsObject = {
    chatId: string,
    operation: DeleteChatRequest.OperationMap[keyof DeleteChatRequest.OperationMap],
  }

  export interface OperationMap {
    DELETE: 0;
    SOFT_DELETE: 1;
  }

  export const Operation: OperationMap;
}

export class DeleteChatResponse extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DeleteChatResponse.AsObject;
  static toObject(includeInstance: boolean, msg: DeleteChatResponse): DeleteChatResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DeleteChatResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DeleteChatResponse;
  static deserializeBinaryFromReader(message: DeleteChatResponse, reader: jspb.BinaryReader): DeleteChatResponse;
}

export namespace DeleteChatResponse {
  export type AsObject = {
    message: string,
  }
}

export class DeleteDocumentRequest extends jspb.Message {
  getProjectId(): string;
  setProjectId(value: string): void;

  getDocId(): string;
  setDocId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DeleteDocumentRequest.AsObject;
  static toObject(includeInstance: boolean, msg: DeleteDocumentRequest): DeleteDocumentRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DeleteDocumentRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DeleteDocumentRequest;
  static deserializeBinaryFromReader(message: DeleteDocumentRequest, reader: jspb.BinaryReader): DeleteDocumentRequest;
}

export namespace DeleteDocumentRequest {
  export type AsObject = {
    projectId: string,
    docId: string,
  }
}

export class DeleteDocumentResponse extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DeleteDocumentResponse.AsObject;
  static toObject(includeInstance: boolean, msg: DeleteDocumentResponse): DeleteDocumentResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DeleteDocumentResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DeleteDocumentResponse;
  static deserializeBinaryFromReader(message: DeleteDocumentResponse, reader: jspb.BinaryReader): DeleteDocumentResponse;
}

export namespace DeleteDocumentResponse {
  export type AsObject = {
    message: string,
  }
}

export class RAGDocumentReferenceRequest extends jspb.Message {
  getMessageId(): string;
  setMessageId(value: string): void;

  getProjectId(): string;
  setProjectId(value: string): void;

  getDocid(): string;
  setDocid(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RAGDocumentReferenceRequest.AsObject;
  static toObject(includeInstance: boolean, msg: RAGDocumentReferenceRequest): RAGDocumentReferenceRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RAGDocumentReferenceRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RAGDocumentReferenceRequest;
  static deserializeBinaryFromReader(message: RAGDocumentReferenceRequest, reader: jspb.BinaryReader): RAGDocumentReferenceRequest;
}

export namespace RAGDocumentReferenceRequest {
  export type AsObject = {
    messageId: string,
    projectId: string,
    docid: string,
  }
}

export class RAGDocumentReferenceResponse extends jspb.Message {
  hasReference(): boolean;
  clearReference(): void;
  getReference(): RAGDocumentReference | undefined;
  setReference(value?: RAGDocumentReference): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RAGDocumentReferenceResponse.AsObject;
  static toObject(includeInstance: boolean, msg: RAGDocumentReferenceResponse): RAGDocumentReferenceResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RAGDocumentReferenceResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RAGDocumentReferenceResponse;
  static deserializeBinaryFromReader(message: RAGDocumentReferenceResponse, reader: jspb.BinaryReader): RAGDocumentReferenceResponse;
}

export namespace RAGDocumentReferenceResponse {
  export type AsObject = {
    reference?: RAGDocumentReference.AsObject,
  }
}

export class TestConnectionRequest extends jspb.Message {
  getUrl(): string;
  setUrl(value: string): void;

  getConnectionType(): ConnectionTypeMap[keyof ConnectionTypeMap];
  setConnectionType(value: ConnectionTypeMap[keyof ConnectionTypeMap]): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): TestConnectionRequest.AsObject;
  static toObject(includeInstance: boolean, msg: TestConnectionRequest): TestConnectionRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: TestConnectionRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): TestConnectionRequest;
  static deserializeBinaryFromReader(message: TestConnectionRequest, reader: jspb.BinaryReader): TestConnectionRequest;
}

export namespace TestConnectionRequest {
  export type AsObject = {
    url: string,
    connectionType: ConnectionTypeMap[keyof ConnectionTypeMap],
  }
}

export class TestConnectionResponse extends jspb.Message {
  getSuccess(): boolean;
  setSuccess(value: boolean): void;

  getMessage(): string;
  setMessage(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): TestConnectionResponse.AsObject;
  static toObject(includeInstance: boolean, msg: TestConnectionResponse): TestConnectionResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: TestConnectionResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): TestConnectionResponse;
  static deserializeBinaryFromReader(message: TestConnectionResponse, reader: jspb.BinaryReader): TestConnectionResponse;
}

export namespace TestConnectionResponse {
  export type AsObject = {
    success: boolean,
    message: string,
  }
}

export class Settings extends jspb.Message {
  getOpenaiApiKey(): string;
  setOpenaiApiKey(value: string): void;

  getOpenaiApiUrl(): string;
  setOpenaiApiUrl(value: string): void;

  getGeminiApiKey(): string;
  setGeminiApiKey(value: string): void;

  getOllamaUrl(): string;
  setOllamaUrl(value: string): void;

  getClaudeApiUrl(): string;
  setClaudeApiUrl(value: string): void;

  getClaudeApiKey(): string;
  setClaudeApiKey(value: string): void;

  getGeminiApiUrl(): string;
  setGeminiApiUrl(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Settings.AsObject;
  static toObject(includeInstance: boolean, msg: Settings): Settings.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Settings, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Settings;
  static deserializeBinaryFromReader(message: Settings, reader: jspb.BinaryReader): Settings;
}

export namespace Settings {
  export type AsObject = {
    openaiApiKey: string,
    openaiApiUrl: string,
    geminiApiKey: string,
    ollamaUrl: string,
    claudeApiUrl: string,
    claudeApiKey: string,
    geminiApiUrl: string,
  }
}

export class GetSettingRequest extends jspb.Message {
  getName(): string;
  setName(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetSettingRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetSettingRequest): GetSettingRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetSettingRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetSettingRequest;
  static deserializeBinaryFromReader(message: GetSettingRequest, reader: jspb.BinaryReader): GetSettingRequest;
}

export namespace GetSettingRequest {
  export type AsObject = {
    name: string,
  }
}

export class GetSettingResponse extends jspb.Message {
  hasSettings(): boolean;
  clearSettings(): void;
  getSettings(): google_protobuf_struct_pb.Struct | undefined;
  setSettings(value?: google_protobuf_struct_pb.Struct): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetSettingResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetSettingResponse): GetSettingResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetSettingResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetSettingResponse;
  static deserializeBinaryFromReader(message: GetSettingResponse, reader: jspb.BinaryReader): GetSettingResponse;
}

export namespace GetSettingResponse {
  export type AsObject = {
    settings?: google_protobuf_struct_pb.Struct.AsObject,
  }
}

export class SetSettingRequest extends jspb.Message {
  getName(): string;
  setName(value: string): void;

  hasSettings(): boolean;
  clearSettings(): void;
  getSettings(): google_protobuf_struct_pb.Struct | undefined;
  setSettings(value?: google_protobuf_struct_pb.Struct): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SetSettingRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SetSettingRequest): SetSettingRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SetSettingRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SetSettingRequest;
  static deserializeBinaryFromReader(message: SetSettingRequest, reader: jspb.BinaryReader): SetSettingRequest;
}

export namespace SetSettingRequest {
  export type AsObject = {
    name: string,
    settings?: google_protobuf_struct_pb.Struct.AsObject,
  }
}

export class SetSettingResponse extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SetSettingResponse.AsObject;
  static toObject(includeInstance: boolean, msg: SetSettingResponse): SetSettingResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SetSettingResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SetSettingResponse;
  static deserializeBinaryFromReader(message: SetSettingResponse, reader: jspb.BinaryReader): SetSettingResponse;
}

export namespace SetSettingResponse {
  export type AsObject = {
    message: string,
  }
}

export class ProviderSettings extends jspb.Message {
  getApiUrl(): string;
  setApiUrl(value: string): void;

  getApiKey(): string;
  setApiKey(value: string): void;

  getIsEnabled(): boolean;
  setIsEnabled(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ProviderSettings.AsObject;
  static toObject(includeInstance: boolean, msg: ProviderSettings): ProviderSettings.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ProviderSettings, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ProviderSettings;
  static deserializeBinaryFromReader(message: ProviderSettings, reader: jspb.BinaryReader): ProviderSettings;
}

export namespace ProviderSettings {
  export type AsObject = {
    apiUrl: string,
    apiKey: string,
    isEnabled: boolean,
  }
}

export class GetProviderSettingRequest extends jspb.Message {
  getName(): string;
  setName(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetProviderSettingRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetProviderSettingRequest): GetProviderSettingRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetProviderSettingRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetProviderSettingRequest;
  static deserializeBinaryFromReader(message: GetProviderSettingRequest, reader: jspb.BinaryReader): GetProviderSettingRequest;
}

export namespace GetProviderSettingRequest {
  export type AsObject = {
    name: string,
  }
}

export class GetProviderSettingResponse extends jspb.Message {
  hasSettings(): boolean;
  clearSettings(): void;
  getSettings(): ProviderSettings | undefined;
  setSettings(value?: ProviderSettings): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetProviderSettingResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetProviderSettingResponse): GetProviderSettingResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetProviderSettingResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetProviderSettingResponse;
  static deserializeBinaryFromReader(message: GetProviderSettingResponse, reader: jspb.BinaryReader): GetProviderSettingResponse;
}

export namespace GetProviderSettingResponse {
  export type AsObject = {
    settings?: ProviderSettings.AsObject,
  }
}

export class SetProviderSettingRequest extends jspb.Message {
  getName(): string;
  setName(value: string): void;

  hasSettings(): boolean;
  clearSettings(): void;
  getSettings(): ProviderSettings | undefined;
  setSettings(value?: ProviderSettings): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SetProviderSettingRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SetProviderSettingRequest): SetProviderSettingRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SetProviderSettingRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SetProviderSettingRequest;
  static deserializeBinaryFromReader(message: SetProviderSettingRequest, reader: jspb.BinaryReader): SetProviderSettingRequest;
}

export namespace SetProviderSettingRequest {
  export type AsObject = {
    name: string,
    settings?: ProviderSettings.AsObject,
  }
}

export class SetProviderSettingResponse extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SetProviderSettingResponse.AsObject;
  static toObject(includeInstance: boolean, msg: SetProviderSettingResponse): SetProviderSettingResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SetProviderSettingResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SetProviderSettingResponse;
  static deserializeBinaryFromReader(message: SetProviderSettingResponse, reader: jspb.BinaryReader): SetProviderSettingResponse;
}

export namespace SetProviderSettingResponse {
  export type AsObject = {
    message: string,
  }
}

export class GetAllProviderSettingsRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetAllProviderSettingsRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetAllProviderSettingsRequest): GetAllProviderSettingsRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetAllProviderSettingsRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetAllProviderSettingsRequest;
  static deserializeBinaryFromReader(message: GetAllProviderSettingsRequest, reader: jspb.BinaryReader): GetAllProviderSettingsRequest;
}

export namespace GetAllProviderSettingsRequest {
  export type AsObject = {
  }
}

export class GetAllProviderSettingsResponse extends jspb.Message {
  getSettingsMap(): jspb.Map<string, ProviderSettings>;
  clearSettingsMap(): void;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetAllProviderSettingsResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetAllProviderSettingsResponse): GetAllProviderSettingsResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetAllProviderSettingsResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetAllProviderSettingsResponse;
  static deserializeBinaryFromReader(message: GetAllProviderSettingsResponse, reader: jspb.BinaryReader): GetAllProviderSettingsResponse;
}

export namespace GetAllProviderSettingsResponse {
  export type AsObject = {
    settingsMap: Array<[string, ProviderSettings.AsObject]>,
  }
}

export class SetAllProviderSettingsRequest extends jspb.Message {
  getSettingsMap(): jspb.Map<string, ProviderSettings>;
  clearSettingsMap(): void;
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SetAllProviderSettingsRequest.AsObject;
  static toObject(includeInstance: boolean, msg: SetAllProviderSettingsRequest): SetAllProviderSettingsRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SetAllProviderSettingsRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SetAllProviderSettingsRequest;
  static deserializeBinaryFromReader(message: SetAllProviderSettingsRequest, reader: jspb.BinaryReader): SetAllProviderSettingsRequest;
}

export namespace SetAllProviderSettingsRequest {
  export type AsObject = {
    settingsMap: Array<[string, ProviderSettings.AsObject]>,
  }
}

export class SetAllProviderSettingsResponse extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SetAllProviderSettingsResponse.AsObject;
  static toObject(includeInstance: boolean, msg: SetAllProviderSettingsResponse): SetAllProviderSettingsResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SetAllProviderSettingsResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SetAllProviderSettingsResponse;
  static deserializeBinaryFromReader(message: SetAllProviderSettingsResponse, reader: jspb.BinaryReader): SetAllProviderSettingsResponse;
}

export namespace SetAllProviderSettingsResponse {
  export type AsObject = {
    message: string,
  }
}

export class IsFirstBootRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): IsFirstBootRequest.AsObject;
  static toObject(includeInstance: boolean, msg: IsFirstBootRequest): IsFirstBootRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: IsFirstBootRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): IsFirstBootRequest;
  static deserializeBinaryFromReader(message: IsFirstBootRequest, reader: jspb.BinaryReader): IsFirstBootRequest;
}

export namespace IsFirstBootRequest {
  export type AsObject = {
  }
}

export class IsFirstBootResponse extends jspb.Message {
  getIsFirstBoot(): boolean;
  setIsFirstBoot(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): IsFirstBootResponse.AsObject;
  static toObject(includeInstance: boolean, msg: IsFirstBootResponse): IsFirstBootResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: IsFirstBootResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): IsFirstBootResponse;
  static deserializeBinaryFromReader(message: IsFirstBootResponse, reader: jspb.BinaryReader): IsFirstBootResponse;
}

export namespace IsFirstBootResponse {
  export type AsObject = {
    isFirstBoot: boolean,
  }
}

export class CreateChatRequest extends jspb.Message {
  getName(): string;
  setName(value: string): void;

  getProjectId(): string;
  setProjectId(value: string): void;

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
    projectId: string,
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

export class ProjectContext extends jspb.Message {
  getProjectId(): string;
  setProjectId(value: string): void;

  getRagEnabled(): boolean;
  setRagEnabled(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ProjectContext.AsObject;
  static toObject(includeInstance: boolean, msg: ProjectContext): ProjectContext.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ProjectContext, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ProjectContext;
  static deserializeBinaryFromReader(message: ProjectContext, reader: jspb.BinaryReader): ProjectContext;
}

export namespace ProjectContext {
  export type AsObject = {
    projectId: string,
    ragEnabled: boolean,
  }
}

export class MessageContent extends jspb.Message {
  getType(): string;
  setType(value: string): void;

  getText(): string;
  setText(value: string): void;

  hasImageUrl(): boolean;
  clearImageUrl(): void;
  getImageUrl(): ImageUrl | undefined;
  setImageUrl(value?: ImageUrl): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): MessageContent.AsObject;
  static toObject(includeInstance: boolean, msg: MessageContent): MessageContent.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: MessageContent, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): MessageContent;
  static deserializeBinaryFromReader(message: MessageContent, reader: jspb.BinaryReader): MessageContent;
}

export namespace MessageContent {
  export type AsObject = {
    type: string,
    text: string,
    imageUrl?: ImageUrl.AsObject,
  }
}

export class ImageUrl extends jspb.Message {
  getUrl(): string;
  setUrl(value: string): void;

  getDetail(): string;
  setDetail(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ImageUrl.AsObject;
  static toObject(includeInstance: boolean, msg: ImageUrl): ImageUrl.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ImageUrl, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ImageUrl;
  static deserializeBinaryFromReader(message: ImageUrl, reader: jspb.BinaryReader): ImageUrl;
}

export namespace ImageUrl {
  export type AsObject = {
    url: string,
    detail: string,
  }
}

export class ChatRequest extends jspb.Message {
  getText(): string;
  setText(value: string): void;

  getChatid(): string;
  setChatid(value: string): void;

  getModel(): string;
  setModel(value: string): void;

  hasProjectContext(): boolean;
  clearProjectContext(): void;
  getProjectContext(): ProjectContext | undefined;
  setProjectContext(value?: ProjectContext): void;

  clearContentsList(): void;
  getContentsList(): Array<MessageContent>;
  setContentsList(value: Array<MessageContent>): void;
  addContents(value?: MessageContent, index?: number): MessageContent;

  getProvider(): string;
  setProvider(value: string): void;

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
    projectContext?: ProjectContext.AsObject,
    contentsList: Array<MessageContent.AsObject>,
    provider: string,
  }
}

export class ChatResponse extends jspb.Message {
  hasText(): boolean;
  clearText(): void;
  getText(): string;
  setText(value: string): void;

  hasSummary(): boolean;
  clearSummary(): void;
  getSummary(): ResponseSummary | undefined;
  setSummary(value?: ResponseSummary): void;

  hasRequestMessageId(): boolean;
  clearRequestMessageId(): void;
  getRequestMessageId(): string;
  setRequestMessageId(value: string): void;

  hasDocumentReference(): boolean;
  clearDocumentReference(): void;
  getDocumentReference(): RAGDocumentReferenceSummaryList | undefined;
  setDocumentReference(value?: RAGDocumentReferenceSummaryList): void;

  hasChatMetadata(): boolean;
  clearChatMetadata(): void;
  getChatMetadata(): ChatInfo | undefined;
  setChatMetadata(value?: ChatInfo): void;

  hasProgress(): boolean;
  clearProgress(): void;
  getProgress(): ChatProgress | undefined;
  setProgress(value?: ChatProgress): void;

  getResponseCase(): ChatResponse.ResponseCase;
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
    summary?: ResponseSummary.AsObject,
    requestMessageId: string,
    documentReference?: RAGDocumentReferenceSummaryList.AsObject,
    chatMetadata?: ChatInfo.AsObject,
    progress?: ChatProgress.AsObject,
  }

  export enum ResponseCase {
    RESPONSE_NOT_SET = 0,
    TEXT = 1,
    SUMMARY = 2,
    REQUEST_MESSAGE_ID = 3,
    DOCUMENT_REFERENCE = 4,
    CHAT_METADATA = 5,
    PROGRESS = 6,
  }
}

export class ChatProgress extends jspb.Message {
  getState(): ChatProgress.StateMap[keyof ChatProgress.StateMap];
  setState(value: ChatProgress.StateMap[keyof ChatProgress.StateMap]): void;

  getMessage(): string;
  setMessage(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ChatProgress.AsObject;
  static toObject(includeInstance: boolean, msg: ChatProgress): ChatProgress.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ChatProgress, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ChatProgress;
  static deserializeBinaryFromReader(message: ChatProgress, reader: jspb.BinaryReader): ChatProgress;
}

export namespace ChatProgress {
  export type AsObject = {
    state: ChatProgress.StateMap[keyof ChatProgress.StateMap],
    message: string,
  }

  export interface StateMap {
    SENDING_REQUEST_TO_LLM: 0;
    REQUEST_SENT_TO_LLM: 1;
    FIRST_RESPONSE_RECEIVED: 2;
    FIRST_TOKEN_RECEIVED: 3;
    TOKENS_STREAMING: 4;
    TOKENS_STOPPED: 5;
  }

  export const State: StateMap;
}

export class RAGDocumentReferenceSummaryList extends jspb.Message {
  clearSummaryList(): void;
  getSummaryList(): Array<RAGDocumentReferenceSummaryList.Summary>;
  setSummaryList(value: Array<RAGDocumentReferenceSummaryList.Summary>): void;
  addSummary(value?: RAGDocumentReferenceSummaryList.Summary, index?: number): RAGDocumentReferenceSummaryList.Summary;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RAGDocumentReferenceSummaryList.AsObject;
  static toObject(includeInstance: boolean, msg: RAGDocumentReferenceSummaryList): RAGDocumentReferenceSummaryList.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RAGDocumentReferenceSummaryList, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RAGDocumentReferenceSummaryList;
  static deserializeBinaryFromReader(message: RAGDocumentReferenceSummaryList, reader: jspb.BinaryReader): RAGDocumentReferenceSummaryList;
}

export namespace RAGDocumentReferenceSummaryList {
  export type AsObject = {
    summaryList: Array<RAGDocumentReferenceSummaryList.Summary.AsObject>,
  }

  export class Summary extends jspb.Message {
    getDocId(): string;
    setDocId(value: string): void;

    getFileName(): string;
    setFileName(value: string): void;

    getChunkcount(): number;
    setChunkcount(value: number): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Summary.AsObject;
    static toObject(includeInstance: boolean, msg: Summary): Summary.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Summary, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Summary;
    static deserializeBinaryFromReader(message: Summary, reader: jspb.BinaryReader): Summary;
  }

  export namespace Summary {
    export type AsObject = {
      docId: string,
      fileName: string,
      chunkcount: number,
    }
  }
}

export class RAGDocumentReference extends jspb.Message {
  getDocId(): string;
  setDocId(value: string): void;

  getFileName(): string;
  setFileName(value: string): void;

  clearChunksList(): void;
  getChunksList(): Array<RAGDocumentReference.Chunk>;
  setChunksList(value: Array<RAGDocumentReference.Chunk>): void;
  addChunks(value?: RAGDocumentReference.Chunk, index?: number): RAGDocumentReference.Chunk;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): RAGDocumentReference.AsObject;
  static toObject(includeInstance: boolean, msg: RAGDocumentReference): RAGDocumentReference.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: RAGDocumentReference, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): RAGDocumentReference;
  static deserializeBinaryFromReader(message: RAGDocumentReference, reader: jspb.BinaryReader): RAGDocumentReference;
}

export namespace RAGDocumentReference {
  export type AsObject = {
    docId: string,
    fileName: string,
    chunksList: Array<RAGDocumentReference.Chunk.AsObject>,
  }

  export class Chunk extends jspb.Message {
    getChunkText(): string;
    setChunkText(value: string): void;

    getStartByte(): number;
    setStartByte(value: number): void;

    getEndByte(): number;
    setEndByte(value: number): void;

    getSimillarity(): number;
    setSimillarity(value: number): void;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Chunk.AsObject;
    static toObject(includeInstance: boolean, msg: Chunk): Chunk.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Chunk, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Chunk;
    static deserializeBinaryFromReader(message: Chunk, reader: jspb.BinaryReader): Chunk;
  }

  export namespace Chunk {
    export type AsObject = {
      chunkText: string,
      startByte: number,
      endByte: number,
      simillarity: number,
    }
  }
}

export class ResponseSummary extends jspb.Message {
  getMessageId(): string;
  setMessageId(value: string): void;

  getModel(): string;
  setModel(value: string): void;

  getInputTokens(): number;
  setInputTokens(value: number): void;

  getOutputTokens(): number;
  setOutputTokens(value: number): void;

  getCachedTokens(): number;
  setCachedTokens(value: number): void;

  getCost(): number;
  setCost(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ResponseSummary.AsObject;
  static toObject(includeInstance: boolean, msg: ResponseSummary): ResponseSummary.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ResponseSummary, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ResponseSummary;
  static deserializeBinaryFromReader(message: ResponseSummary, reader: jspb.BinaryReader): ResponseSummary;
}

export namespace ResponseSummary {
  export type AsObject = {
    messageId: string,
    model: string,
    inputTokens: number,
    outputTokens: number,
    cachedTokens: number,
    cost: number,
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

  hasChatMetadata(): boolean;
  clearChatMetadata(): void;
  getChatMetadata(): ChatInfo | undefined;
  setChatMetadata(value?: ChatInfo): void;

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
    chatMetadata?: ChatInfo.AsObject,
  }
}

export class ChatMessage extends jspb.Message {
  getRole(): string;
  setRole(value: string): void;

  getContent(): string;
  setContent(value: string): void;

  getMessageId(): string;
  setMessageId(value: string): void;

  clearReferencesList(): void;
  getReferencesList(): Array<RAGDocumentReference>;
  setReferencesList(value: Array<RAGDocumentReference>): void;
  addReferences(value?: RAGDocumentReference, index?: number): RAGDocumentReference;

  getRagEnabled(): boolean;
  setRagEnabled(value: boolean): void;

  getModel(): string;
  setModel(value: string): void;

  getInputTokens(): number;
  setInputTokens(value: number): void;

  getOutputTokens(): number;
  setOutputTokens(value: number): void;

  getCachedTokens(): number;
  setCachedTokens(value: number): void;

  getCost(): number;
  setCost(value: number): void;

  clearContentsList(): void;
  getContentsList(): Array<MessageContent>;
  setContentsList(value: Array<MessageContent>): void;
  addContents(value?: MessageContent, index?: number): MessageContent;

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
    messageId: string,
    referencesList: Array<RAGDocumentReference.AsObject>,
    ragEnabled: boolean,
    model: string,
    inputTokens: number,
    outputTokens: number,
    cachedTokens: number,
    cost: number,
    contentsList: Array<MessageContent.AsObject>,
  }
}

export class GetChatListRequest extends jspb.Message {
  getProjectId(): string;
  setProjectId(value: string): void;

  getSoftDeleted(): boolean;
  setSoftDeleted(value: boolean): void;

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
    projectId: string,
    softDeleted: boolean,
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

  getCost(): number;
  setCost(value: number): void;

  getInputTokenCount(): number;
  setInputTokenCount(value: number): void;

  getOutputTokenCount(): number;
  setOutputTokenCount(value: number): void;

  getCachedTokenCount(): number;
  setCachedTokenCount(value: number): void;

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
    cost: number,
    inputTokenCount: number,
    outputTokenCount: number,
    cachedTokenCount: number,
  }
}

export class ChatSearchRequest extends jspb.Message {
  getQuery(): string;
  setQuery(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ChatSearchRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ChatSearchRequest): ChatSearchRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ChatSearchRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ChatSearchRequest;
  static deserializeBinaryFromReader(message: ChatSearchRequest, reader: jspb.BinaryReader): ChatSearchRequest;
}

export namespace ChatSearchRequest {
  export type AsObject = {
    query: string,
  }
}

export class SearchResult extends jspb.Message {
  getChatName(): string;
  setChatName(value: string): void;

  getChatId(): string;
  setChatId(value: string): void;

  getMatchedText(): string;
  setMatchedText(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): SearchResult.AsObject;
  static toObject(includeInstance: boolean, msg: SearchResult): SearchResult.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: SearchResult, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): SearchResult;
  static deserializeBinaryFromReader(message: SearchResult, reader: jspb.BinaryReader): SearchResult;
}

export namespace SearchResult {
  export type AsObject = {
    chatName: string,
    chatId: string,
    matchedText: string,
  }
}

export class ChatSearchResponse extends jspb.Message {
  getQuery(): string;
  setQuery(value: string): void;

  clearResultsList(): void;
  getResultsList(): Array<SearchResult>;
  setResultsList(value: Array<SearchResult>): void;
  addResults(value?: SearchResult, index?: number): SearchResult;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ChatSearchResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ChatSearchResponse): ChatSearchResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ChatSearchResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ChatSearchResponse;
  static deserializeBinaryFromReader(message: ChatSearchResponse, reader: jspb.BinaryReader): ChatSearchResponse;
}

export namespace ChatSearchResponse {
  export type AsObject = {
    query: string,
    resultsList: Array<SearchResult.AsObject>,
  }
}

export class CreateProjectRequest extends jspb.Message {
  getName(): string;
  setName(value: string): void;

  getDescription(): string;
  setDescription(value: string): void;

  getAdditionalData(): string;
  setAdditionalData(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateProjectRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CreateProjectRequest): CreateProjectRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CreateProjectRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateProjectRequest;
  static deserializeBinaryFromReader(message: CreateProjectRequest, reader: jspb.BinaryReader): CreateProjectRequest;
}

export namespace CreateProjectRequest {
  export type AsObject = {
    name: string,
    description: string,
    additionalData: string,
  }
}

export class CreateProjectResponse extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  getProjectId(): string;
  setProjectId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CreateProjectResponse.AsObject;
  static toObject(includeInstance: boolean, msg: CreateProjectResponse): CreateProjectResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CreateProjectResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CreateProjectResponse;
  static deserializeBinaryFromReader(message: CreateProjectResponse, reader: jspb.BinaryReader): CreateProjectResponse;
}

export namespace CreateProjectResponse {
  export type AsObject = {
    message: string,
    projectId: string,
  }
}

export class GetProjectsRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetProjectsRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetProjectsRequest): GetProjectsRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetProjectsRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetProjectsRequest;
  static deserializeBinaryFromReader(message: GetProjectsRequest, reader: jspb.BinaryReader): GetProjectsRequest;
}

export namespace GetProjectsRequest {
  export type AsObject = {
  }
}

export class GetProjectsResponse extends jspb.Message {
  clearProjectsList(): void;
  getProjectsList(): Array<Project>;
  setProjectsList(value: Array<Project>): void;
  addProjects(value?: Project, index?: number): Project;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetProjectsResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetProjectsResponse): GetProjectsResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetProjectsResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetProjectsResponse;
  static deserializeBinaryFromReader(message: GetProjectsResponse, reader: jspb.BinaryReader): GetProjectsResponse;
}

export namespace GetProjectsResponse {
  export type AsObject = {
    projectsList: Array<Project.AsObject>,
  }
}

export class Project extends jspb.Message {
  getId(): string;
  setId(value: string): void;

  getName(): string;
  setName(value: string): void;

  getDescription(): string;
  setDescription(value: string): void;

  getAdditionalData(): string;
  setAdditionalData(value: string): void;

  getCreatedAt(): string;
  setCreatedAt(value: string): void;

  getUpdatedAt(): string;
  setUpdatedAt(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Project.AsObject;
  static toObject(includeInstance: boolean, msg: Project): Project.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Project, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Project;
  static deserializeBinaryFromReader(message: Project, reader: jspb.BinaryReader): Project;
}

export namespace Project {
  export type AsObject = {
    id: string,
    name: string,
    description: string,
    additionalData: string,
    createdAt: string,
    updatedAt: string,
  }
}

export class ListDocumentsRequest extends jspb.Message {
  getProjectId(): string;
  setProjectId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ListDocumentsRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ListDocumentsRequest): ListDocumentsRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ListDocumentsRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ListDocumentsRequest;
  static deserializeBinaryFromReader(message: ListDocumentsRequest, reader: jspb.BinaryReader): ListDocumentsRequest;
}

export namespace ListDocumentsRequest {
  export type AsObject = {
    projectId: string,
  }
}

export class ListDocumentsResponse extends jspb.Message {
  clearDocumentsList(): void;
  getDocumentsList(): Array<Document>;
  setDocumentsList(value: Array<Document>): void;
  addDocuments(value?: Document, index?: number): Document;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ListDocumentsResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ListDocumentsResponse): ListDocumentsResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ListDocumentsResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ListDocumentsResponse;
  static deserializeBinaryFromReader(message: ListDocumentsResponse, reader: jspb.BinaryReader): ListDocumentsResponse;
}

export namespace ListDocumentsResponse {
  export type AsObject = {
    documentsList: Array<Document.AsObject>,
  }
}

export class Document extends jspb.Message {
  getId(): number;
  setId(value: number): void;

  getProjectId(): string;
  setProjectId(value: string): void;

  getDocsId(): string;
  setDocsId(value: string): void;

  getFileName(): string;
  setFileName(value: string): void;

  getCreatedAt(): string;
  setCreatedAt(value: string): void;

  getUpdatedAt(): string;
  setUpdatedAt(value: string): void;

  getEmbeddingStatus(): Embedding_StatusMap[keyof Embedding_StatusMap];
  setEmbeddingStatus(value: Embedding_StatusMap[keyof Embedding_StatusMap]): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Document.AsObject;
  static toObject(includeInstance: boolean, msg: Document): Document.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Document, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Document;
  static deserializeBinaryFromReader(message: Document, reader: jspb.BinaryReader): Document;
}

export namespace Document {
  export type AsObject = {
    id: number,
    projectId: string,
    docsId: string,
    fileName: string,
    createdAt: string,
    updatedAt: string,
    embeddingStatus: Embedding_StatusMap[keyof Embedding_StatusMap],
  }
}

export class GenerateEmbeddingRequest extends jspb.Message {
  getProjectId(): string;
  setProjectId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GenerateEmbeddingRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GenerateEmbeddingRequest): GenerateEmbeddingRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GenerateEmbeddingRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GenerateEmbeddingRequest;
  static deserializeBinaryFromReader(message: GenerateEmbeddingRequest, reader: jspb.BinaryReader): GenerateEmbeddingRequest;
}

export namespace GenerateEmbeddingRequest {
  export type AsObject = {
    projectId: string,
  }
}

export class GenerateEmbeddingResponse extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GenerateEmbeddingResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GenerateEmbeddingResponse): GenerateEmbeddingResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GenerateEmbeddingResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GenerateEmbeddingResponse;
  static deserializeBinaryFromReader(message: GenerateEmbeddingResponse, reader: jspb.BinaryReader): GenerateEmbeddingResponse;
}

export namespace GenerateEmbeddingResponse {
  export type AsObject = {
    message: string,
  }
}

export class GenerateChatNameRequest extends jspb.Message {
  getChatId(): string;
  setChatId(value: string): void;

  getMessage(): string;
  setMessage(value: string): void;

  getModel(): string;
  setModel(value: string): void;

  getProvider(): string;
  setProvider(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GenerateChatNameRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GenerateChatNameRequest): GenerateChatNameRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GenerateChatNameRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GenerateChatNameRequest;
  static deserializeBinaryFromReader(message: GenerateChatNameRequest, reader: jspb.BinaryReader): GenerateChatNameRequest;
}

export namespace GenerateChatNameRequest {
  export type AsObject = {
    chatId: string,
    message: string,
    model: string,
    provider: string,
  }
}

export class GenerateChatNameResponse extends jspb.Message {
  getChatName(): string;
  setChatName(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GenerateChatNameResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GenerateChatNameResponse): GenerateChatNameResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GenerateChatNameResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GenerateChatNameResponse;
  static deserializeBinaryFromReader(message: GenerateChatNameResponse, reader: jspb.BinaryReader): GenerateChatNameResponse;
}

export namespace GenerateChatNameResponse {
  export type AsObject = {
    chatName: string,
  }
}

export class BranchAChatRequest extends jspb.Message {
  getSourceChatId(): string;
  setSourceChatId(value: string): void;

  getBranchFromMessageId(): string;
  setBranchFromMessageId(value: string): void;

  getBranchName(): string;
  setBranchName(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): BranchAChatRequest.AsObject;
  static toObject(includeInstance: boolean, msg: BranchAChatRequest): BranchAChatRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: BranchAChatRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): BranchAChatRequest;
  static deserializeBinaryFromReader(message: BranchAChatRequest, reader: jspb.BinaryReader): BranchAChatRequest;
}

export namespace BranchAChatRequest {
  export type AsObject = {
    sourceChatId: string,
    branchFromMessageId: string,
    branchName: string,
  }
}

export class BranchAChatResponse extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  getNewChatId(): string;
  setNewChatId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): BranchAChatResponse.AsObject;
  static toObject(includeInstance: boolean, msg: BranchAChatResponse): BranchAChatResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: BranchAChatResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): BranchAChatResponse;
  static deserializeBinaryFromReader(message: BranchAChatResponse, reader: jspb.BinaryReader): BranchAChatResponse;
}

export namespace BranchAChatResponse {
  export type AsObject = {
    message: string,
    newChatId: string,
  }
}

export class ListChatBranchRequest extends jspb.Message {
  getChatId(): string;
  setChatId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ListChatBranchRequest.AsObject;
  static toObject(includeInstance: boolean, msg: ListChatBranchRequest): ListChatBranchRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ListChatBranchRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ListChatBranchRequest;
  static deserializeBinaryFromReader(message: ListChatBranchRequest, reader: jspb.BinaryReader): ListChatBranchRequest;
}

export namespace ListChatBranchRequest {
  export type AsObject = {
    chatId: string,
  }
}

export class ListChatBranchResponse extends jspb.Message {
  clearBranchChatListList(): void;
  getBranchChatListList(): Array<ChatInfo>;
  setBranchChatListList(value: Array<ChatInfo>): void;
  addBranchChatList(value?: ChatInfo, index?: number): ChatInfo;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ListChatBranchResponse.AsObject;
  static toObject(includeInstance: boolean, msg: ListChatBranchResponse): ListChatBranchResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ListChatBranchResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ListChatBranchResponse;
  static deserializeBinaryFromReader(message: ListChatBranchResponse, reader: jspb.BinaryReader): ListChatBranchResponse;
}

export namespace ListChatBranchResponse {
  export type AsObject = {
    branchChatListList: Array<ChatInfo.AsObject>,
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

  hasCapabilities(): boolean;
  clearCapabilities(): void;
  getCapabilities(): ModelCapabilities | undefined;
  setCapabilities(value?: ModelCapabilities): void;

  getIsDownloaded(): boolean;
  setIsDownloaded(value: boolean): void;

  getIsDownloadable(): boolean;
  setIsDownloadable(value: boolean): void;

  getIsEmbeddingModel(): boolean;
  setIsEmbeddingModel(value: boolean): void;

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
    capabilities?: ModelCapabilities.AsObject,
    isDownloaded: boolean,
    isDownloadable: boolean,
    isEmbeddingModel: boolean,
  }
}

export class ModelCapabilities extends jspb.Message {
  hasText(): boolean;
  clearText(): void;
  getText(): Capability | undefined;
  setText(value?: Capability): void;

  hasAudio(): boolean;
  clearAudio(): void;
  getAudio(): Capability | undefined;
  setAudio(value?: Capability): void;

  hasVideo(): boolean;
  clearVideo(): void;
  getVideo(): Capability | undefined;
  setVideo(value?: Capability): void;

  hasImage(): boolean;
  clearImage(): void;
  getImage(): Capability | undefined;
  setImage(value?: Capability): void;

  getRealtime(): boolean;
  setRealtime(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ModelCapabilities.AsObject;
  static toObject(includeInstance: boolean, msg: ModelCapabilities): ModelCapabilities.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ModelCapabilities, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ModelCapabilities;
  static deserializeBinaryFromReader(message: ModelCapabilities, reader: jspb.BinaryReader): ModelCapabilities;
}

export namespace ModelCapabilities {
  export type AsObject = {
    text?: Capability.AsObject,
    audio?: Capability.AsObject,
    video?: Capability.AsObject,
    image?: Capability.AsObject,
    realtime: boolean,
  }
}

export class Capability extends jspb.Message {
  getInput(): boolean;
  setInput(value: boolean): void;

  getOutput(): boolean;
  setOutput(value: boolean): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Capability.AsObject;
  static toObject(includeInstance: boolean, msg: Capability): Capability.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Capability, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Capability;
  static deserializeBinaryFromReader(message: Capability, reader: jspb.BinaryReader): Capability;
}

export namespace Capability {
  export type AsObject = {
    input: boolean,
    output: boolean,
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

export interface ConnectionTypeMap {
  OLLAMA: 0;
  OPENAI: 1;
}

export const ConnectionType: ConnectionTypeMap;

export interface Embedding_StatusMap {
  STATUS_QUEUED: 0;
  STATUS_IN_PROGRESS: 1;
  STATUS_ERROR: 2;
  STATUS_SUCCESS: 3;
}

export const Embedding_Status: Embedding_StatusMap;

