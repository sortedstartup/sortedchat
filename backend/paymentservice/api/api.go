package api

import (
	"fmt"
	"log/slog"

	"sortedstartup/paymentservice/dao"
	pb "sortedstartup/paymentservice/proto"
	"sortedstartup/paymentservice/service"

	"google.golang.org/grpc"
)

type PaymentServiceAPI struct {
	pb.UnimplementedPaymentServiceServer
	service *service.PaymentService
}

func NewPaymentServiceAPI(daoFactory dao.DAOFactory) *PaymentServiceAPI {

	service, err := service.NewPaymentService(daoFactory)
	if err != nil {
		slog.Error("paymentservice:api:NewPaymentServiceAPI", "error", err)
		return nil
	}

	s := &PaymentServiceAPI{
		service: service,
	}

	return s
}

func (s *PaymentServiceAPI) Infer(_ *pb.InferRequest, stream grpc.ServerStreamingServer[pb.InferResponse]) error {
	return s.service.Infer(stream.Context(), "dummy")
}

func (s *PaymentServiceAPI) Init(config *dao.Config) error {
	switch config.Database.Type {
	case dao.DatabaseTypeSQLite:
		slog.Info("PaymentService: Running SQLite migrations")
		if err := dao.MigrateSQLite(config.Database.SQLite.URL); err != nil {
			slog.Error("paymentservice:api:Init", "error", err)
			return fmt.Errorf("failed to migrate SQLite database: %w", err)
		}
		if err := dao.SeedSqlite(config.Database.SQLite.URL); err != nil {
			slog.Error("paymentservice:api:Init", "error", err)
			return fmt.Errorf("failed to seed SQLite database: %w", err)
		}
	case dao.DatabaseTypePostgres:
		slog.Info("PaymentService: Running PostgreSQL migrations")
		dsn := config.Database.Postgres.GetPostgresDSN()
		if err := dao.MigratePostgres(dsn); err != nil {
			slog.Error("paymentservice:api:Init", "error", err)
			return fmt.Errorf("failed to migrate PostgreSQL database: %w", err)
		}
		if err := dao.SeedPostgres(dsn); err != nil {
			slog.Error("paymentservice:api:Init", "error", err)
			return fmt.Errorf("failed to seed PostgreSQL database: %w", err)
		}
	default:
		slog.Error("paymentservice:api:Init", "error", fmt.Errorf("unsupported database type: %s", config.Database.Type))
		return fmt.Errorf("unsupported database type: %s", config.Database.Type)
	}

	return nil
}
