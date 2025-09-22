# James join our meeting
1. We will implement a meeting API 
2. To which  Let's say three users connected, using web rtc. 
3. We converted each user voice into text and send to llm
4. James will clearly know who is talking what because we pass each user voice => text share with james
5. We want to James to listen to us, and clearly undertand who is talking what. 
6. Then we want James to share screen to us and explain what he understand

# realtime voice
1. We already explored webrtc and implemented
2. OpenAI and Gemini realtim voice LLM models support webrtc
3. We did experiment using webrtc, sortedchat API will take voice input and give it to LLM and LLM voice response return back to webrtc.

# James first feature
1. As we know real time voice chat feature with LLM, we want to implement a agent who join the meeting and listen to multiple people conversation in meeting and answer the question
2. Also james(agent) should be able to share screen and explain what he understand and answer the question

# local LLM
1. We also learned how to run local LLM models using lama.cpp 
2. We can run realtime voice support models also locally, this will be helpful for us to run James locally without any internet connection.

# Design: James Meeting Assistant

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                      SortedChat + James                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐   │
│  │   Web Client 1  │ │   Web Client 2  │ │   Web Client 3  │   │
│  │   (User Alice)  │ │   (User Bob)    │ │   (User Carol)  │   │
│  │                 │ │                 │ │                 │   │
│  │ • Audio Stream  │ │ • Audio Stream  │ │ • Audio Stream  │   │
│  │ • Screen Share  │ │ • Screen Share  │ │ • Screen Share  │   │
│  │ • Chat UI       │ │ • Chat UI       │ │ • Chat UI       │   │
│  └─────────────────┘ └─────────────────┘ └─────────────────┘   │
│           │                   │                   │             │
│           └───────────────────┼───────────────────┘             │
│                               │                                 │
│  ┌─────────────────────────────┼─────────────────────────────┐   │
│  │          James Agent        │                             │   │
│  │                             │                             │   │
│  │  • Virtual Participant     │                             │   │
│  │  • Audio Output Stream     │                             │   │
│  │  • Screen Sharing          │                             │   │
│  │  • Meeting Understanding   │                             │   │
│  └─────────────────────────────┼─────────────────────────────┘   │
│                               │                                 │
└───────────────────────────────┼─────────────────────────────────┘
                                │
┌───────────────────────────────┼─────────────────────────────────┐
│              Backend Services │                                 │
├───────────────────────────────┼─────────────────────────────────┤
│                               ▼                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │               Meeting Service                           │   │
│  │                                                         │   │
│  │ • WebRTC Signaling                                      │   │
│  │ • Meeting Room Management                               │   │
│  │ • Multi-user Audio Coordination                        │   │
│  │ • James Integration                                     │   │
│  └─────────────────────────────────────────────────────────┘   │
│                               │                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │               James Service                             │   │
│  │                                                         │   │
│  │ • Speech-to-Text (Multi-speaker)                       │   │
│  │ • Conversation Context Management                      │   │
│  │ • LLM Integration (Local/Cloud)                        │   │
│  │ • Text-to-Speech                                       │   │
│  │ • Screen Generation & Sharing                          │   │
│  └─────────────────────────────────────────────────────────┘   │
│                               │                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │            Existing Services                            │   │
│  │                                                         │   │
│  │ • ChatService (Text conversations)                     │   │
│  │ • RealtimeService (WebRTC base)                        │   │
│  │ • InferenceService (Local models)                      │   │
│  │ • AuthService (User management)                        │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

## Required Components

### 1. Meeting Service (New)

**Proto Definition**: `proto/meetingservice.proto`
```protobuf
service MeetingService {
    rpc CreateMeeting(CreateMeetingRequest) returns (CreateMeetingResponse);
    rpc JoinMeeting(JoinMeetingRequest) returns (stream MeetingEvent);
    rpc LeaveMeeting(LeaveMeetingRequest) returns (LeaveMeetingResponse);
    rpc SendAudio(stream AudioChunk) returns (stream AudioChunk);
    rpc InviteJames(InviteJamesRequest) returns (InviteJamesResponse);
}

message CreateMeetingRequest {
    string meeting_name = 1;
    bool james_enabled = 2;
}

message MeetingEvent {
    oneof event {
        UserJoined user_joined = 1;
        UserLeft user_left = 2;
        AudioReceived audio_received = 3;
        JamesResponse james_response = 4;
        ScreenShare screen_share = 5;
    }
}

message AudioChunk {
    string user_id = 1;
    string user_name = 2;
    bytes audio_data = 3;
    int64 timestamp = 4;
}

message JamesResponse {
    string text = 1;
    bytes audio_data = 2;
    bytes screen_data = 3; // Generated visualization
    string meeting_summary = 4;
}
```

