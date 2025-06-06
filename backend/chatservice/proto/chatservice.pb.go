// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.30.2
// source: chatservice.proto

package proto

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type CreateChatRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Name          string                 `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateChatRequest) Reset() {
	*x = CreateChatRequest{}
	mi := &file_chatservice_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateChatRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateChatRequest) ProtoMessage() {}

func (x *CreateChatRequest) ProtoReflect() protoreflect.Message {
	mi := &file_chatservice_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateChatRequest.ProtoReflect.Descriptor instead.
func (*CreateChatRequest) Descriptor() ([]byte, []int) {
	return file_chatservice_proto_rawDescGZIP(), []int{0}
}

func (x *CreateChatRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type CreateChatResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Message       string                 `protobuf:"bytes,1,opt,name=message,proto3" json:"message,omitempty"`
	ChatId        string                 `protobuf:"bytes,2,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *CreateChatResponse) Reset() {
	*x = CreateChatResponse{}
	mi := &file_chatservice_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *CreateChatResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CreateChatResponse) ProtoMessage() {}

func (x *CreateChatResponse) ProtoReflect() protoreflect.Message {
	mi := &file_chatservice_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CreateChatResponse.ProtoReflect.Descriptor instead.
func (*CreateChatResponse) Descriptor() ([]byte, []int) {
	return file_chatservice_proto_rawDescGZIP(), []int{1}
}

func (x *CreateChatResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

func (x *CreateChatResponse) GetChatId() string {
	if x != nil {
		return x.ChatId
	}
	return ""
}

type ChatRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Text          string                 `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
	ChatId        string                 `protobuf:"bytes,2,opt,name=chatId,proto3" json:"chatId,omitempty"`
	Model         string                 `protobuf:"bytes,3,opt,name=model,proto3" json:"model,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ChatRequest) Reset() {
	*x = ChatRequest{}
	mi := &file_chatservice_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChatRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChatRequest) ProtoMessage() {}

func (x *ChatRequest) ProtoReflect() protoreflect.Message {
	mi := &file_chatservice_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChatRequest.ProtoReflect.Descriptor instead.
func (*ChatRequest) Descriptor() ([]byte, []int) {
	return file_chatservice_proto_rawDescGZIP(), []int{2}
}

func (x *ChatRequest) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

func (x *ChatRequest) GetChatId() string {
	if x != nil {
		return x.ChatId
	}
	return ""
}

func (x *ChatRequest) GetModel() string {
	if x != nil {
		return x.Model
	}
	return ""
}

type ChatResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Text          string                 `protobuf:"bytes,1,opt,name=text,proto3" json:"text,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ChatResponse) Reset() {
	*x = ChatResponse{}
	mi := &file_chatservice_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChatResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChatResponse) ProtoMessage() {}

func (x *ChatResponse) ProtoReflect() protoreflect.Message {
	mi := &file_chatservice_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChatResponse.ProtoReflect.Descriptor instead.
func (*ChatResponse) Descriptor() ([]byte, []int) {
	return file_chatservice_proto_rawDescGZIP(), []int{3}
}

func (x *ChatResponse) GetText() string {
	if x != nil {
		return x.Text
	}
	return ""
}

type GetHistoryRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ChatId        string                 `protobuf:"bytes,1,opt,name=chatId,proto3" json:"chatId,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetHistoryRequest) Reset() {
	*x = GetHistoryRequest{}
	mi := &file_chatservice_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetHistoryRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetHistoryRequest) ProtoMessage() {}

func (x *GetHistoryRequest) ProtoReflect() protoreflect.Message {
	mi := &file_chatservice_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetHistoryRequest.ProtoReflect.Descriptor instead.
func (*GetHistoryRequest) Descriptor() ([]byte, []int) {
	return file_chatservice_proto_rawDescGZIP(), []int{4}
}

func (x *GetHistoryRequest) GetChatId() string {
	if x != nil {
		return x.ChatId
	}
	return ""
}

type GetHistoryResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	History       []*ChatMessage         `protobuf:"bytes,1,rep,name=history,proto3" json:"history,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetHistoryResponse) Reset() {
	*x = GetHistoryResponse{}
	mi := &file_chatservice_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetHistoryResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetHistoryResponse) ProtoMessage() {}

func (x *GetHistoryResponse) ProtoReflect() protoreflect.Message {
	mi := &file_chatservice_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetHistoryResponse.ProtoReflect.Descriptor instead.
func (*GetHistoryResponse) Descriptor() ([]byte, []int) {
	return file_chatservice_proto_rawDescGZIP(), []int{5}
}

func (x *GetHistoryResponse) GetHistory() []*ChatMessage {
	if x != nil {
		return x.History
	}
	return nil
}

type ChatMessage struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Role          string                 `protobuf:"bytes,1,opt,name=role,proto3" json:"role,omitempty"`
	Content       string                 `protobuf:"bytes,2,opt,name=content,proto3" json:"content,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ChatMessage) Reset() {
	*x = ChatMessage{}
	mi := &file_chatservice_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChatMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChatMessage) ProtoMessage() {}

func (x *ChatMessage) ProtoReflect() protoreflect.Message {
	mi := &file_chatservice_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChatMessage.ProtoReflect.Descriptor instead.
func (*ChatMessage) Descriptor() ([]byte, []int) {
	return file_chatservice_proto_rawDescGZIP(), []int{6}
}

func (x *ChatMessage) GetRole() string {
	if x != nil {
		return x.Role
	}
	return ""
}

func (x *ChatMessage) GetContent() string {
	if x != nil {
		return x.Content
	}
	return ""
}

type GetChatListRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetChatListRequest) Reset() {
	*x = GetChatListRequest{}
	mi := &file_chatservice_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetChatListRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetChatListRequest) ProtoMessage() {}

func (x *GetChatListRequest) ProtoReflect() protoreflect.Message {
	mi := &file_chatservice_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetChatListRequest.ProtoReflect.Descriptor instead.
func (*GetChatListRequest) Descriptor() ([]byte, []int) {
	return file_chatservice_proto_rawDescGZIP(), []int{7}
}

type GetChatListResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Chats         []*ChatInfo            `protobuf:"bytes,1,rep,name=chats,proto3" json:"chats,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *GetChatListResponse) Reset() {
	*x = GetChatListResponse{}
	mi := &file_chatservice_proto_msgTypes[8]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GetChatListResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GetChatListResponse) ProtoMessage() {}

func (x *GetChatListResponse) ProtoReflect() protoreflect.Message {
	mi := &file_chatservice_proto_msgTypes[8]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GetChatListResponse.ProtoReflect.Descriptor instead.
func (*GetChatListResponse) Descriptor() ([]byte, []int) {
	return file_chatservice_proto_rawDescGZIP(), []int{8}
}

func (x *GetChatListResponse) GetChats() []*ChatInfo {
	if x != nil {
		return x.Chats
	}
	return nil
}

type ChatInfo struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ChatId        string                 `protobuf:"bytes,1,opt,name=chatId,proto3" json:"chatId,omitempty"`
	Name          string                 `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ChatInfo) Reset() {
	*x = ChatInfo{}
	mi := &file_chatservice_proto_msgTypes[9]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChatInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChatInfo) ProtoMessage() {}

func (x *ChatInfo) ProtoReflect() protoreflect.Message {
	mi := &file_chatservice_proto_msgTypes[9]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChatInfo.ProtoReflect.Descriptor instead.
func (*ChatInfo) Descriptor() ([]byte, []int) {
	return file_chatservice_proto_rawDescGZIP(), []int{9}
}

func (x *ChatInfo) GetChatId() string {
	if x != nil {
		return x.ChatId
	}
	return ""
}

func (x *ChatInfo) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type ModelListInfo struct {
	state           protoimpl.MessageState `protogen:"open.v1"`
	Id              string                 `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Label           string                 `protobuf:"bytes,2,opt,name=label,proto3" json:"label,omitempty"`
	Provider        string                 `protobuf:"bytes,3,opt,name=provider,proto3" json:"provider,omitempty"`
	Url             string                 `protobuf:"bytes,4,opt,name=url,proto3" json:"url,omitempty"`
	InputTokenCost  float32                `protobuf:"fixed32,5,opt,name=input_token_cost,json=inputTokenCost,proto3" json:"input_token_cost,omitempty"`
	OutputTokenCost float32                `protobuf:"fixed32,6,opt,name=output_token_cost,json=outputTokenCost,proto3" json:"output_token_cost,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *ModelListInfo) Reset() {
	*x = ModelListInfo{}
	mi := &file_chatservice_proto_msgTypes[10]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ModelListInfo) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ModelListInfo) ProtoMessage() {}

func (x *ModelListInfo) ProtoReflect() protoreflect.Message {
	mi := &file_chatservice_proto_msgTypes[10]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ModelListInfo.ProtoReflect.Descriptor instead.
func (*ModelListInfo) Descriptor() ([]byte, []int) {
	return file_chatservice_proto_rawDescGZIP(), []int{10}
}

func (x *ModelListInfo) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *ModelListInfo) GetLabel() string {
	if x != nil {
		return x.Label
	}
	return ""
}

func (x *ModelListInfo) GetProvider() string {
	if x != nil {
		return x.Provider
	}
	return ""
}

func (x *ModelListInfo) GetUrl() string {
	if x != nil {
		return x.Url
	}
	return ""
}

func (x *ModelListInfo) GetInputTokenCost() float32 {
	if x != nil {
		return x.InputTokenCost
	}
	return 0
}

func (x *ModelListInfo) GetOutputTokenCost() float32 {
	if x != nil {
		return x.OutputTokenCost
	}
	return 0
}

type ListModelsRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListModelsRequest) Reset() {
	*x = ListModelsRequest{}
	mi := &file_chatservice_proto_msgTypes[11]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListModelsRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListModelsRequest) ProtoMessage() {}

