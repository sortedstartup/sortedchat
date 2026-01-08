package main

import (
	"bytes"
	"context"
	"embed"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"sortedstartup/chat/mono/util"
	"sortedstartup/chatservice/api"
	"sortedstartup/chatservice/dao"
	"sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/queue"
	"sortedstartup/chatservice/settings"

	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"github.com/joho/godotenv"
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/log/global"
	otellog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure" // Needed for bufconn client
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/test/bufconn" // New Import for in-memory client

	authApi "sortedstartup/authservice/api"
	authDao "sortedstartup/authservice/dao"
	authService "sortedstartup/authservice/service"
	auth "sortedstartup/common/auth"
	panicInterceptor "sortedstartup/common/panic"

	inferenceApi "sortedstartup/inferenceservice/api"
	inferenceDao "sortedstartup/inferenceservice/dao"
	infereceProto "sortedstartup/inferenceservice/proto"

	realtimeApi "sortedstartup/realtimeservice/api"
	realtimeDao "sortedstartup/realtimeservice/dao"

	realtimeProto "sortedstartup/realtimeservice/proto"
)

const (
	defaultGrpcPort = "8000"
	defaultHttpPort = "8080"
	defaultHost     = ""
	bufSize         = 1024 * 1024 // Buffer size for bufconn
)

//go:embed public
var staticUIFS embed.FS

func newBufDialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
}

