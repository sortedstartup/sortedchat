# SortedChat - System Architecture Diagram

## High-Level Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              CLIENT LAYER                                       │
├─────────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────┐                    ┌─────────────────────┐            │
│  │    Web Application  │                    │  Desktop Application│            │
│  │   (React + TS)      │                    │   (React + TS)      │            │
│  │   • Tailwind CSS    │                    │   • Tailwind CSS    │            │
│  │   • gRPC-Web        │                    │   • gRPC-Web        │            │
│  └─────────────────────┘                    └─────────────────────┘            │
└─────────────────────────────────────────────────────────────────────────────────┘
                                    │
                              ┌─────┴─────┐
                              │  HTTPS    │
                              │  gRPC-Web │
                              └─────┬─────┘
                                    │
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              API GATEWAY LAYER                                  │
├─────────────────────────────────────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────────────────────────────────────────┐ │
│  │                        Monolith (Go)                                        │ │
│  │  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐                │ │
│  │  │   gRPC-Web      │ │  Auth Middleware│ │   CORS Handler  │                │ │
│  │  │   Handler       │ │  (JWT + SSO)    │ │                 │                │ │
│  │  └─────────────────┘ └─────────────────┘ └─────────────────┘                │ │
│  │  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐                │ │
│  │  │  Static Files   │ │   Rate Limiting │ │   Health Check  │                │ │
│  │  │  (SPA Routing)  │ │                 │ │   /health       │                │ │
│  │  └─────────────────┘ └─────────────────┘ └─────────────────┘                │ │
│  └─────────────────────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
                                    │
                              ┌─────┴─────┐
                              │   gRPC    │
                              │ Internal  │
                              └─────┬─────┘
                                    │
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            MICROSERVICES LAYER                                  │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│ ┌─────────────────────┐ ┌─────────────────────┐ ┌─────────────────────────────┐ │
│ │   Auth Service      │ │   Chat Service      │ │   Inference Service         │ │
│ │                     │ │                     │ │                             │ │
│ │ • SSO Integration   │ │ • Chat Management   │ │ • Model Download            │ │
│ │   - Google OAuth    │ │ • Message History   │ │ • Local Model Management    │ │
│ │   - Apple (Future)  │ │ • RAG Pipeline      │ │ • llama.cpp Integration     │ │
│ │   - GitHub (Future) │ │ • Token Tracking    │ │ • Model Status Tracking     │ │
│ │ • JWT Management    │ │ • Project Management│ │ • File Storage              │ │
│ │ • User Sessions     │ │ • Search & Branch   │ │                             │ │
│ │                     │ │ • Usage Dashboard   │ │                             │ │
│ └─────────────────────┘ └─────────────────────┘ └─────────────────────────────┘ │
│           │                         │                           │               │
│           │                         │                           │               │
│           ▼                         ▼                           ▼               │
│ ┌─────────────────────┐ ┌─────────────────────┐ ┌─────────────────────────────┐ │
│ │   Auth DAO          │ │   Chat DAO          │ │   Inference DAO             │ │
│ │                     │ │                     │ │                             │ │
│ │ • User Management   │ │ • Message Storage   │ │ • Model Metadata            │ │
│ │ • Session Storage   │ │ • Chat Metadata     │ │ • Download Progress         │ │
│ │                     │ │ • RAG Chunks        │ │ • File Paths                │ │
│ │                     │ │ • Vector Embeddings │ │                             │ │
│ │                     │ │ • Cost Tracking     │ │                             │ │
│ └─────────────────────┘ └─────────────────────┘ └─────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────────────────────────┐
│                            EXTERNAL SERVICES LAYER                              │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│ ┌─────────────────────┐ ┌─────────────────────┐ ┌─────────────────────────────┐ │
│ │    LiteLLM Proxy    │ │       Ollama        │ │    Local File System        │ │
│ │                     │ │                     │ │                             │ │
│ │ • Multi-LLM Gateway │ │ • Embedding Models  │ │ • Document Storage          │ │
│ │   - OpenAI GPT      │ │   - nomic-embed     │ │   /filestore/objects/       │ │
│ │   - Google Gemini   │ │ • Local Inference   │ │ • Downloaded Models         │ │
│ │   - Anthropic Claude│ │ • Vector Generation │ │   /filestore/models/        │ │
│ │ • API Key Management│ │                     │ │ • Static Assets             │ │
│ │ • Rate Limiting     │ │                     │ │                             │ │
│ │ • Cost Tracking     │ │                     │ │                             │ │
│ └─────────────────────┘ └─────────────────────┘ └─────────────────────────────┘ │
│           │                         │                           │               │
│           │                         │                           │               │
│           ▼                         ▼                           ▼               │
│ ┌─────────────────────┐ ┌─────────────────────┐ ┌─────────────────────────────┐ │
│ │   Cloud LLM APIs    │ │  Local Model Store  │ │    Volume Mounts            │ │
│ │                     │ │                     │ │                             │ │
│ │ • OpenAI API        │ │ • Downloaded Models │ │ • Persistent Storage        │ │
│ │ • Google AI API     │ │ • llama.cpp Runtime │ │ • Docker Volumes            │ │
│ │ • Anthropic API     │ │ • Model Binaries    │ │                             │ │
│ └─────────────────────┘ └─────────────────────┘ └─────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────────────────────┘
                                    │
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              DATABASE LAYER                                     │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│          ┌─────────────────────────────────────────────────────────┐            │
│          │                 Database Engine                         │            │
│          │                                                         │            │
│          │  ┌─────────────────────┐ ┌─────────────────────────────┐ │            │
│          │  │    PostgreSQL       │ │         SQLite              │ │            │
│          │  │                     │ │                             │ │            │
│          │  │ • pgvector Extension│ │ • sqlite-vss Extension      │ │            │
│          │  │ • Vector Similarity │ │ • Vector Similarity         │ │            │
│          │  │ • HNSW Indexing     │ │ • Vector Search             │ │            │
│          │  │ • Production Ready  │ │ • Development/Embedded      │ │            │
│          │  └─────────────────────┘ └─────────────────────────────┘ │            │
│          └─────────────────────────────────────────────────────────┘            │
│                                    │                                             │
│          ┌─────────────────────────────────────────────────────────┐            │
│          │                    Data Schema                          │            │
│          │                                                         │            │
│          │ ┌─────────────────┐ ┌─────────────────┐ ┌─────────────┐ │            │
│          │ │   Chat Data     │ │    RAG Data     │ │ System Data │ │            │
│          │ │                 │ │                 │ │             │ │            │
│          │ │ • chat_messages │ │ • rag_chunks    │ │ • users     │ │            │
│          │ │ • chat_list     │ │ • embeddings    │ │ • sessions  │ │            │
│          │ │ • projects      │ │ • documents     │ │ • settings  │ │            │
│          │ │ • models        │ │ • project_docs  │ │ • models    │ │            │
│          │ │ • cost tracking │ │ • vector search │ │             │ │            │
│          │ └─────────────────┘ └─────────────────┘ └─────────────┘ │            │
│          └─────────────────────────────────────────────────────────┘            │
└─────────────────────────────────────────────────────────────────────────────────┘
```

## Detailed Component Breakdown

### 1. Client Layer
- **Web Application**: React + TypeScript + Tailwind CSS
- **Desktop Application**: Same React codebase, no Electron framework
- **Communication**: gRPC-Web for real-time streaming and efficient data transfer

### 2. API Gateway (Monolith)
- **HTTP Server**: Handles both gRPC-Web and static file serving
- **Authentication**: JWT + OAuth2 (Google SSO, future: Apple, GitHub)
- **SPA Routing**: Fallback to index.html for React routing
- **CORS**: Cross-origin support for web clients

### 3. Microservices
#### Auth Service
- User management and authentication
- SSO integration (Google OAuth)
- JWT token generation and validation
- Session management

#### Chat Service
- Chat and message management
- RAG pipeline integration
- Token usage tracking
- Project and document management
- Usage dashboard APIs

#### Inference Service
- Local model download and management
- llama.cpp integration for local inference
- Model status and progress tracking
- File storage management

### 4. External Services
#### LiteLLM Proxy
- Unified API for multiple LLM providers
- API key management
- Rate limiting and cost tracking
- OpenAI, Google Gemini, Anthropic Claude support

#### Ollama
- Local embedding model serving
- Vector generation for RAG
- nomic-embed-text model for embeddings

### 5. Database Layer
#### PostgreSQL (Production)
- pgvector extension for vector operations
- HNSW indexing for similarity search
- Full ACID compliance

#### SQLite (Development)
- sqlite-vss extension for vector operations
- Embedded database for easy development
- Same schema as PostgreSQL

## Data Flow Diagrams

### Chat Flow with External LLMs
```
User Input → Frontend → API Gateway → Chat Service → LiteLLM → Cloud APIs
    ↑                                       ↓
    └── Streaming Response ←─────────────────┘
                ↓
         Token Tracking → Database