func (x *ListModelsRequest) ProtoReflect() protoreflect.Message {
	mi := &file_chatservice_proto_msgTypes[11]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListModelsRequest.ProtoReflect.Descriptor instead.
func (*ListModelsRequest) Descriptor() ([]byte, []int) {
	return file_chatservice_proto_rawDescGZIP(), []int{11}
}

type ListModelsResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Models        []*ModelListInfo       `protobuf:"bytes,1,rep,name=models,proto3" json:"models,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ListModelsResponse) Reset() {
	*x = ListModelsResponse{}
	mi := &file_chatservice_proto_msgTypes[12]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ListModelsResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ListModelsResponse) ProtoMessage() {}

func (x *ListModelsResponse) ProtoReflect() protoreflect.Message {
	mi := &file_chatservice_proto_msgTypes[12]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ListModelsResponse.ProtoReflect.Descriptor instead.
func (*ListModelsResponse) Descriptor() ([]byte, []int) {
	return file_chatservice_proto_rawDescGZIP(), []int{12}
}

func (x *ListModelsResponse) GetModels() []*ModelListInfo {
	if x != nil {
		return x.Models
	}
	return nil
}

type ChatSearchRequest struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Query         string                 `protobuf:"bytes,1,opt,name=query,proto3" json:"query,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ChatSearchRequest) Reset() {
	*x = ChatSearchRequest{}
	mi := &file_chatservice_proto_msgTypes[13]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChatSearchRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChatSearchRequest) ProtoMessage() {}