func main() {
	ctx := context.Background()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	// Parse command line flags
	serverOnly := flag.Bool("server", false, "Start only the server without Wails GUI")
	host := flag.String("host", defaultHost, "Host to bind the server to (default: all interfaces)")
	grpcPort := flag.String("grpc-port", defaultGrpcPort, "Port for gRPC server")
	httpPort := flag.String("http-port", defaultHttpPort, "Port for HTTP server")
	uiConfigPath := flag.String("ui-config-path", "", "Path to UI config file (default: embedded public/ui-config.json)")
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
	log.Printf("Main gRPC server listening on TCP: %s", grpcAddr)
	inProcessListener := bufconn.Listen(bufSize)
	inProcessDialer := newBufDialer(inProcessListener)
	log.Println("Created in-process bufconn listener for client communication")

	if os.Getenv("OTEL_EXPORTER_OTLP_HEADERS") != "" && os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" {
		res, err := newResource()
		if err != nil {
			log.Fatalf("Failed to create OTel resource: %v", err)
		}

		loggerProvider, err := newLoggerProvider(ctx, res)
		if err != nil {
			log.Fatalf("Failed to create OTel logger provider: %v", err)
		}
		defer func() {
			if err := loggerProvider.Shutdown(ctx); err != nil {
				fmt.Println("OTel logger shutdown error:", err)
			}
		}()
		global.SetLoggerProvider(loggerProvider)

		otelLogger := otelslog.NewLogger("my-app")
		slog.SetDefault(otelLogger)
	}

	// Adding Interceptors
	// Create JWT validator
	jwtSecret := os.Getenv("APP_JWT_SECRET")
	issuer := os.Getenv("APP_ISSUER")

	defaultJwtSecret, defaultIssuer := getJWTDefaults()
	if jwtSecret == "" {
		jwtSecret = defaultJwtSecret
	}

	if issuer == "" {
		issuer = defaultIssuer
	}

	validator := auth.NewJWTValidator([]byte(jwtSecret), issuer)

	// Create gRPC auth interceptor
	authInterceptor := auth.NewGRPCAuthInterceptor(validator, true) // requireAuth = true

	// Skip authentication for certain gRPC methods
	authInterceptor.SkipMethods([]string{
		"/grpc.health.v1.Health/Check",
	})

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			panicInterceptor.UnaryInterceptor(),
			authInterceptor.UnaryInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			panicInterceptor.StreamInterceptor(),
			authInterceptor.StreamInterceptor(),
		),
	)

	// Internal gRPC server for in-process calls
	// Creating a inmemory unauthentication server just for interservice communication when hosted on the SAME sever,
	// this does not work when services are on multiple server.
	// at that time we need to make a deision : either pass api level authentication (jwt) to the other service or use mtls or some other microservice communication auth
	internalGrpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(panicInterceptor.UnaryInterceptor()),
		grpc.ChainStreamInterceptor(panicInterceptor.StreamInterceptor()),
	)

	// Create HTTP auth middleware
	authMiddleware := auth.NewHTTPAuthMiddleware(validator, false) // requireAuth = false for flexibility

	// Skip authentication for certain paths
	authMiddleware.SkipPaths([]string{
		"/health",
		"/login",
		"/auth/callback",
		"/",
		"/index.html",
		"/webhook",
	})

	// Skip authentication for path prefixes
	authMiddleware.SkipPrefixes([]string{
		"/public/",
		"/auth/",
		"/static/",
		"/assets/",
	})

	mux := http.NewServeMux()

	// Load configuration
	config, err := dao.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
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
		slog.Error("Failed to create DAO factory", "error", err)
		log.Fatalf("Failed to create DAO factory: %v", err)
	}
	defer func() {
		if err := daoFactory.Close(); err != nil {
			log.Printf("Error closing DAO factory: %v", err)
		}
	}()

	//inference service
	inferenceConfig, err := inferenceDao.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	inferenceDaoFactory, err := inferenceDao.NewDAOFactory(inferenceConfig)
	if err != nil {
		log.Fatalf("Failed to create DAO factory: %v", err)
	}
	defer func() {
		if err := inferenceDaoFactory.Close(); err != nil {
			slog.Error("Error closing DAO factory", "error", err)
			log.Printf("Error closing DAO factory: %v", err)
		}
	}()

	//realtime service
	realtimeConfig, err := realtimeDao.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	realtimeDaoFactory, err := realtimeDao.NewDAOFactory(realtimeConfig)
	if err != nil {
		log.Fatalf("Failed to create DAO factory: %v", err)
	}
	defer func() {
		if err := realtimeDaoFactory.Close(); err != nil {
			slog.Error("Error closing DAO factory", "error", err)
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

	agentsServiceApi, err := api.NewAgentService(daoFactory, settingsManager)
	if err != nil {
		log.Fatalf("Failed to create agent service: %v", err)
	}
	// agentsServiceApi.Init()
	proto.RegisterSettingServiceServer(grpcServer, settingServiceApi)
	proto.RegisterSettingServiceServer(internalGrpcServer, settingServiceApi)
	proto.RegisterAgentServiceServer(grpcServer, agentsServiceApi)
	proto.RegisterAgentServiceServer(internalGrpcServer, agentsServiceApi)

	// Run both servers in parallel
	serverErr := make(chan error)

	go func() {
		log.Printf("Starting gRPC server on in-memory bufconn")
		serverErr <- internalGrpcServer.Serve(inProcessListener)
	}()

	settingsClientConn, err := grpc.Dial(
		"bufconn",
		grpc.WithContextDialer(inProcessDialer),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to create in-memory SettingsService client connection: %v", err)
	}
	defer settingsClientConn.Close()

	// Create SettingsService client for in-process calls
	settingsClient := proto.NewSettingServiceClient(settingsClientConn)

	inferenceServiceApi := inferenceApi.NewInferenceServiceAPI(inferenceDaoFactory, queue)
	inferenceServiceApi.Init(inferenceConfig)
	infereceProto.RegisterInferenceServiceServer(grpcServer, inferenceServiceApi)

	realtimeServiceApi := realtimeApi.NewRealtimeServiceAPI(realtimeDaoFactory, settingsClient)
	realtimeServiceApi.Init(realtimeConfig)
	realtimeProto.RegisterRealtimeServiceServer(grpcServer, realtimeServiceApi)

	authConfig, err := authDao.LoadConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		log.Fatalf("Failed to load configuration: %v", err)
	}

	authDaoFactory, err := authDao.NewDAOFactory(authConfig)
	if err != nil {
		slog.Error("Failed to create user service DAO", "error", err)
		log.Fatalf("Failed to create user service DAO: %v", err)
	}

	userServiceDao, err := authDaoFactory.CreateDAO()
	if err != nil {
		slog.Error("Failed to create user service DAO", "error", err)
		log.Fatalf("Failed to create user service DAO: %v", err)
	}

	userService := authService.NewUserService(userServiceDao)
	userService.Init(authConfig)
	authService := authService.NewAuthService(userService)
	authServiceApi := authApi.NewAuthServiceAPI(mux, authService)
	authServiceApi.Init()

	// Enable reflection, TODO: may be remove in production ?
	reflection.Register(grpcServer)

	// gRPC-Web wrapper
	wrappedGrpc := grpcweb.WrapServer(grpcServer)

	// serve static UI
	publicFS, err := fs.Sub(staticUIFS, "public")
	if err != nil {
		slog.Error("Failed to create sub FS", "error", err)
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
				slog.Error("index.html not found", "error", indexErr)
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
				slog.Error("failed to read index.html", "error", readErr)
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

	// Add health endpoint (public, no auth required)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// If uiConfigPath is set, serve that file, else serve embedded one
	mux.HandleFunc("/ui-config.json", func(w http.ResponseWriter, r *http.Request) {
		var configContent io.ReadCloser
		var err error

		if *uiConfigPath != "" {
			configContent, err = os.Open(*uiConfigPath)
			if err != nil {
				slog.Error("Failed to open config file from custom path", "file", *uiConfigPath, "error", err)
				http.Error(w, "Config file not found", http.StatusNotFound)
				return
			}
		} else {
			configContent, err = publicFS.Open("ui-config.json")
			if err != nil {
				slog.Error("Failed to open config file from embedded FS", "error", err)
				http.Error(w, "Config file not found", http.StatusNotFound)
				return
			}
		}
		defer configContent.Close()

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")

		_, err = io.Copy(w, configContent)
		if err != nil {
			slog.Error("Failed to serve config file", "error", err)
			http.Error(w, "Failed to serve config", http.StatusInternalServerError)
		}
	})

	mux.HandleFunc("/", httpHandler)

	wrappedHandler := util.EnableCORS(authMiddleware.Middleware(mux))

	// HTTP server with CORS and auth middleware
	httpServer := &http.Server{
		Addr:    httpAddr,
		Handler: wrappedHandler,
	}

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
		Wails(wrappedHandler)
	} else {
		log.Println("Running in server-only mode")
		err := <-serverErr
		if err != nil {
			slog.Error("Server error", "error", err)
			log.Fatalf("Server error: %v", err)
		}
	}

	WaitForServerError(serverErr)
}

func newResource() (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(semconv.SchemaURL,
			semconv.ServiceName("SortedChat"),
		),
	)
}

func newLoggerProvider(ctx context.Context, res *resource.Resource) (*otellog.LoggerProvider, error) {
	exporter, err := otlploghttp.New(ctx) //exporter
	if err != nil {
		return nil, err
	}
	processor := otellog.NewBatchProcessor(exporter)
	provider := otellog.NewLoggerProvider(
		otellog.WithResource(res),
		otellog.WithProcessor(processor),
	)
	return provider, nil
}
