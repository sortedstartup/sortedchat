package api

import (
	"context"
	"log"
	"log/slog"
	"sortedstartup/inferenceservice/dao"
	db "sortedstartup/inferenceservice/dao"
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
		return &pb.DownloadModelResponse{
			Message: "Failed to start download: " + err.Error(),
		}, err
	}

	message := "Download started successfully"
	return &pb.DownloadModelResponse{
		Message: message,
	}, nil
}

func (s *InferenceServiceAPI) Init(config *dao.Config) {
	switch config.Database.Type {
	case db.DatabaseTypeSQLite:
		slog.Info("InferenceService: Running SQLite migrations")
		if err := db.MigrateSQLite(config.Database.SQLite.URL); err != nil {
			log.Fatalf("InferenceService: Failed to migrate SQLite database: %v", err)
		}
		if err := db.SeedSqlite(config.Database.SQLite.URL); err != nil {
			log.Fatalf("InferenceService: Failed to seed SQLite database: %v", err)
		}
	case db.DatabaseTypePostgres:
		slog.Info("InferenceService: Running PostgreSQL migrations")
		dsn := config.Database.Postgres.GetPostgresDSN()
		if err := db.MigratePostgres(dsn); err != nil {
			log.Fatalf("InferenceService: Failed to migrate PostgreSQL database: %v", err)
		}
		if err := db.SeedPostgres(dsn); err != nil {
			log.Fatalf("InferenceService: Failed to seed PostgreSQL database: %v", err)
		}
	default:
		log.Fatalf("InferenceService: Unsupported database type: %s", config.Database.Type)
	}
}
