package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/pion/webrtc/v4"
)

type OpenAIWebRTC struct {
	apiKey string
}

func NewOpenAIWebRTC(apiKey string) *OpenAIWebRTC {
	return &OpenAIWebRTC{
		apiKey: apiKey,
	}
}

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

func (o *OpenAIWebRTC) ConnectToOpenAI(userID string, userConn *PeerConnection) error {
	backendToOpenAIpc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		slog.Error("error creating backend to OpenAI PC", "userID", userID, "error", err)
		return err
	}

	backendToOpenAITrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000, Channels: 2},
		"browser-to-openai-track", "pion-browser",
	)
	if err != nil {
		slog.Error("error creating backend to OpenAI mic track", "userID", userID, "error", err)
		return err
	}

	if _, err := backendToOpenAIpc.AddTrack(backendToOpenAITrack); err != nil {
		slog.Error("error adding backend to OpenAI mic track", "userID", userID, "error", err)
		return err
	}

	userConn.backendToOpenAITrack = backendToOpenAITrack
	userConn.openaiConnection = backendToOpenAIpc

	backendToOpenAIpc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		slog.Info("Received track from OpenAI", "trackID", track.ID(), "kind", track.Kind(), "userID", userID)
		if track.Kind() == webrtc.RTPCodecTypeAudio {
			slog.Info("Audio track received from OpenAI", "userID", userID, "SSRC", track.SSRC())
			if userConn.aiBackendTrack != nil {
				slog.Info("Copying audio track from OpenAI to browser", "userID", userID)
				copyAudioTrack(track, userConn.aiBackendTrack, userID, "OpenAI->Browser")
			}
		}
	})

	o.setupConnectionEvents(backendToOpenAIpc, userID)
	o.setupDataChannel(backendToOpenAIpc, userID)

	offerForOpenAI, err := backendToOpenAIpc.CreateOffer(nil)
	if err != nil {
		slog.Error("Error creating OpenAI offer", "userID", userID, "error", err)
		return err
	}

	backendToOpenAIpc.SetLocalDescription(offerForOpenAI)
	slog.Info("SDP Offer for OpenAI created", "userID", userID, "length", len(offerForOpenAI.SDP))

	ephemeralToken, err := o.getEphemeralToken()
	if err != nil {
		slog.Error("Ephemeral Token error", "userID", userID, "error", err)
		return err
	}

	openAISDPAnswer, err := o.getOpenAISDPAnswer(offerForOpenAI, ephemeralToken)
	if err != nil {
		slog.Error("OpenAI SDP Answer error", "userID", userID, "error", err)
		return err
	}

	if err := backendToOpenAIpc.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  openAISDPAnswer,
	}); err != nil {
		slog.Error("error setting backend to OpenAI PC remote description", "userID", userID, "error", err)
		return err
	}

	slog.Info("Connected to OpenAI with bidirectional audio copying", "userID", userID)
	return nil
}

func (o *OpenAIWebRTC) setupConnectionEvents(pc *webrtc.PeerConnection, userID string) {
	var disconnectInitiator string = "unknown"

	pc.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		switch state {
		case webrtc.PeerConnectionStateDisconnected:
			disconnectInitiator = "network-issue"
			slog.Warn("OpenAI connection DISCONNECTED", "timestamp", timestamp, "initiator", "network-or-timeout", "userID", userID)
		case webrtc.PeerConnectionStateFailed:
			disconnectInitiator = "connection-failed"
			slog.Error("OpenAI connection FAILED", "timestamp", timestamp, "initiator", "ice-failure", "userID", userID)
		case webrtc.PeerConnectionStateClosed:
			slog.Info("OpenAI connection CLOSED", "timestamp", timestamp, "initiator", disconnectInitiator, "userID", userID)
		}
	})

	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		switch state {
		case webrtc.ICEConnectionStateDisconnected:
			disconnectInitiator = "ice-timeout"
			slog.Warn("ICE Disconnected - OpenAI side timeout", "timestamp", timestamp, "reason", "keepalive-timeout-or-network-issue", "userID", userID)
		case webrtc.ICEConnectionStateFailed:
			disconnectInitiator = "ice-failed"
			slog.Error("ICE Failed - Network unreachable", "timestamp", timestamp, "reason", "network-path-failed", "userID", userID)
		}
	})
}

func (o *OpenAIWebRTC) setupDataChannel(pc *webrtc.PeerConnection, userID string) {
	openaiDataChannel, _ := pc.CreateDataChannel("openai", nil)
	initEventLogging()

	openaiDataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		var message map[string]interface{}
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			if eventLogFile != nil {
				timestamp := time.Now().Format("2006-01-02 15:04:05")
				eventLogFile.WriteString(fmt.Sprintf("\n=== RAW_DATA | %s ===\n", timestamp))
				eventLogFile.WriteString(string(msg.Data))
				eventLogFile.WriteString("\n")
				eventLogFile.Sync()
			}
		} else {
			msgType, _ := message["type"].(string)
			if eventLogFile != nil {
				timestamp := time.Now().Format("2006-01-02 15:04:05")
				eventLogFile.WriteString(fmt.Sprintf("\n=== %s | %s ===\n", msgType, timestamp))
				prettyJSON, _ := json.MarshalIndent(message, "", "  ")
				eventLogFile.WriteString(string(prettyJSON))
				eventLogFile.WriteString("\n")
				eventLogFile.Sync()
			}
		}
	})

	openaiDataChannel.OnClose(func() {
		timestamp := time.Now().Format("2006-01-02 15:04:05.000")
		peerState := pc.ConnectionState()
		iceState := pc.ICEConnectionState()
		var closeReason string
		switch {
		case peerState == webrtc.PeerConnectionStateClosed:
			closeReason = "peer-connection-closed"
		case iceState == webrtc.ICEConnectionStateDisconnected:
			closeReason = "ice-timeout"
		case iceState == webrtc.ICEConnectionStateFailed:
			closeReason = "ice-failed"
		default:
			closeReason = "data-channel-closed"
		}
		slog.Info("❌ OpenAI data channel closed", "timestamp", timestamp, "closeReason", closeReason, "userID", userID)
	})
}

func (o *OpenAIWebRTC) getEphemeralToken() (string, error) {
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
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
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

func (o *OpenAIWebRTC) getOpenAISDPAnswer(offer webrtc.SessionDescription, ephemeralToken string) (string, error) {
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

	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		slog.Error("error getting OpenAI SDP", "status", resp.StatusCode, "body", responseBody)
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, responseBody)
	}
	return responseBody, nil
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
