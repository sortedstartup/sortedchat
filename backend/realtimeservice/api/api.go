package api

import (
	"context"
	"fmt"
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

	return &RealtimeServiceAPI{
		service: service.NewRealtimeService(daoFactory),
	}
}

func (s *RealtimeServiceAPI) Init(config *dao.Config) {
	s.service.Init(config)
}

func (s *RealtimeServiceAPI) Offer(ctx context.Context, req *pb.OfferRequest) (*pb.OfferResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		return nil, err
	}

	offer, err := s.service.Offer(req, userID)
	if err != nil {
		return nil, err
	}
	return &pb.OfferResponse{
		Offer: offer,
	}, nil
}

func (s *RealtimeServiceAPI) IceCandidate(ctx context.Context, req *pb.IceCandidateRequest) (*pb.IceCandidateResponse, error) {
	userID, err := auth.GetUserIDFromContext_WithError(ctx)
	if err != nil {
		return nil, err
	}
	message, err := s.service.IceCandidate(req.Candidate, userID)
	if err != nil {
		fmt.Println("Error adding ICE candidate", err)
		return nil, err
	}
	return &pb.IceCandidateResponse{
		Message: message,
	}, nil
}