```

### RAG Document Processing Flow
```
Document Upload → File Storage → RAG Pipeline → Ollama Embedding → Vector DB
                                      ↓
                              Chunk Processing → Database Storage
```

### Local Model Inference Flow
```
Model Download → Inference Service → llama.cpp → Local Response
       ↓                                            ↑
File Storage ← Progress Tracking ←──────────────────┘
```

## Technology Stack Summary

| Layer | Technology | Purpose |
|-------|------------|---------|
| **Frontend** | React + TypeScript + Tailwind | Web & Desktop UI |
| **API** | Go + gRPC + gRPC-Web | Backend services |
| **Authentication** | JWT + OAuth2 (Google) | User management |
| **LLM Gateway** | LiteLLM | Multi-provider LLM access |
| **Local Inference** | llama.cpp | Local model execution |
| **Embeddings** | Ollama + nomic-embed | RAG vector generation |
| **Database** | PostgreSQL/SQLite + Vector Extensions | Data persistence |
| **Deployment** | Docker + Docker Compose | Containerization |
| **File Storage** | Local File System | Document & model storage |

## Key Features Supported

✅ **Multi-Platform**: Web and Desktop applications  
✅ **Multi-LLM**: GPT, Gemini, Claude via LiteLLM  
✅ **Local Models**: Download and run models locally  
✅ **RAG**: Document upload and vector search  
✅ **Real-time**: Streaming responses via gRPC  
✅ **Authentication**: SSO with Google (extensible)  
✅ **Cost Tracking**: Token usage and cost analytics  
✅ **Project Management**: Organize chats and documents  
✅ **Flexible Database**: SQLite or PostgreSQL support

## Deployment Architecture

```
Docker Compose Network (llm_network)
├── sortedchat:8080     (Main Application)
├── litellm:4000        (LLM Proxy - Internal)
└── ollama:11434        (Embeddings - Internal)

External Dependencies:
├── OpenAI API          (via LiteLLM)
├── Google AI API       (via LiteLLM)  
├── Anthropic API       (via LiteLLM)
└── Google OAuth        (SSO)
```

This architecture provides a scalable, modular design that supports both cloud and local AI models while maintaining clean separation of concerns across different services. 