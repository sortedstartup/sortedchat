package main

import (
	"embed"
	"io/fs"
	"log"
	"net"
	"net/http"

	"sortedstartup/chatservice/api"
	"sortedstartup/chatservice/dao"
	"sortedstartup/chatservice/store"

	"sortedstartup/chat/mono/util"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "sortedstartup/chatservice/proto"
)

const (
	grpcPort = ":8000"
	httpPort = ":8080"
)

//go:embed public
var staticUIFS embed.FS

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system env")
	}

	lis, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	db, err := dao.NewSQLiteDAO("chatservice.db")
	if err != nil {
		log.Fatalf("Failed to initialize DAO: %v", err)
	}
	// defer db.Close()
	// db.InitDB()

	grpcServer := grpc.NewServer()
	apiServer := api.Server{}

	apiServer.Init()
	pb.RegisterSortedChatServer(grpcServer, &apiServer)

	// Enable reflection, TODO: may be remove in production ?
	reflection.Register(grpcServer)

	// grpc web server
	wrappedGrpc := grpcweb.WrapServer(grpcServer)
	mux := http.NewServeMux()

	// serve static UI
	publicFS, err := fs.Sub(staticUIFS, "public")
	staticUI := http.FileServer(http.FS(publicFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(r) || wrappedGrpc.IsAcceptableGrpcCorsRequest(r) {
			util.EnableCORS(wrappedGrpc).ServeHTTP(w, r)
			wrappedGrpc.ServeHTTP(w, r)
			return
		}
		// For non-matched requests, serve static UI as fallback
		staticUI.ServeHTTP(w, r)
	})

	store, err := store.NewDiskObjectStore("filestore")
	if err != nil {
		log.Fatalf("Failed to init object store: %v", err)
	}

	httpHandler := api.NewHTTPHandler(db, store)
	httpHandler.RegisterRoutes(mux)

	httpServer := &http.Server{
		Addr:    httpPort,
		Handler: util.EnableCORS(mux),
	}

	// Start servers both GRPC and GRPC Web in parallel
	serverErr := make(chan error)

	go func() {
		log.Println("Starting gRPC server", "addr", grpcPort)
		err = grpcServer.Serve(lis)
		if err != nil {
			serverErr <- err
		}
	}()

	go func() {
		log.Println("Starting gRPC web server", "addr", httpPort)
		err = httpServer.ListenAndServe()
		if err != nil {
			serverErr <- err
		}
	}()

	// Wait for either server to error
	err = <-serverErr
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}

}
