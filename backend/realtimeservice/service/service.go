package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"sortedstartup/realtimeservice/dao"
	"time"

	"github.com/pion/webrtc/v3"
)

type RealtimeService struct {
	dao dao.Dao
}

func NewRealtimeService(dao dao.Dao) *RealtimeService {
	return &RealtimeService{dao: dao}
}

func (s *RealtimeService) Init(config *dao.Config) {
	slog.Info("RealtimeService: Init", "config", config)
}

var OPENAI_API_KEY = os.Getenv("OPENAI_API_KEY")

type PeerConnection struct {
	browserConnection    *webrtc.PeerConnection
	openaiConnection     *webrtc.PeerConnection
	geminiRealtime       *GeminiRealtime
	openaiRealtime       *OpenAIRealtime
	backendToOpenAITrack *webrtc.TrackLocalStaticRTP
	openaiBackendTrack   *webrtc.TrackLocalStaticRTP
	dataChannelManager   *DataChannelManager
}

// userID -> PeerConnection
var userConnections = make(map[string]*PeerConnection)

func (s *RealtimeService) Offer(offer string, model string, userID string) (string, error) {
	browserToBackendPC, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		slog.Error("error creating browser to backend PC", "userID", userID, "error", err)
		return "", err
	}

	openaiBackendTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000, Channels: 2},
		"ai-backend-track", "pion-ai",
	)
	if err != nil {
		slog.Error("error creating AI to browser track", "userID", userID, "error", err)
		return "", err
	}
	if _, err := browserToBackendPC.AddTrack(openaiBackendTrack); err != nil {
		slog.Error("error adding AI to backend track", "userID", userID, "error", err)
		return "", err
	}

	// Create user connection early so we can reference it in callbacks
	userConnections[userID] = &PeerConnection{
		browserConnection:    browserToBackendPC,
		openaiConnection:     nil,
		geminiRealtime:       nil,
		backendToOpenAITrack: nil,
		openaiBackendTrack:   openaiBackendTrack,
		dataChannelManager:   nil, // Will be set when data channel is created
	}

	// OnTrack handler (get browser audio)
	browserToBackendPC.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		slog.Info("Received track from browser", "trackID", track.ID(), "kind", track.Kind(), "userID", userID, "model", model)
		if track.Kind() == webrtc.RTPCodecTypeAudio {
			slog.Info("Audio track received from browser", "userID", userID, "SSRC", track.SSRC(), "model", model)

			userConn := userConnections[userID]
			if userConn == nil {
				slog.Error("User connection not found", "userID", userID)
				return
			}

			if model == "gemini" {
				// Handle audio through Gemini
				if userConn.geminiRealtime != nil {
					slog.Info("Handling audio track with Gemini", "userID", userID)
					go userConn.geminiRealtime.HandleAudioTrack(track)
				}
			} else {
				// Handle audio through OpenAI (existing logic)
				// if userConn.backendToOpenAITrack != nil {
				slog.Info("Copying audio track from browser to OpenAI", "userID", userID)
				// copyAudioTrack(track, userConn.backendToOpenAITrack, userID, "Browser->OpenAI")
				if userConn.openaiRealtime != nil {
					slog.Info("Handling audio track with OpenAI", "userID", userID)
					go userConn.openaiRealtime.HandleAudioTrack(track)
				}
				// }
			}
		}
	})

	// Handle data channel creation
	browserToBackendPC.OnDataChannel(func(dc *webrtc.DataChannel) {
		slog.Info("Data channel created", "userID", userID)

		userConn := userConnections[userID]
		if userConn == nil {
			slog.Error("User connection not found when setting up data channel", "userID", userID)
			return
		}

		// Create and assign the data channel manager
		userConn.dataChannelManager = NewDataChannelManager(userID, dc, s)

		// If OpenAI realtime is already created, set the data channel manager
		if userConn.openaiRealtime != nil {
			userConn.openaiRealtime.SetDataChannelManager(userConn.dataChannelManager)
		}
	})

	if err := browserToBackendPC.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offer,
	}); err != nil {
		slog.Error("error setting backendPC.remote description", "userID", userID, "error", err)
		return "", err
	}

	answerForBrowser, err := browserToBackendPC.CreateAnswer(nil)
	if err != nil {
		slog.Error("error creating backendPC.answer", "userID", userID, "error", err)
		return "", err
	}
	if err := browserToBackendPC.SetLocalDescription(answerForBrowser); err != nil {
		slog.Error("error setting backendPC.local description", "userID", userID, "error", err)
		return "", err
	}

	// Connect to the appropriate AI service
	if model == "gemini" {
		go s.connectToGemini(userID)
	} else {
		go s.connectToOpenai(userID)
	}

	return answerForBrowser.SDP, nil
}

