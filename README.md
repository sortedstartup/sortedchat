# SortedChat

Sorted Chat is a UI to chat with multiple LLM models without being locked into one. It supports both cloud-based LLMs (OpenAI, Google Gemini, Anthropic Claude) and local models with RAG (Retrieval-Augmented Generation) capabilities.

## Quick Start Options

### Option 1: Download Pre-built Binaries (Recommended)

Download the latest binaries from [GitHub Releases](https://github.com/sortedstartup/sortedchat/releases)

**Choose ONE binary based on your preference:**
- **`sortedchat-app-*`**: Desktop application with native GUI *(Recommended)*
- **`sortedchat-server-*`**: Web application (access via browser at http://localhost:8080)

> **Note**: Both are complete, standalone applications. Choose desktop for native experience or server for browser-based usage.

**Platform support:**
- **Linux**: `amd64`, `arm64`
- **macOS**: `amd64` (Intel), `arm64` (Apple Silicon)  
- **Windows**: `amd64`

#### Quick Setup for Binaries:

**1. Install required system dependencies (Linux only):**
```bash
# Ubuntu/Debian
sudo apt-get update && sudo apt-get install pkg-config libopus-dev libopusfile-dev

# CentOS/RHEL/Fedora
sudo yum install pkgconfig opus-devel opusfile-devel
# or
sudo dnf install pkgconfig opus-devel opusfile-devel
```

**2. Download and make executable:**
```bash
# Example for Linux amd64

# Desktop app (standalone GUI application)
wget https://github.com/sortedstartup/sortedchat/releases/download/v0.0.6-rc1/sortedchat-app-linux-amd64
chmod +x sortedchat-app-linux-amd64

# OR Web server (access via browser at http://localhost:8080)
wget https://github.com/sortedstartup/sortedchat/releases/download/v0.0.6-rc1/sortedchat-server-linux-amd64
chmod +x sortedchat-server-linux-amd64
```

**3. Set minimum environment variables:**

> **Note**: These variables are needed for **BOTH** desktop and web app binaries.

```bash
# Backend configuration (needed by both binaries)
export OPENAI_API_KEY=your-openai-key

# Frontend configuration (embedded in both binaries) 
export VITE_API_URL=http://localhost:8080
export VITE_API_UPLOAD_URL=http://localhost:8080/upload

# Optional: For Google SSO (frontend only needs Client ID)
export VITE_GOOGLE_CLIENT_ID=your-client-id
export VITE_GOOGLE_REDIRECT_URL=http://localhost:8080/auth/callback
export VITE_GOOGLE_OAUTH_URL=https://accounts.google.com/o/oauth2/v2/auth

# Backend SSO configuration (only if using Google SSO)
export GOOGLE_CLIENT_ID=your-client-id
export GOOGLE_CLIENT_SECRET=your-client-secret
export GOOGLE_REDIRECT_URL=http://localhost:8080/auth/callback
export APP_JWT_SECRET=any-secret-string
export APP_ISSUER=http://localhost:8080
export OAUTH_ISSUER_URL=https://accounts.google.com
```

**4. Run the application:**
```bash
# Desktop app (includes both frontend + backend)
./sortedchat-app-linux-amd64

# OR web app (backend only, frontend served via browser)
./sortedchat-server-linux-amd64
# Then open http://localhost:8080 in browser
```

### Why Both Binaries Need Environment Variables:

| Binary | Contains | Needs Variables |
|--------|----------|----------------|
| **`sortedchat-app-*`** | Frontend + Backend | All variables (VITE_* + backend vars) |
| **`sortedchat-server-*`** | Backend + Static Frontend | All variables (serves frontend files) |

**Both binaries are complete applications** - they include the compiled frontend code and need the same configuration.

### Option 2: Build from Source

Follow the [Development Setup](#development-setup) section below.

## Prerequisites

### Required Dependencies

1. **Go 1.24.3+** (for building from source)
   ```bash
   # Install Go from https://golang.org/dl/
   go version  # Should be 1.24.3 or higher
   ```

2. **SQLite3** (for database functionality)
   ```bash
   # macOS
   brew install sqlite3
   
   # Ubuntu/Debian
   sudo apt-get update && sudo apt-get install sqlite3 libsqlite3-dev
   
   # CentOS/RHEL/Fedora
   sudo yum install sqlite sqlite-devel
   # or
   sudo dnf install sqlite sqlite-devel
   
   # Windows
   # Download from https://sqlite.org/download.html
   # Or use chocolatey: choco install sqlite
   ```

3. **Opus Audio Libraries** (for real-time voice features)
   ```bash
   # macOS
   brew install opus opusfile pkg-config
   
   # Ubuntu/Debian
   sudo apt-get update && sudo apt-get install pkg-config libopus-dev libopusfile-dev
   
   # CentOS/RHEL/Fedora
   sudo yum install pkgconfig opus-devel opusfile-devel
   # or
   sudo dnf install pkgconfig opus-devel opusfile-devel
   
   # Windows
   # Usually included in pre-built binaries
   # For development: Install via MSYS2 or vcpkg
   ```

4. **GTK and WebKit Libraries** (for desktop application development)
   ```bash
   # Ubuntu/Debian
   sudo apt-get update && sudo apt-get install libgtk-3-dev libwebkit2gtk-4.1-dev libsoup-3.0-dev
   
   # CentOS/RHEL/Fedora
   sudo yum install gtk3-devel webkit2gtk4.1-devel libsoup3-devel
   # or
   sudo dnf install gtk3-devel webkit2gtk4.1-devel libsoup3-devel
   
   # macOS (via Homebrew)
   brew install gtk+3 webkit2gtk
   ```

5. **Node.js 20+ and pnpm** (for frontend development only)
   ```bash
   # Install Node.js from https://nodejs.org/
   npm install -g pnpm
   ```

### Optional Dependencies

6. **Ollama** (for local embeddings and RAG functionality)
   ```bash
   # Install Ollama from https://ollama.ai/
   # macOS/Linux
   curl -fsSL https://ollama.ai/install.sh | sh
   
   # Start Ollama service
   ollama serve
   
   # Pull the embedding model (required for RAG)
   ollama pull nomic-embed-text
   ```

7. **Docker** (for PostgreSQL or full Docker deployment)
   ```bash
   # Install Docker from https://docker.com/
   docker --version
   ```

## Database Options

SortedChat supports **two database backends**:

### Option A: SQLite (Default - Recommended for Local Development)
- **Pros**: No setup required, embedded database, perfect for single-user
- **Cons**: No concurrent users, limited scalability
- **Extensions**: sqlite-vss (for vector search), FTS5 (for full-text search)

### Option B: PostgreSQL (Recommended for Production)
- **Pros**: Production-ready, concurrent users, better performance
- **Cons**: Requires PostgreSQL setup
- **Extensions**: pgvector (for vector search)

**Setup PostgreSQL:**
```bash
docker run -d \
  --name sortedchat_postgres_dev \
  -e POSTGRES_DB=sortedchat_dev \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=dev_password \
  -p 5432:5432 \
  --restart unless-stopped \
  pgvector/pgvector:pg15
```

**Configure environment variables for PostgreSQL:**
```bash
export DB_TYPE=postgres
export POSTGRES_HOST=localhost  
export POSTGRES_PASSWORD=dev_password
export POSTGRES_DATABASE=sortedchat_dev
export POSTGRES_PORT=5432
export POSTGRES_USERNAME=postgres
```

## Development Setup

### 1. Clone Repository
```bash
git clone https://github.com/sortedstartup/sortedchat.git
cd sortedchat
```

### 2. Build Frontend 
```bash
cd frontend
pnpm install
pnpm run build
cd ..
```

### 3. Run Backend
```bash
cd backend
CGO_CFLAGS="-I$(pwd)/sqlite3" go run -tags "sqlite_fts5" ./mono/
```

**Application will be available at:** http://localhost:8080

### 4. Run Desktop Application (Development)

> **Note**: Desktop application requires a GUI environment (X11 on Linux, native on macOS/Windows). For headless servers, containers, or WSL without GUI, use the web version (step 3) instead.

```bash
# Build frontend first (required for desktop app)
cd frontend
pnpm install
pnpm run build

# Create and copy frontend build to wails expected location
mkdir -p ../backend/frontend-build-wails/dist
cp -r dist/* ../backend/frontend-build-wails/dist/
cd ..

# Run desktop application (requires GUI environment)
cd backend
CGO_CFLAGS="-I$(pwd)/sqlite3" go run -tags "sqlite_fts5,dev,webkit2_41" mono/env_dev.go mono/main.go mono/wails.go
```

**If you get GTK/X11 errors**, use the web version instead:
```bash
cd backend
CGO_CFLAGS="-I$(pwd)/sqlite3" go run -tags "sqlite_fts5" ./mono/
# Then open http://localhost:8080 in your browser
```

## Configuration

### Environment Variables

**Database Configuration:**
```bash
# SQLite (default)
# No configuration needed

# PostgreSQL  
export DB_TYPE=postgres
export POSTGRES_HOST=localhost
export POSTGRES_PASSWORD=dev_password
export POSTGRES_DATABASE=sortedchat_dev
export POSTGRES_PORT=5432
export POSTGRES_USERNAME=postgres
```

**LLM API Configuration:**
```bash
# For cloud LLM APIs (configure in app settings page)
export OPENAI_API_KEY=your-openai-key
export GEMINI_API_KEY=your-gemini-key  
export ANTHROPIC_API_KEY=your-claude-key
```

**Authentication (Required for Google SSO):**

> ** First time?** [Get Google OAuth credentials](#google-oauth-setup) before setting these variables.

```bash
# Backend configuration
export GOOGLE_CLIENT_ID=your-app.apps.googleusercontent.com
export GOOGLE_CLIENT_SECRET=your-client-secret
export GOOGLE_REDIRECT_URL=http://localhost:8080/auth/callback
export APP_JWT_SECRET=your-jwt-secret-key
export APP_ISSUER=http://localhost:8080
export OAUTH_ISSUER_URL=https://accounts.google.com

# Frontend configuration (for pre-built binaries)
# Note: Frontend only needs Client ID, not Client Secret
export VITE_GOOGLE_CLIENT_ID=your-app.apps.googleusercontent.com
export VITE_GOOGLE_REDIRECT_URL=http://localhost:8080/auth/callback
export VITE_GOOGLE_OAUTH_URL=https://accounts.google.com/o/oauth2/v2/auth
export VITE_API_URL=http://localhost:8080
export VITE_API_UPLOAD_URL=http://localhost:8080/upload
```

### In-App Settings

**Access Settings Page:** Click the settings icon in the application

**Configure:**
1. **LLM API Keys**: OpenAI, Gemini, Claude API keys
2. **API URLs**: Custom endpoints (default: official APIs)
3. **Ollama URL**: Default `http://localhost:11434` (configure if running elsewhere)

## Google OAuth Setup

To enable Google SSO login, you need to create OAuth credentials:

### Quick Steps:
1. **Go to [Google Cloud Console](https://console.cloud.google.com/)**
2. **Create/select project** → Enable "Google+ API" 
3. **APIs & Services** → **Credentials** → **"+ CREATE CREDENTIALS"** → **"OAuth Client IDs"**
4. **Configure consent screen** (first time only):
   - App name: "SortedChat Local" 
   - Add your email as test user
5. **Create credentials**:
   - Type: **"Web application"**
   - Name: "SortedChat Local"
   - Authorized redirect URI: `http://localhost:8080/auth/callback`
   - Click on "Create"
6. **Copy Client ID and Secret** to your environment variables

### Alternative: Skip SSO Setup
```bash
# Use fake credentials to test without SSO
export GOOGLE_CLIENT_ID="fake_client_id"
export GOOGLE_CLIENT_SECRET="fake_client_secret"
export VITE_GOOGLE_CLIENT_ID="fake_client_id"
# App will work but SSO login will show "not configured"
```

## Ollama Setup (For RAG Features)

**Why Ollama?** Ollama provides local embedding models for RAG (document search) functionality. It's optional but recommended for privacy and offline document processing.

### 1. Install Ollama
```bash
# macOS/Linux
curl -fsSL https://ollama.ai/install.sh | sh

# Windows: Download from https://ollama.ai/
```

### 2. Start Ollama Service
```bash
ollama serve  # Runs on http://localhost:11434
```

### 3. Install Embedding Model
```bash
ollama pull nomic-embed-text  # Required for document embeddings
```

### 4. Configure in App
- **Settings Page** → **Ollama URL**: `http://localhost:11434`
- **For Desktop App**: Ollama must be running locally
- **For Docker**: Ollama is included in docker-compose setup

## Docker Deployment

For production deployment with all services:

```bash
cd deployment
cp litellm-config.yaml.example litellm-config.yaml
# Edit litellm-config.yaml with your API keys

docker-compose up -d
```

**Services included:**
- **SortedChat**: Main application (port 8080)
- **LiteLLM**: Multi-LLM proxy (internal)
- **Ollama**: Local embeddings (internal)

## Usage

### Desktop Application (Recommended)
1. Download and run `sortedchat-app-*` binary for your platform
2. Application opens with native GUI (no browser needed)
3. Create account or login with Google SSO
4. Configure API keys in Settings
5. Start chatting with multiple LLMs

### Web Application (Alternative)
1. Download and run `sortedchat-server-*` binary for your platform
2. Open http://localhost:8080 in your browser
3. Same functionality as desktop app

### Key Features
- **Multi-LLM Support**: Chat with GPT, Gemini, Claude in one interface
- **Local Models**: Download and run models locally using llama.cpp
- **RAG**: Upload documents for context-aware conversations
- **Project Management**: Organize chats and documents in projects
- **Usage Analytics**: Track token usage and costs

## Advanced Configuration

### Custom SQLite Path
```bash
export SQLITE_DB_PATH=/path/to/your/database.db
```

### PostgreSQL Manual Connection
```bash
psql -h localhost -p 5432 -U postgres -d sortedchat_dev
```

### Build Tags Explanation
- **`sqlite_fts5`**: Enables full-text search in SQLite
- **`wails`**: Enables desktop GUI functionality
- **`dev`**: Development mode with debugging
- **`webkit2_41`**: WebKit version for desktop app

## Testing Your Setup

### Quick Test (No Configuration)
1. **Start desktop app**: `./sortedchat-app-*` (works without any setup)
2. **Basic functionality**: Create chats, navigate UI (SSO will show "not configured")

### Full Setup Test
1. **Set minimum environment variables**:
   ```bash
   # Required for LLM chat
   export OPENAI_API_KEY=your-openai-key
   
   # Required for Google SSO (both backend + frontend)
   export GOOGLE_CLIENT_ID=your-app.apps.googleusercontent.com
   export GOOGLE_CLIENT_SECRET=your-client-secret
   export GOOGLE_REDIRECT_URL=http://localhost:8080/auth/callback
   export APP_JWT_SECRET=any-secret-string
   export APP_ISSUER=http://localhost:8080
   export OAUTH_ISSUER_URL=https://accounts.google.com
   
   # Frontend vars (for pre-built binaries)
   export VITE_GOOGLE_CLIENT_ID=your-app.apps.googleusercontent.com
   export VITE_GOOGLE_REDIRECT_URL=http://localhost:8080/auth/callback
   export VITE_GOOGLE_OAUTH_URL=https://accounts.google.com/o/oauth2/v2/auth
   export VITE_API_URL=http://localhost:8080
   export VITE_API_UPLOAD_URL=http://localhost:8080/upload
   ```

2. **Start application** with environment variables loaded
3. **Test Google SSO** login
4. **Upload a document** to test RAG (requires Ollama)
5. **Download a local model** to test local inference

## Troubleshooting

**Common Issues:**

1. **"sqlite3 not found"**
   - Install SQLite3 development headers
   - Ensure CGO_CFLAGS points to sqlite3 directory

2. **"libopusfile.so.0: cannot open shared object file"**
   - Install Opus audio libraries: `sudo apt-get install pkg-config libopus-dev libopusfile-dev`
   - On macOS: `brew install opus opusfile pkg-config`
   - Required for real-time voice features

3. **"Ollama connection failed"**
   - Check if Ollama is running: `curl http://localhost:11434`
   - Verify OLLAMA_URL in settings page

4. **"Build failed on Windows"**
   - Install MinGW-w64 or use WSL2
   - Ensure proper C compiler is available

5. **"API calls failing"**
   - Verify API keys in settings page
   - Check internet connectivity
   - Confirm API quotas/billing

6. **"Port 8080 already in use"**
   - **Desktop app**: No port needed, runs standalone
   - **Web app**: Stop other services or change port: `PORT=8081 ./sortedchat-server-*`
   - **Don't run both**: Choose either desktop OR web app, not both

7. **"Google SSO redirects to internal.sortedchat.com"**
   - **Root cause**: Missing frontend environment variables
   - **Solution**: Set `VITE_GOOGLE_OAUTH_URL=https://accounts.google.com/o/oauth2/v2/auth`
   - **Also set**: All `VITE_*` variables listed in Testing section above
   - **For desktop app**: Must set VITE_ vars before launching binary

8. **"Desktop app SSO keeps loading"**
   - Set `VITE_API_URL=http://localhost:8080` before starting desktop app
   - Ensure server binary is running on port 8080 
   - Check browser dev tools for CORS errors

9. **"Google OAuth not configured error"**
   - Set `VITE_GOOGLE_CLIENT_ID` environment variable
   - Must match your Google Cloud Console OAuth app client ID
   - Restart application after setting environment variables

**Get Help:**
- GitHub Issues: https://github.com/sortedstartup/sortedchat/issues
- Check logs in application console
- Enable debug mode for detailed logging


**Quick Commands Reference:**
```bash
# Web app (SQLite)
CGO_CFLAGS="-I$(pwd)/sqlite3" go run -tags "sqlite_fts5" ./mono/

# Desktop app  
CGO_CFLAGS="-I$(pwd)/../sqlite3" go run -tags "sqlite_fts5,dev,webkit2_41" main.go wails.go

# PostgreSQL setup
docker run -d --name sortedchat_postgres_dev -e POSTGRES_DB=sortedchat_dev -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=dev_password -p 5432:5432 pgvector/pgvector:pg15

# Ollama setup
ollama serve && ollama pull nomic-embed-text
```

