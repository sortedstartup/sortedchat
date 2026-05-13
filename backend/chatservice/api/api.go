package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	db "sortedstartup/chatservice/dao"
	pb "sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/queue"
	"sortedstartup/chatservice/service"
	settings "sortedstartup/chatservice/settings"
	"sortedstartup/common/auth"

	"google.golang.org/grpc"
)

type SettingServiceAPI struct {
	pb.UnimplementedSettingServiceServer
	service *service.SettingService
}

func NewSettingService(queue queue.Queue, daoFactory db.DAOFactory) *SettingServiceAPI {
	slog.Debug("api:NewSettingService")
	settingService := service.NewSettingService(queue, daoFactory)
	if settingService == nil {
		slog.Error("api:NewSettingService", "error", "failed to create setting service")
		return nil
	}
	slog.Info("SettingServiceAPI initialized")
	return &SettingServiceAPI{service: settingService}
}

func (s *SettingServiceAPI) Init() {
	slog.Debug("api:Init", "settingService", s.service)
	s.service.Init()
}

func (s *SettingServiceAPI) GetSetting(ctx context.Context, req *pb.GetSettingRequest) (*pb.GetSettingResponse, error) {
	slog.Debug("api:GetSetting", "settingService", s.service, "name", req.Name)
	settings, err := s.service.GetSetting(ctx, req.Name)
	if err != nil {
		slog.Error("api:GetSetting", "failed to get settings", "error", err)
		return nil, err
	}

	return &pb.GetSettingResponse{
		Settings: settings,
	}, nil
}

func (s *SettingServiceAPI) SetSetting(ctx context.Context, req *pb.SetSettingRequest) (*pb.SetSettingResponse, error) {
	err := s.service.SetSetting(ctx, req.Name, req.Settings)
	if err != nil {
		slog.Error("api:SetSetting", "message", "failed to set settings", "error", err)
		return nil, err
	}

	return &pb.SetSettingResponse{
		Message: "Setting Saved",
	}, nil
}

func (s *SettingServiceAPI) GetProviderSetting(ctx context.Context, req *pb.GetProviderSettingRequest) (*pb.GetProviderSettingResponse, error) {
	settings, err := s.service.GetProviderSetting(ctx, req.Name)
	if err != nil {
		slog.Error("api:GetProviderSetting", "failed to get provider settings", "error", err)
		return nil, err
	}
	return &pb.GetProviderSettingResponse{Settings: settings}, nil
}

func (s *SettingServiceAPI) SetProviderSetting(ctx context.Context, req *pb.SetProviderSettingRequest) (*pb.SetProviderSettingResponse, error) {
	err := s.service.SetProviderSetting(ctx, req.Name, req.Settings, true)
	if err != nil {
		slog.Error("api:SetProviderSetting", "failed to set provider settings", "error", err)
		return nil, err
	}
	return &pb.SetProviderSettingResponse{Message: "Provider Setting Saved"}, nil
}

func (s *SettingServiceAPI) GetAllProviderSettings(ctx context.Context, req *pb.GetAllProviderSettingsRequest) (*pb.GetAllProviderSettingsResponse, error) {
	settings, err := s.service.GetAllProviderSettings(ctx)
	if err != nil {
		slog.Error("api:GetAllProviderSettings", "failed to get all provider settings", "error", err)
		return nil, err
	}
	return &pb.GetAllProviderSettingsResponse{Settings: settings}, nil
}

func (s *SettingServiceAPI) SetAllProviderSettings(ctx context.Context, req *pb.SetAllProviderSettingsRequest) (*pb.SetAllProviderSettingsResponse, error) {
	err := s.service.SetAllProviderSettings(ctx, req.Settings, true)
	if err != nil {
		slog.Error("api:SetAllProviderSettings", "failed to set all provider settings", "error", err)
		return nil, err
	}
	return &pb.SetAllProviderSettingsResponse{Message: "All Provider Settings Saved"}, nil
}

func (s *SettingServiceAPI) IsFirstBoot(ctx context.Context, req *pb.IsFirstBootRequest) (*pb.IsFirstBootResponse, error) {
	slog.Debug("api:IsFirstBoot")
	isFirstBoot, err := s.service.IsFirstBoot()
	if err != nil {
		slog.Error("api:IsFirstBoot", "message", "failed to check first boot", "error", err)
		return nil, err
	}

	return &pb.IsFirstBootResponse{
		IsFirstBoot: isFirstBoot,
	}, nil
}

