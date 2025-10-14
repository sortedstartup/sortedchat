package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"sortedstartup/common/auth"
	"sortedstartup/paymentservice/dao"
	pb "sortedstartup/paymentservice/proto"
	"sortedstartup/paymentservice/service"

	"github.com/stripe/stripe-go/v83"
	"google.golang.org/grpc"
)

type PaymentServiceAPI struct {
	pb.UnimplementedPaymentServiceServer
	service *service.PaymentService
}

func NewPaymentServiceAPI(mux *http.ServeMux, daoFactory dao.DAOFactory) *PaymentServiceAPI {

	service, err := service.NewPaymentService(daoFactory)
	if err != nil {
		slog.Error("paymentservice:api:NewPaymentServiceAPI", "error", err)
		return nil
	}

	s := &PaymentServiceAPI{
		service: service,
	}

	s.registerRoutes(mux)

	return s
}

func (s *PaymentServiceAPI) Infer(_ *pb.InferRequest, stream grpc.ServerStreamingServer[pb.InferResponse]) error {
	return s.service.Infer(stream.Context(), "dummy")
}

func (s *PaymentServiceAPI) CreateProduct(ctx context.Context, req *pb.CreateProductRequest) (*pb.CreateProductResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("inferenceservice:api:DownloadModel", "error", err)
		return nil, err
	}
	id, err := s.service.CreateProduct(ctx, userID, req.Name, req.Description, req.Price, req.Currency)
	if err != nil {
		return nil, err
	}
	return &pb.CreateProductResponse{
		Id:      id,
		Message: "Product created successfully",
	}, nil
}

func (s *PaymentServiceAPI) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {

	daoProducts, err := s.service.ListProducts(ctx)
	if err != nil {
		return nil, err
	}

	// Convert DAO products to proto products
	products := make([]*pb.Product, len(daoProducts))
	for i, daoProduct := range daoProducts {
		products[i] = &pb.Product{
			Id:          daoProduct.ID,
			Name:        daoProduct.Name,
			Price:       daoProduct.Price,
			Description: daoProduct.Description,
			Currency:    daoProduct.Currency,
		}
	}

	return &pb.ListProductsResponse{
		Products: products,
	}, nil
}

func (s *PaymentServiceAPI) CreateCheckoutSession(ctx context.Context, req *pb.CreateCheckoutSessionRequest) (*pb.CreateCheckoutSessionResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("paymentservice:api:CreateCheckoutSession", "error", err)
		return nil, err
	}
	sessionID, err := s.service.CreateCheckoutSession(ctx, userID, req.ProductId)
	if err != nil {
		return nil, err
	}
	return &pb.CreateCheckoutSessionResponse{
		SessionId: sessionID,
	}, nil
}

func (s *PaymentServiceAPI) Init(config *dao.Config) error {

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

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
