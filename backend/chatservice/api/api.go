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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SettingServiceAPI struct {
	pb.UnimplementedSettingServiceServer
	service *service.SettingService
}

func NewSettingService(queue queue.Queue, daoFactory db.DAOFactory) *SettingServiceAPI {
	slog.Info("api:NewSettingService", "queue", queue, "daoFactory", daoFactory)
	settingService := service.NewSettingService(queue, daoFactory)
	if settingService == nil {
		slog.Error("api:NewSettingService", "error", "failed to create setting service")
		return nil
	}
	return &SettingServiceAPI{service: settingService}
}

func (s *SettingServiceAPI) Init() {
	slog.Info("api:Init", "settingService", s.service)
	s.service.Init()
}

func (s *SettingServiceAPI) GetSetting(ctx context.Context, req *pb.GetSettingRequest) (*pb.GetSettingResponse, error) {
	slog.Info("api:GetSetting", "settingService", s.service)
	settings, err := s.service.GetSetting(ctx)
	if err != nil {
		slog.Error("api:GetSetting", "failed to get settings", "error", err)
		return nil, err
	}

	return &pb.GetSettingResponse{
		Settings: settings,
	}, nil
}

func (s *SettingServiceAPI) SetSetting(ctx context.Context, req *pb.SetSettingRequest) (*pb.SetSettingResponse, error) {
	err := s.service.SetSetting(ctx, req.Settings)
	if err != nil {
		slog.Error("api:SetSetting", "error", "failed to set settings", "error", err)
		return nil, err
	}

	return &pb.SetSettingResponse{
		Message: "Setting Saved",
	}, nil
}

type ChatServiceAPI struct {
	pb.UnimplementedSortedChatServer
	service *service.ChatService
}

func NewChatService(mux *http.ServeMux, queue queue.Queue, settingsManager *settings.SettingsManager, daoFactory db.DAOFactory) *ChatServiceAPI {
	slog.Info("api:NewChatService", "settingsManager", settingsManager, "daoFactory", daoFactory)
	settingsManager.LoadSettingsFromDB()

	chatService, err := service.NewChatService(queue, settingsManager, daoFactory)
	if err != nil {
		slog.Error("api:NewChatService", "error", "failed to initialize ChatService", "error", err)
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
		slog.Error("api:Chat", "error", "failed to get user ID from context", "error", err)
		if st, ok := status.FromError(err); ok {
			return status.Errorf(st.Code(), "failed to get user ID")
		}
		return status.Errorf(codes.Unauthenticated, "failed to get user ID")
	}
	return s.service.Chat(stream.Context(), userID, req, func(response *pb.ChatResponse) error {
		return stream.Send(response)
	})
}

func (s *ChatServiceAPI) GenerateChatName(ctx context.Context, req *pb.GenerateChatNameRequest) (*pb.GenerateChatNameResponse, error) {
	slog.Info("api:GenerateChatName", "request", req)
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:GenerateChatName", "error", "failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
	}
	chatName, err := s.service.GenerateChatName(ctx, userID, req.GetChatId(), req.GetMessage(), req.GetModel())
	if err != nil {
		slog.Error("api:GenerateChatName", "error", "failed to generate chat name", "error", err)
		return nil, fmt.Errorf("failed to generate chat name")
	}

	return &pb.GenerateChatNameResponse{
		ChatName: chatName,
	}, nil
}

func (s *ChatServiceAPI) GetHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	slog.Info("api:GetHistory", "request", req)
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:GetHistory", "error", "failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
	}
	history, chatmetadata, err := s.service.GetHistory(ctx, userID, req.ChatId)
	if err != nil {
		slog.Error("api:GetHistory", "error", "failed to get history", "error", err)
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
		slog.Error("api:GetChatList", "error", "failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
		//error while processing request, please try again
	}
	chats, err := s.service.GetChatList(ctx, userID, req.GetProjectId(), req.GetSoftDeleted())
	if err != nil {
		slog.Error("api:GetHistory", "error", "failed to get chat list", "error", err)
		return nil, fmt.Errorf("failed to get chat list")
	}
	return &pb.GetChatListResponse{Chats: chats}, nil
}

func (s *ChatServiceAPI) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.CreateChatResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:CreateChat", "error", "failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
	}
	chatId, err := s.service.CreateChat(ctx, userID, req.Name, req.GetProjectId())
	if err != nil {
		slog.Error("api:CreateChat", "error", "failed to create chat", "error", err)
		return nil, fmt.Errorf("failed to create chat")
	}

	return &pb.CreateChatResponse{
		Message: "Chat created successfully",
		ChatId:  chatId,
	}, nil
}

func (s *ChatServiceAPI) ListModel(ctx context.Context, req *pb.ListModelsRequest) (*pb.ListModelsResponse, error) {
	models, err := s.service.ListModel(ctx)
	if err != nil {
		slog.Error("api:ListModel", "error", "failed to list models", "error", err)
		return nil, fmt.Errorf("failed to list models")
	}

	return &pb.ListModelsResponse{Models: models}, nil
}