func (x *ChatSearchRequest) ProtoReflect() protoreflect.Message {
	mi := &file_chatservice_proto_msgTypes[13]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChatSearchRequest.ProtoReflect.Descriptor instead.
func (*ChatSearchRequest) Descriptor() ([]byte, []int) {
	return file_chatservice_proto_rawDescGZIP(), []int{13}
}

func (x *ChatSearchRequest) GetQuery() string {
	if x != nil {
		return x.Query
	}
	return ""
}

type SearchResult struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	ChatName      string                 `protobuf:"bytes,1,opt,name=chat_name,json=chatName,proto3" json:"chat_name,omitempty"`
	ChatId        string                 `protobuf:"bytes,2,opt,name=chat_id,json=chatId,proto3" json:"chat_id,omitempty"`
	MatchedText   string                 `protobuf:"bytes,3,opt,name=matched_text,json=matchedText,proto3" json:"matched_text,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SearchResult) Reset() {
	*x = SearchResult{}
	mi := &file_chatservice_proto_msgTypes[14]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SearchResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SearchResult) ProtoMessage() {}

func (x *SearchResult) ProtoReflect() protoreflect.Message {
	mi := &file_chatservice_proto_msgTypes[14]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SearchResult.ProtoReflect.Descriptor instead.
func (*SearchResult) Descriptor() ([]byte, []int) {
	return file_chatservice_proto_rawDescGZIP(), []int{14}
}

func (x *SearchResult) GetChatName() string {
	if x != nil {
		return x.ChatName
	}
	return ""
}

func (x *SearchResult) GetChatId() string {
	if x != nil {
		return x.ChatId
	}
	return ""
}

func (x *SearchResult) GetMatchedText() string {
	if x != nil {
		return x.MatchedText
	}
	return ""
}

type ChatSearchResponse struct {
	state         protoimpl.MessageState `protogen:"open.v1"`
	Query         string                 `protobuf:"bytes,1,opt,name=query,proto3" json:"query,omitempty"`
	Results       []*SearchResult        `protobuf:"bytes,2,rep,name=results,proto3" json:"results,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *ChatSearchResponse) Reset() {
	*x = ChatSearchResponse{}
	mi := &file_chatservice_proto_msgTypes[15]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *ChatSearchResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ChatSearchResponse) ProtoMessage() {}

