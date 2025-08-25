package api

import (
	"context"
	"log"
	"log/slog"
	"sortedstartup/inferenceservice/dao"
	pb "sortedstartup/inferenceservice/proto"
	"sortedstartup/inferenceservice/service"
)

type InferenceServiceAPI struct {
	pb.UnimplementedInferenceServiceServer
	service *service.InferenceService
}

var SQLITE_DB_URL = "db.sqlite"

const HARDCODED_USER_ID = "0"

func NewInferenceServiceAPI(daoFactory dao.DAOFactory) *InferenceServiceAPI {

	s := &InferenceServiceAPI{
		service: service.NewInferenceService(daoFactory),
	}

	return s
}

func (s *InferenceServiceAPI) DownloadModel(ctx context.Context, req *pb.DownloadModelRequest) (*pb.DownloadModelResponse, error) {
	err := s.service.DownloadModel(ctx, HARDCODED_USER_ID, req.GetModelName())
	if err != nil {
		return nil, err
	}

	return &pb.DownloadModelResponse{
		Message: "Download started successfully",
	}, nil
}

func (s *InferenceServiceAPI) GetLLMModels(req *pb.GetLLMModelsRequest, stream pb.InferenceService_GetLLMModelsServer) error {
	return s.service.GetLLMModels(stream.Context(), func(models []*dao.ModelMetadata) error {
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

		// Send the response
		return stream.Send(&pb.GetLLMModelsResponse{
			Models: pbModels,
		})
	})
}

func (s *InferenceServiceAPI) CancelDownload(ctx context.Context, req *pb.CancelDownloadRequest) (*pb.CancelDownloadResponse, error) {
	err := s.service.CancelDownload(ctx, HARDCODED_USER_ID, req.GetModelName())
	if err != nil {
		return nil, err
	}

	return &pb.CancelDownloadResponse{
		Message: "Download cancelled successfully",
	}, nil
}

func (s *InferenceServiceAPI) Init(config *dao.Config) {
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
}