### 2. James Service (New)

**Proto Definition**: `proto/jamesservice.proto`
```protobuf
service JamesService {
    rpc ProcessMeetingAudio(stream MeetingAudioInput) returns (stream JamesOutput);
    rpc GetMeetingInsights(MeetingInsightsRequest) returns (MeetingInsightsResponse);
    rpc GenerateVisualization(VisualizationRequest) returns (VisualizationResponse);
}

message MeetingAudioInput {
    string meeting_id = 1;
    string speaker_id = 2;
    string speaker_name = 3;
    bytes audio_data = 4;
    int64 timestamp = 5;
}

message JamesOutput {
    oneof output {
        TranscriptionResult transcription = 1;
        JamesThought thought = 2;
        JamesResponse response = 3;
        VisualizationUpdate visualization = 4;
    }
}

message TranscriptionResult {
    string speaker_id = 1;
    string speaker_name = 2;
    string text = 3;
    float confidence = 4;
    int64 timestamp = 5;
}

message JamesThought {
    string context_summary = 1;
    repeated string key_points = 2;
    repeated string questions_identified = 3;
    string suggested_response = 4;
}
```

### 3. Backend Implementation Structure

```
backend/
├── meetingservice/
│   ├── service/
│   │   ├── meeting_service.go      # WebRTC coordination
│   │   ├── room_manager.go         # Meeting room state
│   │   └── james_coordinator.go    # James integration
│   ├── dao/
│   │   ├── meeting_dao.go          # Meeting persistence
│   │   └── models.go               # Meeting data models
│   └── api/
│       └── meeting_api.go          # gRPC handlers
│
├── jamesservice/
│   ├── service/
│   │   ├── james_service.go        # Main James logic
│   │   ├── speech_processor.go     # STT/TTS handling
│   │   ├── conversation_analyzer.go # Context understanding
│   │   ├── response_generator.go   # LLM integration
│   │   └── visualization_engine.go # Screen content generation
│   ├── dao/
│   │   ├── conversation_dao.go     # Conversation history
│   │   └── models.go              # James data models
│   └── api/
│       └── james_api.go           # gRPC handlers
│
└── Enhanced realtimeservice/
    ├── webrtc_coordinator.go      # Multi-user WebRTC
    └── audio_router.go            # Audio stream management
```

## James Core Features

### 1. Multi-Speaker Recognition
```go
type ConversationContext struct {
    MeetingID    string
    Participants map[string]Participant
    Timeline     []TranscriptEntry
    CurrentTopic string
    KeyPoints    []string
    Questions    []Question
}

type Participant struct {
    UserID     string
    Name       string
    VoiceID    string // For speaker identification
    SpeakTime  time.Duration
    KeyPoints  []string
}

type TranscriptEntry struct {
    Timestamp   time.Time
    SpeakerID   string
    SpeakerName string
    Text        string
    Confidence  float64
    Context     string
}
```

### 2. Conversation Understanding Engine
```go
type ConversationAnalyzer struct {
    llmClient    LLMClient
    context      *ConversationContext
    summaryModel string // "gpt-4" or local model
}

func (ca *ConversationAnalyzer) ProcessTranscript(entry TranscriptEntry) (*JamesThought, error) {
    // Analyze conversation flow
    // Identify questions directed at James
    // Generate contextual understanding
    // Prepare response suggestions
}
```

### 3. Visualization Generation
```go
type VisualizationEngine struct {
    templates map[string]VisualizationTemplate
    renderer  ScreenRenderer
}

type VisualizationTemplate struct {
    Type        string // "summary", "timeline", "mind_map", "action_items"
    Layout      string
    Components  []Component
}

// Generate visual explanations
func (ve *VisualizationEngine) GenerateExplanation(context *ConversationContext, topic string) ([]byte, error) {
    // Create visual representation of understanding
    // Generate mind maps, timelines, summaries
    // Render as shareable screen content
}
```

## Integration with Existing SortedChat

### 1. Extend ChatService
- Add meeting context to existing chat
- Link meeting transcripts to chat history
- Enable post-meeting chat continuation

### 2. Enhance RealtimeService
- Multi-user WebRTC coordination
- Audio routing to James
- Screen sharing from James

### 3. Leverage InferenceService
- Use local models for James when available
- Fallback to cloud LLMs
- Voice synthesis using local TTS models

## Data Flow

