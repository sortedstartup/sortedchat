package api

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	db "sortedstartup/chatservice/dao"
	pb "sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/queue"
	"sortedstartup/chatservice/service"
	settings "sortedstartup/chatservice/settings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SettingServiceAPI struct {
	pb.UnimplementedSettingServiceServer
	service *service.SettingService
}

func NewSettingService(queue queue.Queue, daoFactory db.DAOFactory) *SettingServiceAPI {
	settingService := service.NewSettingService(queue, daoFactory)
	return &SettingServiceAPI{service: settingService}
}

func (s *SettingServiceAPI) Init() {
	s.service.Init()
}

func (s *SettingServiceAPI) GetSetting(ctx context.Context, req *pb.GetSettingRequest) (*pb.GetSettingResponse, error) {
	settings, err := s.service.GetSetting(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.GetSettingResponse{
		Settings: settings,
	}, nil
}

func (s *SettingServiceAPI) SetSetting(ctx context.Context, req *pb.SetSettingRequest) (*pb.SetSettingResponse, error) {
	err := s.service.SetSetting(ctx, req.Settings)
	if err != nil {
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

const HARDCODED_USER_ID = "0"

func NewChatService(mux *http.ServeMux, queue queue.Queue, settingsManager *settings.SettingsManager, daoFactory db.DAOFactory) *ChatServiceAPI {
	settingsManager.LoadSettingsFromDB()

	chatService, err := service.NewChatService(queue, settingsManager, daoFactory)
	if err != nil {
		log.Fatalf("Failed to initialize ChatService: %v", err)
	}

	s := &ChatServiceAPI{
		service: chatService,
	}

	s.registerRoutes(mux)
	chatService.EmbeddingSubscriber()

	return s
}

func (s *ChatServiceAPI) Chat(req *pb.ChatRequest, stream grpc.ServerStreamingServer[pb.ChatResponse]) error {
	return s.service.Chat(stream.Context(), HARDCODED_USER_ID, req, func(response *pb.ChatResponse) error {
		return stream.Send(response)
	})
}

func (s *ChatServiceAPI) GenerateChatName(ctx context.Context, req *pb.GenerateChatNameRequest) (*pb.GenerateChatNameResponse, error) {
	chatName, err := s.service.GenerateChatName(ctx, HARDCODED_USER_ID, req.GetChatId(), req.GetMessage(), req.GetModel())
	if err != nil {
		return nil, err
	}

	return &pb.GenerateChatNameResponse{
		ChatName: chatName,
	}, nil
}

func (s *ChatServiceAPI) GetHistory(ctx context.Context, req *pb.GetHistoryRequest) (*pb.GetHistoryResponse, error) {
	history, err := s.service.GetHistory(ctx, HARDCODED_USER_ID, req.ChatId)
	if err != nil {
		return nil, err
	}

	return &pb.GetHistoryResponse{
		History: history,
	}, nil
}

func (s *ChatServiceAPI) GetChatList(ctx context.Context, req *pb.GetChatListRequest) (*pb.GetChatListResponse, error) {
	chats, err := s.service.GetChatList(ctx, HARDCODED_USER_ID, req.GetProjectId())
	if err != nil {
		return nil, err
	}
	return &pb.GetChatListResponse{Chats: chats}, nil
}

func (s *ChatServiceAPI) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.CreateChatResponse, error) {
	chatId, err := s.service.CreateChat(ctx, HARDCODED_USER_ID, req.Name, req.GetProjectId())
	if err != nil {
		return nil, err
	}

	return &pb.CreateChatResponse{
		Message: "Chat created successfully",
		ChatId:  chatId,
	}, nil
}

func (s *ChatServiceAPI) ListModel(ctx context.Context, req *pb.ListModelsRequest) (*pb.ListModelsResponse, error) {
	models, err := s.service.ListModel(ctx)
	if err != nil {
		return nil, err
	}

	return &pb.ListModelsResponse{Models: models}, nil
}

func (s *ChatServiceAPI) SearchChat(ctx context.Context, req *pb.ChatSearchRequest) (*pb.ChatSearchResponse, error) {
	results, err := s.service.SearchChat(ctx, HARDCODED_USER_ID, req.Query)
	if err != nil {
		return nil, err
	}

	return &pb.ChatSearchResponse{
		Query:   req.Query,
		Results: results,
	}, nil
}

func (s *ChatServiceAPI) CreateProject(ctx context.Context, req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	projectID, err := s.service.CreateProject(ctx, HARDCODED_USER_ID, req.Name, req.Description, req.AdditionalData)
	if err != nil {
		return nil, err
	}

	return &pb.CreateProjectResponse{
		Message:   "Project created successfully",
		ProjectId: projectID,
	}, nil
}

func (s *ChatServiceAPI) GetProjects(ctx context.Context, req *pb.GetProjectsRequest) (*pb.GetProjectsResponse, error) {
	projects, err := s.service.GetProjects(ctx, HARDCODED_USER_ID)
	if err != nil {
		return nil, err
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
	docs, err := s.service.ListDocuments(ctx, HARDCODED_USER_ID, req.GetProjectId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch documents: %v", err)
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
	err := s.service.SubmitGenerateEmbeddingsJob(ctx, HARDCODED_USER_ID, req.GetProjectId())
	if err != nil {
		return nil, err
	}

	return &pb.GenerateEmbeddingResponse{
		Message: "Embedding job submitted successfully",
	}, nil
}

func (s *ChatServiceAPI) BranchAChat(ctx context.Context, req *pb.BranchAChatRequest) (*pb.BranchAChatResponse, error) {
	newChatId, err := s.service.BranchAChat(ctx, HARDCODED_USER_ID, req.SourceChatId, req.BranchFromMessageId, req.BranchName)
	if err != nil {
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
	branches, err := s.service.ListChatBranch(ctx, HARDCODED_USER_ID, req.GetChatId())
	if err != nil {
		return nil, err
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

func (s *ChatServiceAPI) Init(config *db.Config) {
	switch config.Database.Type {
	case db.DatabaseTypeSQLite:
		slog.Info("Running SQLite migrations")
		if err := db.MigrateSQLite(config.Database.SQLite.URL); err != nil {
			log.Fatalf("Failed to migrate SQLite database: %v", err)
		}
		if err := db.SeedSqlite(config.Database.SQLite.URL); err != nil {
			log.Fatalf("Failed to seed SQLite database: %v", err)
		}
	case db.DatabaseTypePostgres:
		slog.Info("Running PostgreSQL migrations")
		dsn := config.Database.Postgres.GetPostgresDSN()
		if err := db.MigratePostgres(dsn); err != nil {
			log.Fatalf("Failed to migrate PostgreSQL database: %v", err)
		}
		if err := db.SeedPostgres(dsn); err != nil {
			log.Fatalf("Failed to seed PostgreSQL database: %v", err)
		}
	default:
		log.Fatalf("Unsupported database type: %s", config.Database.Type)
	}
}
