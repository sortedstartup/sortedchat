package api

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"sortedstartup/common/auth"
	"sortedstartup/inferenceservice/dao"
	pb "sortedstartup/inferenceservice/proto"
	"sortedstartup/inferenceservice/service"
)

type InferenceServiceAPI struct {
	pb.UnimplementedInferenceServiceServer
	service *service.InferenceService
}

var SQLITE_DB_URL = "db.sqlite"

func NewInferenceServiceAPI(daoFactory dao.DAOFactory) *InferenceServiceAPI {
	slog.Info("inferenceservice:api:NewInferenceServiceAPI")

	s := &InferenceServiceAPI{
		service: service.NewInferenceService(daoFactory),
	}

	return s
}

func (s *InferenceServiceAPI) DownloadModel(ctx context.Context, req *pb.DownloadModelRequest) (*pb.DownloadModelResponse, error) {
	slog.Info("inferenceservice:api:DownloadModel")
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("inferenceservice:api:DownloadModel", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
	}
	err = s.service.DownloadModel(ctx, userID, req.GetModelName())
	if err != nil {
		slog.Error("inferenceservice:api:DownloadModel", "error", err)
		return nil, fmt.Errorf("failed to download model")
	}

	return &pb.DownloadModelResponse{
		Message: "Download started successfully",
	}, nil
}

func (s *InferenceServiceAPI) GetLLMModels(req *pb.GetLLMModelsRequest, stream pb.InferenceService_GetLLMModelsServer) error {
	slog.Info("inferenceservice:api:GetLLMModels")
	return s.service.GetLLMModels(stream.Context(), func(models []*dao.ModelMetadata) error {
		slog.Info("inferenceservice:api:GetLLMModels", "message", "Getting LLM models")
		// Convert DAO models to protobuf models
		pbModels := make([]*pb.Model, len(models))
		for i, model := range models {
			// Convert JSON progress string to DownloadProgress proto structure
			var progressProto *pb.DownloadProgress
			if model.Progress != "" {
				if progress, err := dao.FromJSON(model.Progress); err == nil {
					progressProto = &pb.DownloadProgress{
						FileSize: progress.FileSize,
						Status:   pb.DownloadStatus(progress.Status),
						Progress: int32(progress.Progress),
						Speed:    progress.Speed,
					}
				} else {
					slog.Error("inferenceservice:api:GetLLMModels", "error", "failed to convert progress to JSON", "error", err)
				}
			}
			if progressProto == nil {
				// Default progress for models without progress data
				progressProto = &pb.DownloadProgress{
					FileSize: 0,
					Status:   pb.DownloadStatus(model.Status),
					Progress: 0,
					Speed:    0,
				}
			} else {
				slog.Error("inferenceservice:api:GetLLMModels", "error", "progressProto is nil", "model", model)
			}

			// Convert filestore_id pointer to string
			filestoreID := ""
			if model.FileStoreID != nil {
				filestoreID = *model.FileStoreID
			}

			pbModels[i] = &pb.Model{
				Id:              model.ID,
				Name:            model.Name,
				Url:             model.URL,
				Provider:        model.Provider,
				InputTokenCost:  model.InputTokenCost,
				OutputTokenCost: model.OutputTokenCost,
				Progress:        progressProto,
				IsDownloaded:    model.IsDownloaded,
				IsDownloadable:  model.IsDownloadable,
				Status:          pb.DownloadStatus(model.Status),
				FilestoreId:     filestoreID,
			}
		}
		slog.Info("inferenceservice:api:GetLLMModels", "message", "Sending LLM models")

		// Send the response
		return stream.Send(&pb.GetLLMModelsResponse{
			Models: pbModels,
		})
	})
}

func (s *InferenceServiceAPI) CancelDownload(ctx context.Context, req *pb.CancelDownloadRequest) (*pb.CancelDownloadResponse, error) {
	slog.Info("inferenceservice:api:CancelDownload")
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("inferenceservice:api:CancelDownload", "error", "failed to get user ID", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
	}
	err = s.service.CancelDownload(ctx, userID, req.GetModelName())
	if err != nil {
		slog.Error("inferenceservice:api:CancelDownload", "error", "failed to cancel download", "error", err)
		return nil, fmt.Errorf("failed to cancel download")
	}

	return &pb.CancelDownloadResponse{
		Message: "Download cancelled successfully",
	}, nil
}

func (s *InferenceServiceAPI) DeleteModel(ctx context.Context, req *pb.DeleteModelRequest) (*pb.DeleteModelResponse, error) {
	slog.Info("inferenceservice:api:DeleteModel")
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("inferenceservice:api:DeleteModel", "error", "failed to get user ID", "error", err)
		return nil, fmt.Errorf("failed to get user ID")
	}
	err = s.service.DeleteModel(ctx, userID, req.GetModelName())
	if err != nil {
		slog.Error("inferenceservice:api:DeleteModel", "error", "failed to delete model", "error", err)
		return nil, fmt.Errorf("failed to delete model")
	}

	return &pb.DeleteModelResponse{
		Message: "Model deleted successfully",
	}, nil
}

func (s *InferenceServiceAPI) Init(config *dao.Config) {
	slog.Info("inferenceservice:api:Init")
	switch config.Database.Type {
	case dao.DatabaseTypeSQLite:
		slog.Info("InferenceService: Running SQLite migrations")
		if err := dao.MigrateSQLite(config.Database.SQLite.URL); err != nil {
			log.Fatalf("InferenceService: Failed to migrate SQLite database: %v", err)
		}
		if err := dao.SeedSqlite(config.Database.SQLite.URL); err != nil {
			log.Fatalf("InferenceService: Failed to seed SQLite database: %v", err)
		}
	case dao.DatabaseTypePostgres:
		slog.Info("InferenceService: Running PostgreSQL migrations")
		dsn := config.Database.Postgres.GetPostgresDSN()
		if err := dao.MigratePostgres(dsn); err != nil {
			log.Fatalf("InferenceService: Failed to migrate PostgreSQL database: %v", err)
		}
		if err := dao.SeedPostgres(dsn); err != nil {
			log.Fatalf("InferenceService: Failed to seed PostgreSQL database: %v", err)
		}
	default:
		log.Fatalf("InferenceService: Unsupported database type: %s", config.Database.Type)
	}

	err := s.service.Initialize()
	if err != nil {
		log.Fatalf("InferenceService: Failed to initialize: %v", err)
	}
}
