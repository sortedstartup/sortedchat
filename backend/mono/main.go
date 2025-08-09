package main

import (
	"embed"
	"flag"
	"io/fs"
	"log"
	"net"
	"net/http"

	"sortedstartup/chat/mono/util"
	"sortedstartup/chatservice/api"
	"sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/queue"
	"sortedstartup/chatservice/settings"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	defaultGrpcPort = "8000"
	defaultHttpPort = "8080"
	defaultHost     = ""
)

//go:embed public
var staticUIFS embed.FS

func main() {
	// Parse command line flags
	serverOnly := flag.Bool("server", false, "Start only the server without Wails GUI")
	host := flag.String("host", defaultHost, "Host to bind the server to (default: all interfaces)")
	grpcPort := flag.String("grpc-port", defaultGrpcPort, "Port for gRPC server")
	httpPort := flag.String("http-port", defaultHttpPort, "Port for HTTP server")
	flag.Parse()

	// Build addresses
	grpcAddr := net.JoinHostPort(*host, *grpcPort)
	httpAddr := net.JoinHostPort(*host, *httpPort)

	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using system env")
	}

	listener, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", grpcAddr, err)
	}

	grpcServer := grpc.NewServer()
	mux := http.NewServeMux()

	queue := queue.NewInMemoryQueue()
	settingsManager := settings.NewSettingsManager(queue)

	chatServiceApi := api.NewChatService(mux, queue, settingsManager)
	chatServiceApi.Init()
	proto.RegisterSortedChatServer(grpcServer, chatServiceApi)

	settingServiceApi := api.NewSettingService(queue)
	settingServiceApi.Init()
	proto.RegisterSettingServiceServer(grpcServer, settingServiceApi)

	// Enable reflection, TODO: may be remove in production ?
	reflection.Register(grpcServer)

	// gRPC-Web wrapper
	wrappedGrpc := grpcweb.WrapServer(grpcServer)

	// serve static UI
	publicFS, err := fs.Sub(staticUIFS, "public")
	if err != nil {
		log.Fatalf("Failed to create sub FS: %v", err)
	}
	staticUI := http.FileServer(http.FS(publicFS))

	// HTTP router (fallback to static UI if not gRPC)
	httpHandler := func(w http.ResponseWriter, r *http.Request) {

		if wrappedGrpc.IsGrpcWebRequest(r) || wrappedGrpc.IsAcceptableGrpcCorsRequest(r) {
			util.EnableCORS(wrappedGrpc).ServeHTTP(w, r)
			return
		}
		staticUI.ServeHTTP(w, r)
	}
	mux.HandleFunc("/", httpHandler)

	// HTTP server with CORS
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: util.EnableCORS(mux),
	}

	// Run both servers in parallel
	serverErr := make(chan error)

	go func() {
		log.Printf("Starting gRPC server on %s", grpcAddr)
		serverErr <- grpcServer.Serve(listener)
	}()

	go func() {
		log.Printf("Starting HTTP server on %s", httpAddr)
		serverErr <- httpServer.ListenAndServe()
	}()

	// Start Wails GUI unless --server flag is specified
	if !*serverOnly {
		Wails(mux)
	} else {
		log.Println("Running in server-only mode")
		err := <-serverErr
		if err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}

	WaitForServerError(serverErr)
}
