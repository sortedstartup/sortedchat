package api

import (
	pb "sortedstartup/inferenceservice/proto"
	"sortedstartup/inferenceservice/service"

	"google.golang.org/grpc"
)

type InferenceServiceAPI struct {
	pb.UnimplementedInferenceServiceServer
	service *service.InferenceService
}

var SQLITE_DB_URL = "db.sqlite"

func NewInferenceService() *InferenceServiceAPI {

	s := &InferenceServiceAPI{
		service: service.NewInferenceService(),
	}

	return s
}

func (s *InferenceServiceAPI) Infer(req *pb.InferRequest, stream grpc.ServerStreamingServer[pb.InferResponse]) error {
	return s.service.Infer(stream.Context())
}

func (s *InferenceServiceAPI) Init() {
	//db.InitDB()
	// TODO: handle migration for postgres also
	// db.MigrateSQLite(SQLITE_DB_URL)
	// db.SeedSqlite(SQLITE_DB_URL)
}