func (s *SettingServiceAPI) TestConnection(ctx context.Context, req *pb.TestConnectionRequest) (*pb.TestConnectionResponse, error) {
	slog.Debug("api:TestConnection", "url", req.Url, "type", req.ConnectionType)

	response, err := s.service.TestConnection(ctx, req)
	if err != nil {
		slog.Error("api:TestConnection", "message", "failed to test connection", "error", err)
		return nil, err
	}

	return response, nil
}

type ChatServiceAPI struct {
	pb.UnimplementedSortedChatServer
	service *service.ChatService
}

func NewChatService(mux *http.ServeMux, queue queue.Queue, settingsManager *settings.SettingsManager, daoFactory db.DAOFactory) *ChatServiceAPI {
	slog.Debug("api:NewChatService")
	settingsManager.LoadSettingsFromDB()

	chatService, err := service.NewChatService(queue, settingsManager, daoFactory)
	if err != nil {
		slog.Error("api:NewChatService", "message", "failed to initialize ChatService", "error", err)
		return nil
	}

	s := &ChatServiceAPI{
		service: chatService,
	}

	s.registerRoutes(mux)
	chatService.EmbeddingSubscriber()
	return s
}

func (s *ChatServiceAPI) Chat(req *pb.ChatRequest, stream grpc.ServerStreamingServer[pb.ChatResponse]) error {
	userID, err := auth.GetUserIDFromContext_WithError(stream.Context())
	if err != nil {
		slog.Error("api:Chat", "message", "failed to get user ID from context", "error", err)
		return err
	}
	return s.service.Chat(stream.Context(), userID, req, func(response *pb.ChatResponse) error {
		return stream.Send(response)
	})
}

func (s *ChatServiceAPI) GenerateChatName(ctx context.Context, req *pb.GenerateChatNameRequest) (*pb.GenerateChatNameResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:GenerateChatName", "message", "failed to get user ID from context", "error", err)
		return nil, err
	}
	chatName, err := s.service.GenerateChatName(ctx, userID, req.GetChatId(), req.GetMessage(), req.GetModel(), req.GetProvider())
	if err != nil {
		slog.Error("api:GenerateChatName", "message", "failed to generate chat name", "error", err)
		return nil, fmt.Errorf("failed to generate chat name")
	}

	slog.Info("api:GenerateChatName", "chatName generated successfully", chatName)
	return &pb.GenerateChatNameResponse{
		ChatName: chatName,
	}, nil
}

func (s *ChatServiceAPI) GetHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:GetHistory", "message", "failed to get user ID from context", "error", err)
		return nil, err
	}
	history, chatmetadata, err := s.service.GetHistory(ctx, userID, req.ChatId)
	if err != nil {
		slog.Error("api:GetHistory", "message", "failed to get history", "error", err)
		return nil, fmt.Errorf("failed to get chat message history")
	}

	return &pb.GetHistoryResponse{
		History:      history,
		ChatMetadata: chatmetadata,
	}, nil
}

func (s *ChatServiceAPI) GetChatList(ctx context.Context, req *pb.GetChatListRequest) (*pb.GetChatListResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:GetChatList", "message", "failed to get user ID from context", "error", err)
		return nil, err
	}
	chats, err := s.service.GetChatList(ctx, userID, req.GetProjectId(), req.GetSoftDeleted())
	if err != nil {
		slog.Error("api:GetHistory", "message", "failed to get chat list", "error", err)
		return nil, fmt.Errorf("failed to get chat list")
	}
	return &pb.GetChatListResponse{Chats: chats}, nil
}

func (s *ChatServiceAPI) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.CreateChatResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:CreateChat", "message", "failed to get user ID from context", "error", err)
		return nil, err
	}
	chatId, err := s.service.CreateChat(ctx, userID, req.Name, req.GetProjectId())
	if err != nil {
		slog.Error("api:CreateChat", "message", "failed to create chat", "error", err)
		return nil, fmt.Errorf("failed to create chat")
	}

	return &pb.CreateChatResponse{
		Message: "Chat created successfully",
		ChatId:  chatId,
	}, nil
}

func (s *ChatServiceAPI) SearchChat(ctx context.Context, req *pb.ChatSearchRequest) (*pb.ChatSearchResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:SearchChat", "message", "failed to get user ID from context", "error", err)
		return nil, err
	}
	results, err := s.service.SearchChat(ctx, userID, req.Query)
	if err != nil {
		slog.Error("api:SearchChat", "message", "failed to search chat", "error", err)
		return nil, fmt.Errorf("failed to search chat")
	}

	return &pb.ChatSearchResponse{
		Query:   req.Query,
		Results: results,
	}, nil
}

