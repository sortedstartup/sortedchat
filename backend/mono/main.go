package main

import (
	"bytes"
	"embed"
	"flag"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net"
	"net/http"
	"time"

	"sortedstartup/chat/mono/util"
	"sortedstartup/chatservice/api"
	"sortedstartup/chatservice/dao"
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

	// Load configuration
	config, err := dao.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	slog.Info("Application configuration loaded",
		"database_type", config.Database.Type,
		"postgres_host", config.Database.Postgres.Host,
		"postgres_port", config.Database.Postgres.Port,
		"sqlite_url", config.Database.SQLite.URL)

	// Create DAO factory
	daoFactory, err := dao.NewDAOFactory(config)
	if err != nil {
		log.Fatalf("Failed to create DAO factory: %v", err)
	}
	defer func() {
		if err := daoFactory.Close(); err != nil {
			log.Printf("Error closing DAO factory: %v", err)
		}
	}()

	queue := queue.NewInMemoryQueue()
	settingsManager := settings.NewSettingsManager(queue, daoFactory)

	chatServiceApi := api.NewChatService(mux, queue, settingsManager, daoFactory)
	chatServiceApi.Init(config)
	proto.RegisterSortedChatServer(grpcServer, chatServiceApi)

	settingServiceApi := api.NewSettingService(queue, daoFactory)
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

		// SPA fallback behavior: try to serve the requested file,
		// if it doesn't exist, serve index.html
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Try to open the requested file
		file, err := publicFS.Open(path[1:]) // Remove leading slash
		if err != nil {
			// File doesn't exist, serve index.html for SPA routing
			indexFile, indexErr := publicFS.Open("index.html")
			if indexErr != nil {
				http.Error(w, "index.html not found", http.StatusNotFound)
				return
			}
			defer indexFile.Close()

			// Get file info for modification time
			var modTime time.Time
			if stat, statErr := indexFile.Stat(); statErr == nil {
				modTime = stat.ModTime()
			} else {
				modTime = time.Now()
			}

			// Read the index.html content
			content, readErr := io.ReadAll(indexFile)
			if readErr != nil {
				http.Error(w, "failed to read index.html", http.StatusInternalServerError)
				return
			}

			// Set content type to HTML
			w.Header().Set("Content-Type", "text/html; charset=utf-8")

			// Serve index.html with proper HTTP caching support
			http.ServeContent(w, r, "index.html", modTime, bytes.NewReader(content))
			return
		}
		defer file.Close()

		// File exists, serve it normally
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
