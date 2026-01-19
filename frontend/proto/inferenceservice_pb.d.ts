// package: sortedchat
// file: inferenceservice.proto

import * as jspb from "google-protobuf";

export class DownloadProgress extends jspb.Message {
  getFileSize(): number;
  setFileSize(value: number): void;

  getStatus(): DownloadStatusMap[keyof DownloadStatusMap];
  setStatus(value: DownloadStatusMap[keyof DownloadStatusMap]): void;

  getProgress(): number;
  setProgress(value: number): void;

  getSpeed(): number;
  setSpeed(value: number): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DownloadProgress.AsObject;
  static toObject(includeInstance: boolean, msg: DownloadProgress): DownloadProgress.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DownloadProgress, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DownloadProgress;
  static deserializeBinaryFromReader(message: DownloadProgress, reader: jspb.BinaryReader): DownloadProgress;
}

export namespace DownloadProgress {
  export type AsObject = {
    fileSize: number,
    status: DownloadStatusMap[keyof DownloadStatusMap],
    progress: number,
    speed: number,
  }
}

export class Model extends jspb.Message {
  getId(): string;
  setId(value: string): void;

  getName(): string;
  setName(value: string): void;

  getUrl(): string;
  setUrl(value: string): void;

  getProvider(): string;
  setProvider(value: string): void;

  getInputTokenCost(): number;
  setInputTokenCost(value: number): void;

  getOutputTokenCost(): number;
  setOutputTokenCost(value: number): void;

  hasProgress(): boolean;
  clearProgress(): void;
  getProgress(): DownloadProgress | undefined;
  setProgress(value?: DownloadProgress): void;

  getIsDownloaded(): boolean;
  setIsDownloaded(value: boolean): void;

  getIsDownloadable(): boolean;
  setIsDownloadable(value: boolean): void;

  getStatus(): DownloadStatusMap[keyof DownloadStatusMap];
  setStatus(value: DownloadStatusMap[keyof DownloadStatusMap]): void;

  getFilestoreId(): string;
  setFilestoreId(value: string): void;

  getIsEmbeddingModel(): boolean;
  setIsEmbeddingModel(value: boolean): void;

  getIsEnabled(): boolean;
  setIsEnabled(value: boolean): void;

  hasModelInfo(): boolean;
  clearModelInfo(): void;
  getModelInfo(): ModelInfo | undefined;
  setModelInfo(value?: ModelInfo): void;

  getCreatorName(): string;
  setCreatorName(value: string): void;

  getModifiedBy(): string;
  setModifiedBy(value: string): void;

  getDescription(): string;
  setDescription(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): Model.AsObject;
  static toObject(includeInstance: boolean, msg: Model): Model.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: Model, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): Model;
  static deserializeBinaryFromReader(message: Model, reader: jspb.BinaryReader): Model;
}

export namespace Model {
  export type AsObject = {
    id: string,
    name: string,
    url: string,
    provider: string,
    inputTokenCost: number,
    outputTokenCost: number,
    progress?: DownloadProgress.AsObject,
    isDownloaded: boolean,
    isDownloadable: boolean,
    status: DownloadStatusMap[keyof DownloadStatusMap],
    filestoreId: string,
    isEmbeddingModel: boolean,
    isEnabled: boolean,
    modelInfo?: ModelInfo.AsObject,
    creatorName: string,
    modifiedBy: string,
    description: string,
  }
}

export class ModelInfo extends jspb.Message {
  getHomePageUrl(): string;
  setHomePageUrl(value: string): void;

  getQuantization(): string;
  setQuantization(value: string): void;

  getDownloadSize(): string;
  setDownloadSize(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): ModelInfo.AsObject;
  static toObject(includeInstance: boolean, msg: ModelInfo): ModelInfo.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: ModelInfo, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): ModelInfo;
  static deserializeBinaryFromReader(message: ModelInfo, reader: jspb.BinaryReader): ModelInfo;
}

export namespace ModelInfo {
  export type AsObject = {
    homePageUrl: string,
    quantization: string,
    downloadSize: string,
  }
}

export class DownloadModelRequest extends jspb.Message {
  getModelId(): string;
  setModelId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DownloadModelRequest.AsObject;
  static toObject(includeInstance: boolean, msg: DownloadModelRequest): DownloadModelRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DownloadModelRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DownloadModelRequest;
  static deserializeBinaryFromReader(message: DownloadModelRequest, reader: jspb.BinaryReader): DownloadModelRequest;
}