// connectToGemini establishes connection to Gemini Live API
func (s *RealtimeService) connectToGemini(userID string) error {
	userConn := userConnections[userID]
	if userConn == nil {
		slog.Error("User connection not found for Gemini setup", "userID", userID)
		return fmt.Errorf("user connection not found")
	}

	// Create Gemini realtime instance
	geminiRealtime, err := NewGeminiRealtime(userID, userConn.openaiBackendTrack)
	if err != nil {
		slog.Error("Failed to create Gemini realtime instance", "userID", userID, "error", err)
		return err
	}

	// Store the Gemini instance
	userConn.geminiRealtime = geminiRealtime

	// Connect to Gemini
	if err := geminiRealtime.Connect(); err != nil {
		slog.Error("Failed to connect to Gemini", "userID", userID, "error", err)
		return err
	}

	slog.Info("Successfully connected to Gemini", "userID", userID)
	return nil
}

func (s *RealtimeService) connectToOpenai(userID string) error {
	userConn := userConnections[userID]
	if userConn == nil {
		slog.Error("User connection not found for OpenAI setup", "userID", userID)
		return fmt.Errorf("user connection not found")
	}

	openaiRealtime, err := NewOpenAIRealtime(userID, userConn.openaiBackendTrack, userConn.dataChannelManager)
	if err != nil {
		slog.Error("Failed to create OpenAI realtime instance", "userID", userID, "error", err)
		return err
	}

	userConn.openaiRealtime = openaiRealtime

	if err := openaiRealtime.Connect(); err != nil {
		slog.Error("Failed to connect to OpenAI", "userID", userID, "error", err)
		return err
	}

	slog.Info("Successfully connected to OpenAI", "userID", userID)
	return nil
}

func (s *RealtimeService) Cleanup(userID string) error {
	userConn := userConnections[userID]
	if userConn == nil {
		return nil
	}

	// Close AI connections
	if userConn.geminiRealtime != nil {
		userConn.geminiRealtime.Close()
	}
	if userConn.openaiRealtime != nil {
		userConn.openaiRealtime.Close()
	}

	// Close data channel manager
	if userConn.dataChannelManager != nil {
		userConn.dataChannelManager.Close()
	}

	// Close WebRTC connection
	if userConn.browserConnection != nil {
		userConn.browserConnection.Close()
	}

	delete(userConnections, userID)
	slog.Info("Cleaned up user connection", "userID", userID)
	return nil
}

// func connectToOpenai(userID string) error {
// 	backendToOpenAIpc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
// 	if err != nil {
// 		slog.Error("error creating backend to OpenAI PC", "userID", userID, "error", err)
// 		return err
// 	}

// 	backendToOpenAITrack, err := webrtc.NewTrackLocalStaticRTP(
// 		webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000, Channels: 2},
// 		"browser-to-openai-track", "pion-browser",
// 	)
// 	if err != nil {
// 		slog.Error("error creating backend to OpenAI mic track", "userID", userID, "error", err)
// 		return err
// 	}
// 	if _, err := backendToOpenAIpc.AddTrack(backendToOpenAITrack); err != nil {
// 		slog.Error("error adding backend to OpenAI mic track", "userID", userID, "error", err)
// 		return err
// 	}

// 	userConnections[userID].backendToOpenAITrack = backendToOpenAITrack
// 	userConnections[userID].openaiConnection = backendToOpenAIpc

// 	backendToOpenAIpc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
// 		slog.Info("Received track from OpenAI", "trackID", track.ID(), "kind", track.Kind(), "userID", userID)
// 		if track.Kind() == webrtc.RTPCodecTypeAudio {
// 			slog.Info("Audio track received from OpenAI", "userID", userID, "SSRC", track.SSRC())
// 			if userConnections[userID] != nil && userConnections[userID].openaiBackendTrack != nil {
// 				slog.Info("Copying audio track from OpenAI to browser", "userID", userID)
// 				copyAudioTrack(track, userConnections[userID].openaiBackendTrack, userID, "OpenAI->Browser")
// 			}
// 		}
// 	})

