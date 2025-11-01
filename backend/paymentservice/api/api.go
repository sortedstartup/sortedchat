package api

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"sortedstartup/common/auth"
	"sortedstartup/paymentservice/dao"
	pb "sortedstartup/paymentservice/proto"
	"sortedstartup/paymentservice/service"

	"github.com/stripe/stripe-go/v83"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		slog.Error("paymentservice:api:CreateProduct", "error", err)
		return nil, err
	}

	if strings.TrimSpace(req.Name) == "" {
		return nil, status.Error(codes.InvalidArgument, "Product name cannot be empty")
	}
	if strings.TrimSpace(req.Description) == "" {
		return nil, status.Error(codes.InvalidArgument, "Product description cannot be empty")
	}
	if req.AmountInSmallestUnit <= 0 {
		return nil, status.Error(codes.InvalidArgument, "Invalid request, please try again with valid parameters")
	}

	// Convert Currency enum to string
	var currencyStr string
	switch req.Currency {
	case pb.Currency_USD:
		currencyStr = "USD"
	case pb.Currency_INR:
		currencyStr = "INR"
	default:
		return nil, status.Error(codes.InvalidArgument, "Unsupported currency type")
	}

	// Determine if payment is recurring
	isRecurring := req.PaymentType == pb.PaymentType_RECURRING

	// Convert interval enum to string for database storage
	var intervalPeriod string

	slog.Info("paymentservice:api:CreateProduct", "isRecurring", isRecurring, "interval", req.Interval, "intervalCount", req.IntervalCount, "intervalPeriod", intervalPeriod)
	if isRecurring {
		switch req.Interval {
		case pb.Interval_WEEK:
			intervalPeriod = "week"
		case pb.Interval_MONTH:
			intervalPeriod = "month"
		case pb.Interval_QUARTER:
			intervalPeriod = "quarter"
		case pb.Interval_YEAR:
			intervalPeriod = "year"
		default:
			return nil, status.Error(codes.InvalidArgument, "Invalid interval type")
		}

		// Validate interval count for recurring payments
		if req.IntervalCount <= 0 {
			return nil, status.Error(codes.InvalidArgument, "Interval count must be greater than 0 for recurring payments")
		}
	}

	id, err := s.service.CreateProduct(ctx, userID, req.Name, req.Description, req.AmountInSmallestUnit, currencyStr, isRecurring, req.IntervalCount, intervalPeriod)
	if err != nil {
		return nil, err
	}
	return &pb.CreateProductResponse{
		Id:      id,
		Message: "Product created successfully",
	}, nil
}

func (s *PaymentServiceAPI) ListProducts(ctx context.Context, req *pb.ListProductsRequest) (*pb.ListProductsResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("paymentservice:api:ListProducts", "error", err)
		return nil, err
	}

	daoProducts, err := s.service.ListProducts(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert DAO products to proto products
	products := make([]*pb.Product, len(daoProducts))
	for i, daoProduct := range daoProducts {
		var intervalCount int64
		if daoProduct.IntervalCount.Valid {
			intervalCount = daoProduct.IntervalCount.Int64
		}

		var intervalPeriod string
		if daoProduct.IntervalPeriod.Valid {
			intervalPeriod = daoProduct.IntervalPeriod.String
		}

		products[i] = &pb.Product{
			Id:                   daoProduct.ID,
			StripeProductId:      daoProduct.StripeProductID,
			RazorpayProductId:    daoProduct.RazorpayProductID,
			Name:                 daoProduct.Name,
			AmountInSmallestUnit: daoProduct.Price,
			Description:          daoProduct.Description,
			Currency:             daoProduct.GetCurrencyEnum(),
			IsRecurring:          daoProduct.IsRecurring,
			IntervalCount:        intervalCount,
			IntervalPeriod:       intervalPeriod,
			HasAccess:            daoProduct.HasAccess,
		}
	}

	return &pb.ListProductsResponse{
		Products: products,
	}, nil
}

func (s *PaymentServiceAPI) CreateStripeCheckoutSession(ctx context.Context, req *pb.CreateStripeCheckoutSessionRequest) (*pb.CreateStripeCheckoutSessionResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("paymentservice:api:CreateStripeCheckoutSession", "error", err)
		return nil, err
	}

	if strings.TrimSpace(req.ProductId) == "" {
		return nil, status.Error(codes.InvalidArgument, "Product ID cannot be empty")
	}

	SessionUrl, err := s.service.CreateStripeCheckoutSession(ctx, userID, req.ProductId)
	if err != nil {
		slog.Error("paymentservice:api:CreateStripeCheckoutSession", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create Stripe checkout session: %v", err)
	}
	return &pb.CreateStripeCheckoutSessionResponse{
		SessionUrl: SessionUrl,
	}, nil
}

