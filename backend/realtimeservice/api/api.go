package api

import (
	"context"
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
	return &pb.OfferResponse{
		Offer: "Offer",
	}, nil
}