export namespace DownloadModelRequest {
  export type AsObject = {
    modelId: string,
  }
}

export class DownloadModelResponse extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DownloadModelResponse.AsObject;
  static toObject(includeInstance: boolean, msg: DownloadModelResponse): DownloadModelResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DownloadModelResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DownloadModelResponse;
  static deserializeBinaryFromReader(message: DownloadModelResponse, reader: jspb.BinaryReader): DownloadModelResponse;
}

export namespace DownloadModelResponse {
  export type AsObject = {
    message: string,
  }
}

export class GetLLMModelsRequest extends jspb.Message {
  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetLLMModelsRequest.AsObject;
  static toObject(includeInstance: boolean, msg: GetLLMModelsRequest): GetLLMModelsRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetLLMModelsRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetLLMModelsRequest;
  static deserializeBinaryFromReader(message: GetLLMModelsRequest, reader: jspb.BinaryReader): GetLLMModelsRequest;
}

export namespace GetLLMModelsRequest {
  export type AsObject = {
  }
}

export class GetLLMModelsResponse extends jspb.Message {
  clearModelsList(): void;
  getModelsList(): Array<Model>;
  setModelsList(value: Array<Model>): void;
  addModels(value?: Model, index?: number): Model;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): GetLLMModelsResponse.AsObject;
  static toObject(includeInstance: boolean, msg: GetLLMModelsResponse): GetLLMModelsResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: GetLLMModelsResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): GetLLMModelsResponse;
  static deserializeBinaryFromReader(message: GetLLMModelsResponse, reader: jspb.BinaryReader): GetLLMModelsResponse;
}

export namespace GetLLMModelsResponse {
  export type AsObject = {
    modelsList: Array<Model.AsObject>,
  }
}

export class CancelDownloadRequest extends jspb.Message {
  getModelId(): string;
  setModelId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CancelDownloadRequest.AsObject;
  static toObject(includeInstance: boolean, msg: CancelDownloadRequest): CancelDownloadRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CancelDownloadRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CancelDownloadRequest;
  static deserializeBinaryFromReader(message: CancelDownloadRequest, reader: jspb.BinaryReader): CancelDownloadRequest;
}

export namespace CancelDownloadRequest {
  export type AsObject = {
    modelId: string,
  }
}

export class CancelDownloadResponse extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): CancelDownloadResponse.AsObject;
  static toObject(includeInstance: boolean, msg: CancelDownloadResponse): CancelDownloadResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: CancelDownloadResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): CancelDownloadResponse;
  static deserializeBinaryFromReader(message: CancelDownloadResponse, reader: jspb.BinaryReader): CancelDownloadResponse;
}

export namespace CancelDownloadResponse {
  export type AsObject = {
    message: string,
  }
}

export class DeleteModelRequest extends jspb.Message {
  getModelId(): string;
  setModelId(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DeleteModelRequest.AsObject;
  static toObject(includeInstance: boolean, msg: DeleteModelRequest): DeleteModelRequest.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DeleteModelRequest, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DeleteModelRequest;
  static deserializeBinaryFromReader(message: DeleteModelRequest, reader: jspb.BinaryReader): DeleteModelRequest;
}

export namespace DeleteModelRequest {
  export type AsObject = {
    modelId: string,
  }
}

export class DeleteModelResponse extends jspb.Message {
  getMessage(): string;
  setMessage(value: string): void;

  serializeBinary(): Uint8Array;
  toObject(includeInstance?: boolean): DeleteModelResponse.AsObject;
  static toObject(includeInstance: boolean, msg: DeleteModelResponse): DeleteModelResponse.AsObject;
  static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
  static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
  static serializeBinaryToWriter(message: DeleteModelResponse, writer: jspb.BinaryWriter): void;
  static deserializeBinary(bytes: Uint8Array): DeleteModelResponse;
  static deserializeBinaryFromReader(message: DeleteModelResponse, reader: jspb.BinaryReader): DeleteModelResponse;
}

export namespace DeleteModelResponse {
  export type AsObject = {
    message: string,
  }
}

export interface DownloadStatusMap {
  NONE: 0;
  PENDING: 1;
  DOWNLOADING: 2;
  COMPLETED: 3;
  FAILED: 4;
  CANCELLING: 5;
}

export const DownloadStatus: DownloadStatusMap;

