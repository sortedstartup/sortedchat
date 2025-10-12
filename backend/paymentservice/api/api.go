package api

import (
	"log"
	"log/slog"
	"sortedstartup/paymentservice/dao"
	db "sortedstartup/paymentservice/dao"
	pb "sortedstartup/paymentservice/proto"
	"sortedstartup/paymentservice/service"

	"google.golang.org/grpc"
)

type PaymentServiceAPI struct {
	pb.UnimplementedPaymentServiceServer
	service *service.PaymentService
}

var SQLITE_DB_URL = "db.sqlite"

func NewPaymentServiceAPI(daoFactory dao.DAOFactory) *PaymentServiceAPI {

	s := &PaymentServiceAPI{
		service: service.NewPaymentService(daoFactory),
	}

	return s
}

func (s *PaymentServiceAPI) Infer(req *pb.InferRequest, stream grpc.ServerStreamingServer[pb.InferResponse]) error {
	return s.service.Infer(stream.Context(), "dummy")
}

func (s *PaymentServiceAPI) Init(config *dao.Config) {
	switch config.Database.Type {
	case db.DatabaseTypeSQLite:
		slog.Info("PaymentService: Running SQLite migrations")
		if err := db.MigrateSQLite(config.Database.SQLite.URL); err != nil {
			log.Fatalf("PaymentService: Failed to migrate SQLite database: %v", err)
		}
		if err := db.SeedSqlite(config.Database.SQLite.URL); err != nil {
			log.Fatalf("PaymentService: Failed to seed SQLite database: %v", err)
		}
	case db.DatabaseTypePostgres:
		slog.Info("PaymentService: Running PostgreSQL migrations")
		dsn := config.Database.Postgres.GetPostgresDSN()
		if err := db.MigratePostgres(dsn); err != nil {
			log.Fatalf("PaymentService: Failed to migrate PostgreSQL database: %v", err)
		}
		if err := db.SeedPostgres(dsn); err != nil {
			log.Fatalf("PaymentService: Failed to seed PostgreSQL database: %v", err)
		}
	default:
		log.Fatalf("PaymentService: Unsupported database type: %s", config.Database.Type)
	}
}