func (s *PaymentServiceAPI) CreateRazorpayCheckoutSession(ctx context.Context, req *pb.CreateRazorpayCheckoutSessionRequest) (*pb.CreateRazorpayCheckoutSessionResponse, error) {

	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("paymentservice:api:CreateRazorpayCheckoutSession", "error", err)
		return nil, err
	}

	if strings.TrimSpace(req.ProductId) == "" {
		return nil, status.Error(codes.InvalidArgument, "product ID cannot be empty")
	}

	OrderId, Amount, Currency, err := s.service.CreateRazorpayCheckoutSession(ctx, userID, req.ProductId)
	if err != nil {
		slog.Error("paymentservice:api:CreateRazorpayCheckoutSession", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create Razorpay checkout session: %v", err)
	}
	return &pb.CreateRazorpayCheckoutSessionResponse{
		OrderId:  OrderId,
		Amount:   Amount,
		Currency: Currency,
	}, nil
}

func (s *PaymentServiceAPI) CreateStripeSubscriptionCheckoutSession(ctx context.Context, req *pb.CreateStripeSubscriptionCheckoutSessionRequest) (*pb.CreateStripeSubscriptionCheckoutSessionResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("paymentservice:api:CreateStripeSubscriptionCheckoutSession", "error", err)
		return nil, err
	}

	if strings.TrimSpace(req.ProductId) == "" {
		return nil, status.Error(codes.InvalidArgument, "Product ID cannot be empty")
	}

	sessionURL, err := s.service.CreateStripeSubscriptionCheckoutSession(ctx, userID, req.ProductId)
	if err != nil {
		slog.Error("paymentservice:api:CreateStripeSubscriptionCheckoutSession", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create Stripe subscription checkout session: %v", err)
	}

	return &pb.CreateStripeSubscriptionCheckoutSessionResponse{
		SessionUrl: sessionURL,
	}, nil
}

func (s *PaymentServiceAPI) CreateRazorpaySubscriptionCheckoutSession(ctx context.Context, req *pb.CreateRazorpaySubscriptionCheckoutSessionRequest) (*pb.CreateRazorpaySubscriptionCheckoutSessionResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("paymentservice:api:CreateRazorpaySubscriptionCheckoutSession", "error", err)
		return nil, err
	}

	if strings.TrimSpace(req.ProductId) == "" {
		return nil, status.Error(codes.InvalidArgument, "Product ID cannot be empty")
	}

	subscriptionID, amount, currency, err := s.service.CreateRazorpaySubscriptionCheckoutSession(ctx, userID, req.ProductId)
	if err != nil {
		slog.Error("paymentservice:api:CreateRazorpaySubscriptionCheckoutSession", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to create Razorpay subscription checkout session: %v", err)
	}

	return &pb.CreateRazorpaySubscriptionCheckoutSessionResponse{
		SubscriptionId: subscriptionID,
		Amount:         amount,
		Currency:       currency,
	}, nil
}

func (s *PaymentServiceAPI) Init(config *dao.Config) error {

	key := os.Getenv("STRIPE_SECRET_KEY")
	if key == "" {
		slog.Error("paymentservice:api:Init", "error", "STRIPE_SECRET_KEY is not set")
		return fmt.Errorf("STRIPE_SECRET_KEY is not set")
	}
	stripe.Key = key

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

func (s *PaymentServiceAPI) CheckUserProductAccess(ctx context.Context, req *pb.CheckUserProductAccessRequest) (*pb.CheckUserProductAccessResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("paymentservice:api:CheckUserProductAccess", "error", err)
		return nil, err
	}

	if strings.TrimSpace(req.ProductId) == "" {
		return nil, status.Error(codes.InvalidArgument, "Product ID cannot be empty")
	}

	hasAccess, err := s.service.CheckUserProductAccess(ctx, userID, req.ProductId)
	if err != nil {
		slog.Error("paymentservice:api:CheckUserProductAccess", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to check user product access: %v", err)
	}

	return &pb.CheckUserProductAccessResponse{
		HasAccess: hasAccess,
	}, nil
}