### Meeting Creation & James Invitation
```
1. User creates meeting via MeetingService
2. Other users join meeting room
3. Host invites James to meeting
4. James joins as virtual participant
5. WebRTC connections established (users ↔ users, users ↔ James)
```

### Real-time Processing
```
1. User speaks → Audio captured via WebRTC
2. Audio streamed to JamesService
3. Speech-to-Text conversion with speaker ID
4. Conversation analysis and context building
5. LLM processing for understanding and response
6. Response generation (text + audio + visuals)
7. James responds via WebRTC (audio + screen share)
```

### Database Schema

```sql
-- Meeting management
CREATE TABLE meetings (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    creator_id TEXT NOT NULL,
    james_enabled BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP,
    meeting_summary TEXT
);

CREATE TABLE meeting_participants (
    meeting_id TEXT NOT NULL,
    user_id TEXT NOT NULL,
    joined_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    left_at TIMESTAMP,
    speak_duration INTEGER DEFAULT 0,
    PRIMARY KEY (meeting_id, user_id)
);

-- James conversation tracking
CREATE TABLE meeting_transcripts (
    id TEXT PRIMARY KEY,
    meeting_id TEXT NOT NULL,
    speaker_id TEXT NOT NULL,
    speaker_name TEXT NOT NULL,
    text TEXT NOT NULL,
    confidence REAL NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    context_summary TEXT
);

CREATE TABLE james_thoughts (
    id TEXT PRIMARY KEY,
    meeting_id TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    context_summary TEXT,
    key_points JSON,
    questions_identified JSON,
    response_generated TEXT
);
```

## Local vs Cloud Deployment

### Cloud Mode (Default)
- **STT**: OpenAI Whisper API / Google Speech-to-Text
- **LLM**: GPT-4, Gemini, Claude via LiteLLM
- **TTS**: OpenAI TTS / Google Text-to-Speech
- **Benefits**: High accuracy, low local resource usage

### Local Mode (Privacy-focused)
- **STT**: Local Whisper model via llama.cpp
- **LLM**: Local models (Llama, Mistral) via InferenceService
- **TTS**: Local TTS models (Coqui TTS, etc.)
- **Benefits**: Complete privacy, offline operation

## Implementation Phases

### Phase 1: Basic Meeting Infrastructure
1. Extend RealtimeService for multi-user WebRTC
2. Create MeetingService for room management
3. Basic audio streaming between participants

### Phase 2: James Core Intelligence
1. Implement JamesService with STT/TTS
2. Basic conversation tracking and context
3. Simple text-based responses

### Phase 3: Advanced Features
1. Multi-speaker recognition and identification
2. Visual explanation generation
3. Screen sharing from James
4. Meeting summaries and insights

### Phase 4: Local Model Integration
1. Local STT/TTS via InferenceService
2. Local LLM integration for James
3. Offline meeting capability

## Questions for Implementation

1. **Voice Processing**: What STT/TTS APIs do you prefer for initial implementation?
2. **Speaker Identification**: Should we use voice biometrics or rely on WebRTC participant mapping?
3. **Visualization**: What type of visual explanations are most important (mind maps, timelines, charts)?
4. **Local Models**: Which local models work best for real-time voice processing?
5. **Meeting Persistence**: How long should meeting transcripts and James insights be stored?

This design leverages your existing SortedChat architecture while adding James as a natural extension. The modular approach allows incremental implementation and testing.

## Implementation Recommendations

### Start with Phase 1: Basic Meeting Infrastructure
1. **Extend existing RealtimeService** to support multiple participants
2. **Create MeetingService** as a new microservice in the backend
3. **Add meeting UI** to the existing frontend React components
4. **Test with 2-3 users** before adding James complexity

### Key Integration Points with SortedChat
1. **Reuse AuthService** for meeting participant authentication
2. **Extend ChatService** to link meeting transcripts with chat history
3. **Leverage InferenceService** for both James LLM processing and local model support
4. **Use existing gRPC infrastructure** for real-time communication

### Technical Considerations
1. **Audio Quality**: Ensure high-quality audio capture for accurate STT
2. **Latency**: Target <200ms for James response to feel natural
3. **Scalability**: Design for 10+ participants per meeting
4. **Privacy**: Implement meeting recording permissions and data retention policies

### Next Steps
1. **Prototype Phase 1** with basic multi-user WebRTC
2. **Define James personality** and response patterns  
3. **Choose initial STT/TTS providers** (recommend starting with OpenAI for reliability)
4. **Plan visualization templates** for common meeting scenarios