// 	// ---- [Connection Events/Logging] ----
// 	var disconnectInitiator string = "unknown"
// 	backendToOpenAIpc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
// 		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
// 		switch state {
// 		case webrtc.PeerConnectionStateDisconnected:
// 			disconnectInitiator = "network-issue"
// 			slog.Warn("OpenAI connection DISCONNECTED", "timestamp", timestamp, "initiator", "network-or-timeout", "userID", userID)
// 		case webrtc.PeerConnectionStateFailed:
// 			disconnectInitiator = "connection-failed"
// 			slog.Error("OpenAI connection FAILED", "timestamp", timestamp, "initiator", "ice-failure", "userID", userID)
// 		case webrtc.PeerConnectionStateClosed:
// 			slog.Info("OpenAI connection CLOSED", "timestamp", timestamp, "initiator", disconnectInitiator, "userID", userID)
// 		}
// 	})
// 	backendToOpenAIpc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
// 		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
// 		switch state {
// 		case webrtc.ICEConnectionStateDisconnected:
// 			disconnectInitiator = "ice-timeout"
// 			slog.Warn("ICE Disconnected - OpenAI side timeout", "timestamp", timestamp, "reason", "keepalive-timeout-or-network-issue", "userID", userID)
// 		case webrtc.ICEConnectionStateFailed:
// 			disconnectInitiator = "ice-failed"
// 			slog.Error("ICE Failed - Network unreachable", "timestamp", timestamp, "reason", "network-path-failed", "userID", userID)
// 		}
// 	})

// 	// ---- [Data Channel and Event Logging for testing] ----
// 	openaiDataChannel, _ := backendToOpenAIpc.CreateDataChannel("openai", nil)
// 	initEventLogging()
// 	openaiDataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
// 		var message map[string]interface{}
// 		if err := json.Unmarshal(msg.Data, &message); err != nil {
// 			// Log raw data
// 			if eventLogFile != nil {
// 				timestamp := time.Now().Format("2006-01-02 15:04:05")
// 				eventLogFile.WriteString(fmt.Sprintf("\n=== RAW_DATA | %s ===\n", timestamp))
// 				eventLogFile.WriteString(string(msg.Data))
// 				eventLogFile.WriteString("\n")
// 				eventLogFile.Sync()
// 			}
// 		} else {
// 			msgType, _ := message["type"].(string)
// 			if eventLogFile != nil {
// 				timestamp := time.Now().Format("2006-01-02 15:04:05")
// 				eventLogFile.WriteString(fmt.Sprintf("\n=== %s | %s ===\n", msgType, timestamp))
// 				prettyJSON, _ := json.MarshalIndent(message, "", "  ")
// 				eventLogFile.WriteString(string(prettyJSON))
// 				eventLogFile.WriteString("\n")
// 				eventLogFile.Sync()
// 			}
// 		}
// 	})
// 	openaiDataChannel.OnClose(func() {
// 		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
// 		peerState := backendToOpenAIpc.ConnectionState()
// 		iceState := backendToOpenAIpc.ICEConnectionState()
// 		var closeReason string
// 		switch {
// 		case peerState == webrtc.PeerConnectionStateClosed:
// 			closeReason = "peer-connection-closed"
// 		case iceState == webrtc.ICEConnectionStateDisconnected:
// 			closeReason = "ice-timeout"
// 		case iceState == webrtc.ICEConnectionStateFailed:
// 			closeReason = "ice-failed"
// 		default:
// 			closeReason = "data-channel-closed"
// 		}
// 		slog.Info("❌ OpenAI data channel closed", "timestamp", timestamp, "closeReason", closeReason, "peerState", peerState.String(), "iceState", iceState.String(), "initiator", disconnectInitiator, "userID", userID)
// 		if eventLogFile != nil {
// 			eventLogFile.WriteString(fmt.Sprintf("\n=== DATA_CHANNEL_CLOSED | %s ===\n", timestamp))
// 			eventLogFile.WriteString(fmt.Sprintf("Close Reason: %s\n", closeReason))
// 			eventLogFile.WriteString(fmt.Sprintf("Peer State: %s\n", peerState.String()))
// 			eventLogFile.WriteString(fmt.Sprintf("ICE State: %s\n", iceState.String()))
// 			eventLogFile.WriteString(fmt.Sprintf("Suspected Initiator: %s\n", disconnectInitiator))
// 			eventLogFile.Sync()
// 		}
// 	})

// 	// --- Create Offer for OpenAI ---
// 	offerForOpenAI, err := backendToOpenAIpc.CreateOffer(nil)
// 	if err != nil {
// 		slog.Error("Error creating OpenAI offer", "userID", userID, "error", err)
// 		return err
// 	}

// 	backendToOpenAIpc.SetLocalDescription(offerForOpenAI)
// 	slog.Info("SDP Offer for OpenAI created with %d characters", "userID", userID, "length", len(offerForOpenAI.SDP))

// 	// --- Get ephemeral token/session ---
// 	ephemeralToken, err := getEphemeralToken()
// 	if err != nil {
// 		slog.Error("Ephemeral Token error", "userID", userID, "error", err)
// 		return err
// 	}
// 	openAISDPAnswer, err := getOpenAISDPAnswer(offerForOpenAI, ephemeralToken)
// 	if err != nil {
// 		slog.Error("OpenAI SDP Answer error", "userID", userID, "error", err)
// 		return err
// 	}