func (x *ChatSearchResponse) ProtoReflect() protoreflect.Message {
	mi := &file_chatservice_proto_msgTypes[15]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ChatSearchResponse.ProtoReflect.Descriptor instead.
func (*ChatSearchResponse) Descriptor() ([]byte, []int) {
	return file_chatservice_proto_rawDescGZIP(), []int{15}
}

func (x *ChatSearchResponse) GetQuery() string {
	if x != nil {
		return x.Query
	}
	return ""
}

func (x *ChatSearchResponse) GetResults() []*SearchResult {
	if x != nil {
		return x.Results
	}
	return nil
}

var File_chatservice_proto protoreflect.FileDescriptor

const file_chatservice_proto_rawDesc = "" +
	"\n" +
	"\x11chatservice.proto\x12\n" +
	"sortedchat\"'\n" +
	"\x11CreateChatRequest\x12\x12\n" +
	"\x04name\x18\x01 \x01(\tR\x04name\"G\n" +
	"\x12CreateChatResponse\x12\x18\n" +
	"\amessage\x18\x01 \x01(\tR\amessage\x12\x17\n" +
	"\achat_id\x18\x02 \x01(\tR\x06chatId\"O\n" +
	"\vChatRequest\x12\x12\n" +
	"\x04text\x18\x01 \x01(\tR\x04text\x12\x16\n" +
	"\x06chatId\x18\x02 \x01(\tR\x06chatId\x12\x14\n" +
	"\x05model\x18\x03 \x01(\tR\x05model\"\"\n" +
	"\fChatResponse\x12\x12\n" +
	"\x04text\x18\x01 \x01(\tR\x04text\"+\n" +
	"\x11GetHistoryRequest\x12\x16\n" +
	"\x06chatId\x18\x01 \x01(\tR\x06chatId\"G\n" +
	"\x12GetHistoryResponse\x121\n" +
	"\ahistory\x18\x01 \x03(\v2\x17.sortedchat.ChatMessageR\ahistory\";\n" +
	"\vChatMessage\x12\x12\n" +
	"\x04role\x18\x01 \x01(\tR\x04role\x12\x18\n" +
	"\acontent\x18\x02 \x01(\tR\acontent\"\x14\n" +
	"\x12GetChatListRequest\"A\n" +
	"\x13GetChatListResponse\x12*\n" +
	"\x05chats\x18\x01 \x03(\v2\x14.sortedchat.ChatInfoR\x05chats\"6\n" +
	"\bChatInfo\x12\x16\n" +
	"\x06chatId\x18\x01 \x01(\tR\x06chatId\x12\x12\n" +
	"\x04name\x18\x02 \x01(\tR\x04name\"\xb9\x01\n" +
	"\rModelListInfo\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\tR\x02id\x12\x14\n" +
	"\x05label\x18\x02 \x01(\tR\x05label\x12\x1a\n" +
	"\bprovider\x18\x03 \x01(\tR\bprovider\x12\x10\n" +
	"\x03url\x18\x04 \x01(\tR\x03url\x12(\n" +
	"\x10input_token_cost\x18\x05 \x01(\x02R\x0einputTokenCost\x12*\n" +
	"\x11output_token_cost\x18\x06 \x01(\x02R\x0foutputTokenCost\"\x13\n" +
	"\x11ListModelsRequest\"G\n" +
	"\x12ListModelsResponse\x121\n" +
	"\x06models\x18\x01 \x03(\v2\x19.sortedchat.ModelListInfoR\x06models\")\n" +
	"\x11ChatSearchRequest\x12\x14\n" +
	"\x05query\x18\x01 \x01(\tR\x05query\"g\n" +
	"\fSearchResult\x12\x1b\n" +
	"\tchat_name\x18\x01 \x01(\tR\bchatName\x12\x17\n" +
	"\achat_id\x18\x02 \x01(\tR\x06chatId\x12!\n" +
	"\fmatched_text\x18\x03 \x01(\tR\vmatchedText\"^\n" +
	"\x12ChatSearchResponse\x12\x14\n" +
	"\x05query\x18\x01 \x01(\tR\x05query\x122\n" +
	"\aresults\x18\x02 \x03(\v2\x18.sortedchat.SearchResultR\aresults2\xcc\x03\n" +
	"\n" +
	"SortedChat\x12;\n" +
	"\x04Chat\x12\x17.sortedchat.ChatRequest\x1a\x18.sortedchat.ChatResponse0\x01\x12K\n" +
	"\n" +
	"GetHistory\x12\x1d.sortedchat.GetHistoryRequest\x1a\x1e.sortedchat.GetHistoryResponse\x12N\n" +
	"\vGetChatList\x12\x1e.sortedchat.GetChatListRequest\x1a\x1f.sortedchat.GetChatListResponse\x12K\n" +
	"\n" +
	"CreateChat\x12\x1d.sortedchat.CreateChatRequest\x1a\x1e.sortedchat.CreateChatResponse\x12J\n" +
	"\tListModel\x12\x1d.sortedchat.ListModelsRequest\x1a\x1e.sortedchat.ListModelsResponse\x12K\n" +
	"\n" +
	"SearchChat\x12\x1d.sortedchat.ChatSearchRequest\x1a\x1e.sortedchat.ChatSearchResponseB!Z\x1fsortedstartup/chatservice/protob\x06proto3"

var (
	file_chatservice_proto_rawDescOnce sync.Once
	file_chatservice_proto_rawDescData []byte
)

func file_chatservice_proto_rawDescGZIP() []byte {
	file_chatservice_proto_rawDescOnce.Do(func() {
		file_chatservice_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_chatservice_proto_rawDesc), len(file_chatservice_proto_rawDesc)))
	})
	return file_chatservice_proto_rawDescData
}

