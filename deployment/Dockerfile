# Multi-stage build
FROM node:20 AS frontend-builder

# Install pnpm
RUN npm install -g pnpm

# Copy frontend source
COPY frontend/sortedchat /app/frontend
WORKDIR /app/frontend

# Install dependencies and build frontend
ENV CI=true
ENV npm_config_yes=true
ENV FORCE_COLOR=0
RUN pnpm install
RUN pnpm run build --mode standalone

# Go builder stage
FROM golang:1.24.3 AS go-builder

# Install build dependencies for CGO
# RUN apt-get update && apt-get install -y \
#     gcc \
#     g++ \
#     make \
#     pkg-config \
#     && rm -rf /var/lib/apt/lists/*

# Copy backend source
COPY backend /app/backend
COPY proto /app/proto

# Copy built frontend to backend public directory
COPY --from=frontend-builder /app/frontend/build/client /app/backend/mono/public

WORKDIR /app/backend

# Enable CGO (required for sqlite-vec-go-bindings and other CGO dependencies)
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64
ENV CGO_CFLAGS="-I/app/backend/sqlite3"

# Build the Go application with sqlite_fts5 build tag
RUN go build -tags "sqlite_fts5" -o sortedchat ./mono

# Final runtime stage
FROM ubuntu:24.04

# Install necessary runtime dependencies
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*

# Copy the built application
COPY --from=go-builder /app/backend/sortedchat /usr/local/bin/sortedchat

# Make it executable
RUN chmod +x /usr/local/bin/sortedchat

# Run the application
CMD ["/usr/local/bin/sortedchat"]