package service

import (
	"log/slog"
	"sortedstartup/realtimeservice/dao"
	pb "sortedstartup/realtimeservice/proto"
)

type RealtimeService struct {
	dao dao.Dao
}

func NewRealtimeService(dao dao.Dao) *RealtimeService {
	return &RealtimeService{dao: dao}
}

func (s *RealtimeService) Init(config *dao.Config) {
	slog.Info("RealtimeService: Init", "config", config)
}

func (s *RealtimeService) Offer(offer *pb.OfferRequest) (*pb.OfferResponse, error) {
	slog.Info("RealtimeService: Offer", "offer", offer)
	return &pb.OfferResponse{
		Offer: "OfferResponse",
	}, nil
}
