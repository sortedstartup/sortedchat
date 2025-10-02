package api

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"sortedstartup/common/auth"
	"sortedstartup/realtimeservice/dao"
	pb "sortedstartup/realtimeservice/proto"
	"sortedstartup/realtimeservice/service"
)

type RealtimeServiceAPI struct {
	pb.UnimplementedRealtimeServiceServer
	service *service.RealtimeService
}

var SQLITE_DB_URL = "db.sqlite"

func NewRealtimeServiceAPI(daoFactory dao.DAOFactory) *RealtimeServiceAPI {
	slog.Info("RealtimeService:NewRealtimeServiceAPI")

	r := &RealtimeServiceAPI{
		service: service.NewRealtimeService(daoFactory),
	}

	return r
}

func (s *RealtimeServiceAPI) Init(config *dao.Config) {
	switch config.Database.Type {
	case dao.DatabaseTypeSQLite:
		slog.Info("RealtimeService: Running SQLite migrations")
		if err := dao.MigrateSQLite(config.Database.SQLite.URL); err != nil {
			log.Fatalf("RealtimeService: Failed to migrate SQLite database: %v", err)
		}
		if err := dao.SeedSqlite(config.Database.SQLite.URL); err != nil {
			log.Fatalf("RealtimeService: Failed to seed SQLite database: %v", err)
		}
	case dao.DatabaseTypePostgres:
		slog.Info("RealtimeService: Running PostgreSQL migrations")
		dsn := config.Database.Postgres.GetPostgresDSN()
		if err := dao.MigratePostgres(dsn); err != nil {
			log.Fatalf("RealtimeService: Failed to migrate PostgreSQL database: %v", err)
		}
		if err := dao.SeedPostgres(dsn); err != nil {
			log.Fatalf("RealtimeService: Failed to seed PostgreSQL database: %v", err)
		}
	}
	s.service.Init(config)
}

func (s *RealtimeServiceAPI) Offer(ctx context.Context, req *pb.OfferRequest) (*pb.OfferResponse, error) {
	slog.Info("RealtimeService:api:Offer")
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("RealtimeService:api:Offer", "error", "failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("failed to get user ID from context")
	}

	offer, err := s.service.Offer(req.Offer, req.Model, userID)
	if err != nil {
		slog.Error("RealtimeService:api:Offer", "error", "failed to offer", "error", err)
		return nil, fmt.Errorf("failed to offer")
	}
	return &pb.OfferResponse{
		Offer: offer,
	}, nil
}

func (s *RealtimeServiceAPI) IceCandidate(ctx context.Context, req *pb.IceCandidateRequest) (*pb.IceCandidateResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		slog.Error("RealtimeService:api:IceCandidate", "error", "failed to get user ID from context", "error", err)
		return nil, fmt.Errorf("failed to get user ID from context")
	}
	message, err := s.service.IceCandidate(req.Candidate, userID)
	if err != nil {
		slog.Error("RealtimeService:api:IceCandidate", "error", "failed to add ICE candidate", "error", err)
		return nil, fmt.Errorf("failed to add ICE candidate")
	}
	return &pb.IceCandidateResponse{
		Message: message,
	}, nil
}
