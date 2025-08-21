package api

import (
	"log"
	"log/slog"
	"sortedstartup/common/auth"
	"sortedstartup/inferenceservice/dao"
	db "sortedstartup/inferenceservice/dao"
	pb "sortedstartup/inferenceservice/proto"
	"sortedstartup/inferenceservice/service"

	"google.golang.org/grpc"
)

type InferenceServiceAPI struct {
	pb.UnimplementedInferenceServiceServer
	service *service.InferenceService
}

var SQLITE_DB_URL = "db.sqlite"

func NewInferenceServiceAPI(daoFactory dao.DAOFactory) *InferenceServiceAPI {

	s := &InferenceServiceAPI{
		service: service.NewInferenceService(daoFactory),
	}

	return s
}

func (s *InferenceServiceAPI) Infer(req *pb.InferRequest, stream grpc.ServerStreamingServer[pb.InferResponse]) error {
	userID, err := auth.GetUserIDFromContext_WithError(stream.Context())
	if err != nil {
		return err
	}
	return s.service.Infer(stream.Context(), userID)
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