func (s *ChatServiceAPI) CreateProject(ctx context.Context, req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:CreateProject", "message", "failed to get user ID from context", "error", err)
		return nil, err
	}

	projectID, err := s.service.CreateProject(ctx, userID, req.Name, req.Description, req.AdditionalData)
	if err != nil {
		slog.Error("api:CreateProject", "message", "failed to create project", "error", err)
		return nil, err
	}

	return &pb.CreateProjectResponse{
		Message:   "Project created successfully",
		ProjectId: projectID,
	}, nil
}

func (s *ChatServiceAPI) GetProjects(ctx context.Context, req *pb.GetProjectsRequest) (*pb.GetProjectsResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:GetProjects", "message", "failed to get user ID from context", "error", err)
		return nil, err
	}

	projects, err := s.service.GetProjects(ctx, userID)
	if err != nil {
		slog.Error("api:GetProjects", "message", "failed to get projects", "error", err)
		return nil, fmt.Errorf("failed to get projects")
	}

	var pbProjects []*pb.Project
	for _, p := range projects {
		pbProjects = append(pbProjects, &pb.Project{
			Id:             p.ID,
			Name:           p.Name,
			Description:    p.Description,
			AdditionalData: p.AdditionalData,
			CreatedAt:      p.CreatedAt,
			UpdatedAt:      p.UpdatedAt,
		})
	}

	return &pb.GetProjectsResponse{Projects: pbProjects}, nil
}

func (s *ChatServiceAPI) ListDocuments(ctx context.Context, req *pb.ListDocumentsRequest) (*pb.ListDocumentsResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:ListDocuments", "message", "failed to get user ID from context", "error", err)
		return nil, err
	}
	docs, err := s.service.ListDocuments(ctx, userID, req.GetProjectId())
	if err != nil {
		slog.Error("api:ListDocuments", "message", "failed to fetch documents", "error", err)
		return nil, fmt.Errorf("failed to fetch documents")
	}

	var result []*pb.Document
	for _, doc := range docs {
		result = append(result, &pb.Document{
			Id:              doc.ID,
			ProjectId:       doc.ProjectID,
			DocsId:          doc.DocsID,
			FileName:        doc.FileName,
			CreatedAt:       doc.CreatedAt,
			UpdatedAt:       doc.UpdatedAt,
			EmbeddingStatus: pb.Embedding_Status(doc.EmbeddingStatus),
		})
	}

	return &pb.ListDocumentsResponse{
		Documents: result,
	}, nil
}

func (s *ChatServiceAPI) SubmitGenerateEmbeddingsJob(ctx context.Context, req *pb.GenerateEmbeddingRequest) (*pb.GenerateEmbeddingResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:SubmitGenerateEmbeddingsJob", "message", "failed to get user ID from context", "error", err)
		return nil, err
	}

	err = s.service.SubmitGenerateEmbeddingsJob(ctx, userID, req.GetProjectId())
	if err != nil {
		slog.Error("api:SubmitGenerateEmbeddingsJob", "message", "failed to submit generate embeddings job", "error", err)
		return nil, fmt.Errorf("failed to submit generate embeddings job")
	}

	return &pb.GenerateEmbeddingResponse{
		Message: "Embedding job submitted successfully",
	}, nil
}

func (s *ChatServiceAPI) BranchAChat(ctx context.Context, req *pb.BranchAChatRequest) (*pb.BranchAChatResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:BranchAChat", "message", "failed to get user ID from context", "error", err)
		return nil, err
	}

	newChatId, err := s.service.BranchAChat(ctx, userID, req.SourceChatId, req.BranchFromMessageId, req.BranchName)
	if err != nil {
		slog.Error("api:BranchAChat", "message", "failed to branch a chat", "error", err)
		return &pb.BranchAChatResponse{
			Message: err.Error(),
		}, nil
	}

	return &pb.BranchAChatResponse{
		Message:   "Branch created successfully",
		NewChatId: newChatId,
	}, nil
}

func (s *ChatServiceAPI) ListChatBranch(ctx context.Context, req *pb.ListChatBranchRequest) (*pb.ListChatBranchResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:ListChatBranch", "message", "failed to get user ID from context", "error", err)
		return nil, err
	}

	branches, err := s.service.ListChatBranch(ctx, userID, req.GetChatId())
	if err != nil {
		slog.Error("api:ListChatBranch", "message", "failed to list chat branch", "error", err)
		return nil, fmt.Errorf("failed to list chat branch")
	}

	var pbChats []*pb.ChatInfo
	for i := range branches {
		pbChats = append(pbChats, &pb.ChatInfo{
			ChatId: branches[i].Id,
			Name:   branches[i].Name,
		})
	}

	return &pb.ListChatBranchResponse{
		BranchChatList: pbChats,
	}, nil
}

