# SortedChat

- Chat with multiple cloud models (OpenAI, Gemini, Claude)
- Chat with local models (via Ollama / llama.cpp)
- Chat with docs — upload documents and chat in a project context (RAG)
- Agentic chat
- Branch a chat — fork any conversation and explore different directions
- Realtime audio chat — talk to the LLM in real time

---

## Two Ways to Run
- **Web App** — runs as a backend server; access via browser at `http://localhost:8080`
- **Desktop App** — runs as a native GUI application

---

## Run as Web App

**1. Download the latest binary from [GitHub Releases](https://github.com/sortedstartup/sortedchat/releases)**

Pick the binary for your platform (e.g. `sortedchat-server-linux-amd64`) and make it executable:

```bash
chmod +x ./sortedchat-server-linux-amd64
```

**2. Get Google OAuth credentials**

Go to [Google Cloud Console](https://console.cloud.google.com/) → APIs & Services → Credentials → Create OAuth Client ID (Web application).  
Set the authorized redirect URI to `http://localhost:8080/callback`.  
Copy your **Client ID** and **Client Secret**.

**3. Create a `ui-config.json`**

```json
{
  "API_URL": "http://localhost:8080",
  "API_UPLOAD_URL": "http://localhost:8080",
  "GOOGLE_CLIENT_ID": "<your-google-client-id>",
  "GOOGLE_OAUTH_URL": "https://accounts.google.com/o/oauth2/v2/auth",
  "GOOGLE_REDIRECT_URL": "http://localhost:8080/callback"
}
```

**4. Export your Google Client Secret in the same terminal**

```bash
export GOOGLE_CLIENT_SECRET=<your-google-client-secret>
```

**5. Run the binary**

```bash
./sortedchat-server-linux-amd64 --ui-config-path ./ui-config.json
```

Open `http://localhost:8080` in your browser.


## Run as Desktop App

1. Download the latest app binary from [GitHub Releases](https://github.com/sortedstartup/sortedchat/releases) (e.g. `sortedchat-app-linux-amd64`)

```bash
chmod +x ./sortedchat-app-linux-amd64
./sortedchat-app-linux-amd64
```

The app opens directly

---

## Run using Docker

You have to pass your own `ui-config.json` using a volume mount.

**1. Create your `ui-config.json`**:

```json
{
  "API_URL": "http://localhost:8080",
  "API_UPLOAD_URL": "http://localhost:8080",
  "GOOGLE_CLIENT_ID": "<your-google-client-id>",
  "GOOGLE_OAUTH_URL": "https://accounts.google.com/o/oauth2/v2/auth",
  "GOOGLE_REDIRECT_URL": "http://localhost:8080/callback"
}
```

**2. Run the container**:

```bash
docker run -d \
  --name sortedchat \
  -e GOOGLE_CLIENT_SECRET=<your-google-client-secret> \
  -v $(pwd)/ui-config.json:/config/ui-config.json \
  -p 8080:8080 \
  ghcr.io/sortedstartup/sortedchat:latest \
  --ui-config-path /config/ui-config.json
```

Open `http://localhost:8080` in your browser.

## Development Setup

#### 1. Clone Repository
```bash
git clone https://github.com/sortedstartup/sortedchat.git
cd sortedchat
```

#### 2. Run Frontend 
```bash
cd frontend
pnpm install
pnpm run dev
cd ..
```

#### 3. Run Backend
```bash
cd backend
CGO_CFLAGS="-I$(pwd)/sqlite3" go run -tags "sqlite_fts5" ./mono/
```

**Application will be available at:** http://localhost:5173

#### 4. Run Desktop App (Dev Mode)
To run the native desktop application with hot-reloading:

```bash
cd backend/mono
wails dev -tags "sqlite_fts5,dev,webkit2_41,wails" -s -v 2
```