func (s *ChatServiceAPI) SearchChat(ctx context.Context, req *pb.ChatSearchRequest) (*pb.ChatSearchResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:SearchChat", "error", "failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
	}
	results, err := s.service.SearchChat(ctx, userID, req.Query)
	if err != nil {
		slog.Error("api:SearchChat", "error", "failed to search chat", "error", err)
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
		slog.Error("api:CreateProject", "error", "failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
	}

	projectID, err := s.service.CreateProject(ctx, userID, req.Name, req.Description, req.AdditionalData)
	if err != nil {
		slog.Error("api:CreateProject", "error", "failed to create project", "error", err)
		return nil, fmt.Errorf("failed to create project")
	}

	return &pb.CreateProjectResponse{
		Message:   "Project created successfully",
		ProjectId: projectID,
	}, nil
}

func (s *ChatServiceAPI) GetProjects(ctx context.Context, req *pb.GetProjectsRequest) (*pb.GetProjectsResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:GetProjects", "error", "failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
	}

	projects, err := s.service.GetProjects(ctx, userID)
	if err != nil {
		slog.Error("api:GetProjects", "error", "failed to get projects", "error", err)
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
		slog.Error("api:ListDocuments", "error", "failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
	}
	docs, err := s.service.ListDocuments(ctx, userID, req.GetProjectId())
	if err != nil {
		slog.Error("api:ListDocuments", "error", "failed to fetch documents", "error", err)
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
		slog.Error("api:SubmitGenerateEmbeddingsJob", "error", "failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
	}

	err = s.service.SubmitGenerateEmbeddingsJob(ctx, userID, req.GetProjectId())
	if err != nil {
		slog.Error("api:SubmitGenerateEmbeddingsJob", "error", "failed to submit generate embeddings job", "error", err)
		return nil, fmt.Errorf("failed to submit generate embeddings job")
	}

	return &pb.GenerateEmbeddingResponse{
		Message: "Embedding job submitted successfully",
	}, nil
}

func (s *ChatServiceAPI) BranchAChat(ctx context.Context, req *pb.BranchAChatRequest) (*pb.BranchAChatResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:BranchAChat", "error", "failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
	}

	newChatId, err := s.service.BranchAChat(ctx, userID, req.SourceChatId, req.BranchFromMessageId, req.BranchName)
	if err != nil {
		slog.Error("api:BranchAChat", "error", "failed to branch a chat", "error", err)
		return &pb.BranchAChatResponse{
			Message: "failed to branch a chat",
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
		slog.Error("api:ListChatBranch", "error", "failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
	}

	branches, err := s.service.ListChatBranch(ctx, userID, req.GetChatId())
	if err != nil {
		slog.Error("api:ListChatBranch", "error", "failed to list chat branch", "error", err)
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
		slog.Error("api:GetRAGDocumentReference", "error", "failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
	}
	return s.service.GetRAGDocumentReference(ctx, userID, req)
}

func (s *ChatServiceAPI) DeleteDocument(ctx context.Context, req *pb.DeleteDocumentRequest) (*pb.DeleteDocumentResponse, error) {
	userId, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:DeleteDocument", "error", "failed to get user ID from context to delete document", "error", err)
		if st, ok := status.FromError(err); ok {
			return nil, status.Errorf(st.Code(), "failed to get user ID")
		}
		return nil, status.Errorf(codes.Unauthenticated, "failed to get user ID")
	}
	err = s.service.DeleteDocument(ctx, userId, req.GetProjectId(), req.GetDocId())
	if err != nil {
		slog.Error("api:DeleteDocument", "error", "failed to delete document", "error", err)
		return nil, fmt.Errorf("failed to delete document")
	}
	return &pb.DeleteDocumentResponse{Message: "Document deleted successfully"}, nil
}

func (s *ChatServiceAPI) DeleteChat(ctx context.Context, req *pb.DeleteChatRequest) (*pb.DeleteChatResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:DeleteChat", "error", "failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
	}
	err = s.service.DeleteChat(ctx, userID, req.GetChatId(), req.GetOperation())
	if err != nil {
		slog.Error("api:DeleteChat", "error", "failed to delete chat", "error", err)
		return nil, fmt.Errorf("failed to delete chat")
	}
	return &pb.DeleteChatResponse{Message: "Chat deleted successfully"}, nil
}

func (s *ChatServiceAPI) RestoreChat(ctx context.Context, req *pb.RestoreChatRequest) (*pb.RestoreChatResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:RestoreChat", "error", "failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
	}
	err = s.service.RestoreChat(ctx, userID, req.GetChatId())
	if err != nil {
		slog.Error("api:RestoreChat", "error", "failed to restore chat", "error", err)
		return nil, fmt.Errorf("failed to restore chat")
	}
	return &pb.RestoreChatResponse{Message: "Chat restored successfully"}, nil
}

func (s *ChatServiceAPI) RenameChat(ctx context.Context, req *pb.RenameChatRequest) (*pb.RenameChatResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("api:RenameChat", "error", "failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
	}
	err = s.service.RenameChat(ctx, userID, req.GetChatId(), req.GetName())
	if err != nil {
		slog.Error("api:RenameChat", "error", "failed to rename chat", "error", err)
		return nil, fmt.Errorf("failed to rename chat")
	}
	return &pb.RenameChatResponse{Message: "Chat renamed successfully"}, nil
}

func (s *ChatServiceAPI) Init(config *db.Config) {
	s.service.Init(config)
}