// 	if err := backendToOpenAIpc.SetRemoteDescription(webrtc.SessionDescription{
// 		Type: webrtc.SDPTypeAnswer,
// 		SDP:  openAISDPAnswer,
// 	}); err != nil {
// 		slog.Error("error setting backend to OpenAI PC remote description", "userID", userID, "error", err)
// 		return err
// 	}
// 	slog.Info("Connected to OpenAI with bidirectional audio copying", "userID", userID)
// 	return nil
// }

// Logging file for OpenAI events
// ----------------------for testing----------------------
var eventLogFile *os.File

func initEventLogging() {
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	filename := fmt.Sprintf("openai_events_%s.txt", timestamp)
	var err error
	eventLogFile, err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to create log file: %v", err)
		return
	}
	fmt.Printf("📝 Logging to: %s\n", filename)
}

//----------------------for testing----------------------

func getEphemeralToken() (string, error) {
	config := map[string]interface{}{
		"session": map[string]interface{}{
			"type":  "realtime",
			"model": "gpt-4o-mini-realtime-preview",
			"audio": map[string]interface{}{
				"output": map[string]interface{}{"voice": "alloy"},
				"input": map[string]interface{}{
					"turn_detection": map[string]interface{}{
						"type": "server_vad", "idle_timeout_ms": 30000,
					},
				},
			},
			"instructions": "You are helpful. Answer briefly in ENGLISH only.",
		},
	}
	data, _ := json.Marshal(config)
	req, _ := http.NewRequest("POST", "https://api.openai.com/v1/realtime/client_secrets", bytes.NewBuffer(data))
	req.Header.Set("Authorization", "Bearer "+OPENAI_API_KEY)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("error getting ephemeral token", "error", err)
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if value, ok := result["value"].(string); ok {
		slog.Info("ephemeral token received", "token", value)
		return value, nil
	}
	slog.Error("no token received")
	return "", fmt.Errorf("no token received")
}

func getOpenAISDPAnswer(offer webrtc.SessionDescription, ephemeralToken string) (string, error) {
	req, _ := http.NewRequest("POST",
		"https://api.openai.com/v1/realtime/calls?model=gpt-4o-mini-realtime-preview",
		bytes.NewBufferString(offer.SDP))
	req.Header.Set("Authorization", "Bearer "+ephemeralToken)
	req.Header.Set("Content-Type", "application/sdp")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("error getting OpenAI SDP", "error", err)
		return "", err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	responseBody := buf.String()

	slog.Info("Response Status: %d", "status", resp.StatusCode)
	slog.Info("Response Body Length", "body", string(responseBody))

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		slog.Error("error getting OpenAI SDP", "status", resp.StatusCode, "body", responseBody)
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, responseBody)
	}
	return responseBody, nil
}

func (s *RealtimeService) IceCandidate(candidate string, userID string) (string, error) {
	// First check if the user connection exists
	userConn := userConnections[userID]
	if userConn == nil {
		slog.Error("user connection not found", "userID", userID)
		return "", fmt.Errorf("user connection not found for userID: %s", userID)
	}

	// Then check if the browser connection exists
	if userConn.browserConnection == nil {
		slog.Error("browser peer connection not initialized", "userID", userID)
		return "", fmt.Errorf("browser peer connection not initialized for userID: %s", userID)
	}

	// Check if remote description is set
	if userConn.browserConnection.RemoteDescription() == nil {
		slog.Error("remote description not set, cannot add ICE candidate", "userID", userID)
		return "", fmt.Errorf("remote description not set, cannot add ICE candidate for userID: %s", userID)
	}

	// Add ICE candidate
	if err := userConn.browserConnection.AddICECandidate(webrtc.ICECandidateInit{
		Candidate: candidate,
	}); err != nil {
		slog.Error("error adding ICE candidate", "userID", userID, "error", err)
		return "", err
	}

	return "connected", nil
}

// Robust, safe copy of audio data between tracks
func copyAudioTrack(sourceTrack *webrtc.TrackRemote, destTrack *webrtc.TrackLocalStaticRTP, userID, direction string) {
	go func() {
		buffer := make([]byte, 1400)
		slog.Info("Started goroutine to copy audio track", "direction", direction, "userID", userID)
		for {
			// Validate connection/track
			if userConn := userConnections[userID]; userConn == nil || destTrack == nil {
				slog.Info("Connection or track missing, stopping copy", "direction", direction, "userID", userID)
				return
			}
			n, _, err := sourceTrack.Read(buffer)
			if err != nil {
				slog.Error("Error reading from source track", "direction", direction, "error", err)
				return
			}
			if _, err := destTrack.Write(buffer[:n]); err != nil {
				slog.Error("Error writing to destination track", "direction", direction, "error", err)
				return
			}
		}
	}()
}
