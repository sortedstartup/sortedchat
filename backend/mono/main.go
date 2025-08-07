package main

import (
	"embed"
	"io/fs"
	"log"
	"net"
	"net/http"

	"sortedstartup/chat/mono/util"
	"sortedstartup/chatservice/api"
	"sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/queue"
	settings "sortedstartup/chatservice/settings"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	listener, err := net.Listen("tcp", grpcPort)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	mux := http.NewServeMux()

	queue := queue.NewInMemoryQueue()
	settingsManager := settings.NewSettingsManager()

	chatServiceApi := api.NewChatService(mux, queue, settingsManager)
	chatServiceApi.Init()
	proto.RegisterSortedChatServer(grpcServer, chatServiceApi)

	settingServiceApi := api.NewSettingService(queue, settingsManager)
	settingServiceApi.Init()
	proto.RegisterSettingServiceServer(grpcServer, settingServiceApi)

	// Enable reflection, TODO: may be remove in production ?
	reflection.Register(grpcServer)

	// gRPC-Web wrapper
	wrappedGrpc := grpcweb.WrapServer(grpcServer)

	// serve static UI
	publicFS, err := fs.Sub(staticUIFS, "public")
	staticUI := http.FileServer(http.FS(publicFS))

	// HTTP router (fallback to static UI if not gRPC)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if wrappedGrpc.IsGrpcWebRequest(r) || wrappedGrpc.IsAcceptableGrpcCorsRequest(r) {
			util.EnableCORS(wrappedGrpc).ServeHTTP(w, r)
			wrappedGrpc.ServeHTTP(w, r)
			return
		}
		staticUI.ServeHTTP(w, r)
	})

	// HTTP server with CORS
	httpServer := &http.Server{
		Addr:    httpPort,
		Handler: util.EnableCORS(mux),
	}

	// Run both servers in parallel
	serverErr := make(chan error)

	go func() {
		log.Println("Starting gRPC server on", grpcPort)
		serverErr <- grpcServer.Serve(listener)
	}()

	go func() {
		log.Println("Starting HTTP server on", httpPort)
		serverErr <- httpServer.ListenAndServe()
	}()

	// Wait for either server to error
	err = <-serverErr
	if err != nil {
		log.Fatalf("Server error: %v", err)
	}

}
