package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sortedstartup/chatservice/queue"
	"sortedstartup/common/auth"
	"sortedstartup/inferenceservice/dao"
	"sortedstartup/inferenceservice/llama"
	pb "sortedstartup/inferenceservice/proto"
	"sortedstartup/inferenceservice/service"
)

type InferenceServiceAPI struct {
	pb.UnimplementedInferenceServiceServer
	service *service.InferenceService
}

var SQLITE_DB_URL = "db.sqlite"

func NewInferenceServiceAPI(daoFactory dao.DAOFactory, queue queue.Queue) *InferenceServiceAPI {
	slog.Debug("inferenceservice:api:NewInferenceServiceAPI")

	s := &InferenceServiceAPI{
		service: service.NewInferenceService(daoFactory, queue),
	}

	slog.Info("InferenceServiceAPI initialized")
	return s
}

func (s *InferenceServiceAPI) DownloadModel(ctx context.Context, req *pb.DownloadModelRequest) (*pb.DownloadModelResponse, error) {
	slog.Debug("inferenceservice:api:DownloadModel")
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("inferenceservice:api:DownloadModel", "error", err)
		return nil, err
	}
	err = s.service.DownloadModel(ctx, userID, req.GetModelName())
	if err != nil {
		slog.Error("inferenceservice:api:DownloadModel", "error", err)
		return nil, fmt.Errorf("failed to download model")
	}

	slog.Info("inferenceservice:api:DownloadModel", "message", "Download started successfully")
	return &pb.DownloadModelResponse{
		Message: "Download started successfully",
	}, nil
}

func (s *InferenceServiceAPI) GetLLMModels(req *pb.GetLLMModelsRequest, stream pb.InferenceService_GetLLMModelsServer) error {
	slog.Info("inferenceservice:api:GetLLMModels", "message", "Getting LLM models")
	return s.service.GetLLMModels(stream.Context(), func(models []*dao.ModelMetadata) error {
		slog.Info("inferenceservice:api:GetLLMModels", "message", "Getting LLM models", "models", models)
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
					slog.Error("inferenceservice:api:GetLLMModels", "message", "failed to convert progress to JSON", "error", err)
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

			// Parse model_info JSON to proto structure
			var modelInfoProto *pb.ModelInfo
			if model.ModelInfo != "" && model.ModelInfo != "{}" {
				var modelInfo map[string]string
				if err := json.Unmarshal([]byte(model.ModelInfo), &modelInfo); err == nil {
					modelInfoProto = &pb.ModelInfo{
						HomePageUrl:  modelInfo["homepage_url"],
						Quantization: modelInfo["quantization"],
						DownloadSize: modelInfo["download_size"],
					}
				} else {
					slog.Error("inferenceservice:api:GetLLMModels", "message", "failed to parse model_info JSON", "error", err)
				}
			}

			pbModels[i] = &pb.Model{
				Id:               model.ID,
				Name:             model.Name,
				Url:              model.URL,
				Provider:         model.Provider,
				InputTokenCost:   model.InputTokenCost,
				OutputTokenCost:  model.OutputTokenCost,
				Progress:         progressProto,
				IsDownloaded:     model.IsDownloaded,
				IsDownloadable:   model.IsDownloadable,
				Status:           pb.DownloadStatus(model.Status),
				FilestoreId:      filestoreID,
				IsEmbeddingModel: model.IsEmbeddingModel,
				IsEnabled:        model.IsEnabled,
				ModelInfo:        modelInfoProto,
				CreatorName:      model.CreatorName,
				ModifiedBy:       model.ModifiedBy,
				Description:      model.Description,
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
	slog.Info("inferenceservice:api:CancelDownload", "modelName", req.GetModelName())
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("inferenceservice:api:CancelDownload", "message", "failed to get user ID", "error", err)
		return nil, err
	}
	err = s.service.CancelDownload(ctx, userID, req.GetModelName())
	if err != nil {
		slog.Error("inferenceservice:api:CancelDownload", "message", "failed to cancel download", "error", err)
		return nil, fmt.Errorf("failed to cancel download")
	}

	slog.Info("inferenceservice:api:CancelDownload", "message", "Download cancelled successfully")
	return &pb.CancelDownloadResponse{
		Message: "Download cancelled successfully",
	}, nil
}

func (s *InferenceServiceAPI) DeleteModel(ctx context.Context, req *pb.DeleteModelRequest) (*pb.DeleteModelResponse, error) {
	slog.Info("inferenceservice:api:DeleteModel", "modelName", req.GetModelName())
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("inferenceservice:api:DeleteModel", "message", "failed to get user ID", "error", err)
		return nil, err
	}
	err = s.service.DeleteModel(ctx, userID, req.GetModelName())
	if err != nil {
		slog.Error("inferenceservice:api:DeleteModel", "message", "failed to delete model", "error", err)
		return nil, fmt.Errorf("failed to delete model")
	}

	slog.Info("inferenceservice:api:DeleteModel", "message", "Model deleted successfully")
	return &pb.DeleteModelResponse{
		Message: "Model deleted successfully",
	}, nil
}

func isLamaServerDownloaded() bool {
	filePath := filepath.Join(llama.LLAMASERVER_DIR, "llama_version")
	info, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}
	if err != nil {
		slog.Error("inferenceservice:api:isLamaServerDownloaded", "message", "failed to stat llama_version file", "error", err)
		return false
	}

	return info.Size() > 0
}

func markLLamaServerDownloaded(sversion string) error {
	filePath := filepath.Join(llama.LLAMASERVER_DIR, "llama_version")
	err := os.MkdirAll(llama.LLAMASERVER_DIR, 0755)
	if err != nil {
		slog.Error("inferenceservice:api:markLLamaServerDownloaded", "message", "failed to create llamaserver directory", "error", err)
		return err
	}

	err = os.WriteFile(filePath, []byte(sversion), 0644)
	if err != nil {
		slog.Error("inferenceservice:api:markLLamaServerDownloaded", "message", "failed to write llama_version file", "error", err)
		return err
	}
	return nil
}

func (s *InferenceServiceAPI) Init(config *dao.Config) {
	ctx := context.Background()
	slog.Info("inferenceservice:api:Init", "message", "Downloading LlamaServer")
	if isDownloaded := isLamaServerDownloaded(); !isDownloaded {
		progress, err := llama.DownloadLlamaServer(ctx, llama.LLAMASERVER_DIR)
		if err != nil {
			slog.Error("inferenceservice:api:Init", "message", "failed to download LlamaServer", "error", err)
			return
		}
		// need to have store when llama-sever is successfully downloaded
		// cant use database because it will be shared by all instances
		slog.Info("inferenceservice:api:Init", "message", "LlamaServer downloaded successfully", "progress", progress)
		go func() {
			for p := range progress {
				switch p.Status {
				case llama.StatusDownloading:

				case llama.StatusCompleted:
					slog.Info("Downloaded LlamaServer", "progress", p)
					markLLamaServerDownloaded("not-implemented")
				default:
					slog.Info("Default")

				}
			}
		}()

	} else if isDownloaded {
		slog.Info("inferenceservice:api:Init", "message", "LlamaServer is already downloaded")
	}

	slog.Debug("inferenceservice:api:Init")
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

	slog.Info("inferenceservice:api:Init", "message", "Initializing InferenceService")
	err := s.service.Initialize()
	if err != nil {
		log.Fatalf("InferenceService: Failed to initialize: %v", err)
	}
}