var file_chatservice_proto_msgTypes = make([]protoimpl.MessageInfo, 16)
var file_chatservice_proto_goTypes = []any{
	(*CreateChatRequest)(nil),   // 0: sortedchat.CreateChatRequest
	(*CreateChatResponse)(nil),  // 1: sortedchat.CreateChatResponse
	(*ChatRequest)(nil),         // 2: sortedchat.ChatRequest
	(*ChatResponse)(nil),        // 3: sortedchat.ChatResponse
	(*GetHistoryRequest)(nil),   // 4: sortedchat.GetHistoryRequest
	(*GetHistoryResponse)(nil),  // 5: sortedchat.GetHistoryResponse
	(*ChatMessage)(nil),         // 6: sortedchat.ChatMessage
	(*GetChatListRequest)(nil),  // 7: sortedchat.GetChatListRequest
	(*GetChatListResponse)(nil), // 8: sortedchat.GetChatListResponse
	(*ChatInfo)(nil),            // 9: sortedchat.ChatInfo
	(*ModelListInfo)(nil),       // 10: sortedchat.ModelListInfo
	(*ListModelsRequest)(nil),   // 11: sortedchat.ListModelsRequest
	(*ListModelsResponse)(nil),  // 12: sortedchat.ListModelsResponse
	(*ChatSearchRequest)(nil),   // 13: sortedchat.ChatSearchRequest
	(*SearchResult)(nil),        // 14: sortedchat.SearchResult
	(*ChatSearchResponse)(nil),  // 15: sortedchat.ChatSearchResponse
}
var file_chatservice_proto_depIdxs = []int32{
	6,  // 0: sortedchat.GetHistoryResponse.history:type_name -> sortedchat.ChatMessage
	9,  // 1: sortedchat.GetChatListResponse.chats:type_name -> sortedchat.ChatInfo
	10, // 2: sortedchat.ListModelsResponse.models:type_name -> sortedchat.ModelListInfo
	14, // 3: sortedchat.ChatSearchResponse.results:type_name -> sortedchat.SearchResult
	2,  // 4: sortedchat.SortedChat.Chat:input_type -> sortedchat.ChatRequest
	4,  // 5: sortedchat.SortedChat.GetHistory:input_type -> sortedchat.GetHistoryRequest
	7,  // 6: sortedchat.SortedChat.GetChatList:input_type -> sortedchat.GetChatListRequest
	0,  // 7: sortedchat.SortedChat.CreateChat:input_type -> sortedchat.CreateChatRequest
	11, // 8: sortedchat.SortedChat.ListModel:input_type -> sortedchat.ListModelsRequest
	13, // 9: sortedchat.SortedChat.SearchChat:input_type -> sortedchat.ChatSearchRequest
	3,  // 10: sortedchat.SortedChat.Chat:output_type -> sortedchat.ChatResponse
	5,  // 11: sortedchat.SortedChat.GetHistory:output_type -> sortedchat.GetHistoryResponse
	8,  // 12: sortedchat.SortedChat.GetChatList:output_type -> sortedchat.GetChatListResponse
	1,  // 13: sortedchat.SortedChat.CreateChat:output_type -> sortedchat.CreateChatResponse
	12, // 14: sortedchat.SortedChat.ListModel:output_type -> sortedchat.ListModelsResponse
	15, // 15: sortedchat.SortedChat.SearchChat:output_type -> sortedchat.ChatSearchResponse
	10, // [10:16] is the sub-list for method output_type
	4,  // [4:10] is the sub-list for method input_type
	4,  // [4:4] is the sub-list for extension type_name
	4,  // [4:4] is the sub-list for extension extendee
	0,  // [0:4] is the sub-list for field type_name
}

func init() { file_chatservice_proto_init() }
func file_chatservice_proto_init() {
	if File_chatservice_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_chatservice_proto_rawDesc), len(file_chatservice_proto_rawDesc)),
			NumEnums:      0,
			NumMessages:   16,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_chatservice_proto_goTypes,
		DependencyIndexes: file_chatservice_proto_depIdxs,
		MessageInfos:      file_chatservice_proto_msgTypes,
	}.Build()
	File_chatservice_proto = out.File
	file_chatservice_proto_goTypes = nil
	file_chatservice_proto_depIdxs = nil
}
