package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"sortedstartup/chatservice/api"
	"sortedstartup/chatservice/dao"
	"sortedstartup/chatservice/store"

	"sortedstartup/chat/mono/util"

	"github.com/google/uuid"
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

	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		projectID := r.FormValue("project_id")
		if projectID == "" {
			http.Error(w, "Missing project_id", http.StatusBadRequest)
			return
		}

		// Get file from form
		file, header, err := r.FormFile("file")
		filename := header.Filename
		if err != nil {
			http.Error(w, "File not provided", http.StatusBadRequest)
			return
		}
		defer file.Close()

		objectID := uuid.New().String()

		err = store.StoreObject(r.Context(), objectID, file)
		err = db.FileSave(projectID, objectID, filename)
		if err != nil {
			http.Error(w, "Failed to save file: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message": "File uploaded successfully", "id": "%s"}`, objectID)
	})

	mux.HandleFunc("/documents/", func(w http.ResponseWriter, r *http.Request) {

		docsId := strings.TrimPrefix(r.URL.Path, "/documents/")
		if docsId == "" {
			http.Error(w, "Missing document ID", http.StatusBadRequest)
			return
		}

		doc, err := db.GetFileMetadata(docsId)
		if err != nil {
			fmt.Printf("Database error: %v\n", err)
			http.Error(w, "Document not found", http.StatusNotFound)
			return
		}

		filePath := filepath.Join("filestore", "objects", docsId)

		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("File does not exist at path: %s\n", filePath)
			http.Error(w, "File not found on disk", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, doc.FileName))

		http.ServeFile(w, r, filePath)
	})

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