func (s *ChatServiceAPI) GetRAGDocumentReference(ctx context.Context, req *pb.RAGDocumentReferenceRequest) (*pb.RAGDocumentReferenceResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:GetRAGDocumentReference", "message", "failed to get user ID from context", "error", err)
		return nil, err
	}
	return s.service.GetRAGDocumentReference(ctx, userID, req)
}

func (s *ChatServiceAPI) DeleteDocument(ctx context.Context, req *pb.DeleteDocumentRequest) (*pb.DeleteDocumentResponse, error) {
	userId, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:DeleteDocument", "message", "failed to get user ID from context to delete document", "error", err)
		return nil, err
	}
	err = s.service.DeleteDocument(ctx, userId, req.GetProjectId(), req.GetDocId())
	if err != nil {
		slog.Error("api:DeleteDocument", "message", "failed to delete document", "error", err)
		return nil, fmt.Errorf("failed to delete document")
	}
	return &pb.DeleteDocumentResponse{Message: "Document deleted successfully"}, nil
}

func (s *ChatServiceAPI) DeleteChat(ctx context.Context, req *pb.DeleteChatRequest) (*pb.DeleteChatResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:DeleteChat", "message", "failed to get user ID from context", "error", err)
		return nil, err
	}
	err = s.service.DeleteChat(ctx, userID, req.GetChatId(), req.GetOperation())
	if err != nil {
		slog.Error("api:DeleteChat", "message", "failed to delete chat", "error", err)
		return nil, fmt.Errorf("failed to delete chat")
	}
	return &pb.DeleteChatResponse{Message: "Chat deleted successfully"}, nil
}

func (s *ChatServiceAPI) RestoreChat(ctx context.Context, req *pb.RestoreChatRequest) (*pb.RestoreChatResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:RestoreChat", "message", "failed to get user ID from context", "error", err)
		return nil, err
	}
	err = s.service.RestoreChat(ctx, userID, req.GetChatId())
	if err != nil {
		slog.Error("api:RestoreChat", "message", "failed to restore chat", "error", err)
		return nil, fmt.Errorf("failed to restore chat")
	}
	return &pb.RestoreChatResponse{Message: "Chat restored successfully"}, nil
}

func (s *ChatServiceAPI) RenameItem(ctx context.Context, req *pb.RenameItemRequest) (*pb.RenameItemResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:RenameItem", "message", "failed to get user ID from context", "error", err)
		return nil, err
	}
	msg, err := s.service.RenameItem(ctx, userID, req.GetItemId(), req.GetName(), req.GetItemType())
	if err != nil {
		slog.Error("api:RenameItem", "message", "failed to rename item", "error", err)
		return nil, err
	}

	return &pb.RenameItemResponse{Message: msg}, nil
}

func (s *ChatServiceAPI) ListModel(ctx context.Context, req *pb.ListModelsRequest) (*pb.ListModelsResponse, error) {
	models, err := s.service.ListModel(ctx)
	if err != nil {
		slog.Error("api:ListModel", "message", "failed to list models", "error", err)
		return nil, fmt.Errorf("failed to list models")
	}

	return &pb.ListModelsResponse{Models: models}, nil
}

func (s *ChatServiceAPI) AddModel(ctx context.Context, req *pb.AddModelRequest) (*pb.AddModelResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:AddModel", "message", "failed to get user ID from context", "error", err)
		return nil, err
	}

	if req.GetProviderName() == "" || req.GetModelId() == "" {
		slog.Error("api:AddModel", "message", "provider name, model ID  are required")
		return nil, fmt.Errorf("provider name, model ID  are required")
	}

	slog.Debug("api:AddModel", "provider", req.ProviderName, "modelID", req.ModelId, "userID", userID)
	msg, err := s.service.AddModel(ctx, req.GetModelId(), req.GetProviderName(), req.GetModelName(), req.GetInputTokenCost(), req.GetOutputTokenCost(), req.GetCachedTokenCost(), req.GetIsEmbeddingModel(), req.GetUrl())
	if err != nil {
		slog.Error("api:AddModel", "message", "failed to add model", "error", err)
		return nil, err
	}
	return &pb.AddModelResponse{Message: msg}, nil
}

func (s *ChatServiceAPI) Init(config *db.Config) {
	s.service.Init(config)
}
