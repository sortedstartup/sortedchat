package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sortedstartup/realtimeservice/dao"
	pb "sortedstartup/realtimeservice/proto"

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
	browserConnection *webrtc.PeerConnection
	openaiConnection  *webrtc.PeerConnection

	backendToOpenAITrack *webrtc.TrackLocalStaticRTP
	openaiBackendTrack   *webrtc.TrackLocalStaticRTP
}

// userID -> PeerConnection (browserConnection and openaiConnection)
var userConnections = make(map[string]*PeerConnection)

func (s *RealtimeService) Offer(offer *pb.OfferRequest, userID string) (string, error) {
	browserToBackendPC, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		slog.Error("error creating browser to backend PC", "userID", userID, "error", err)
		return "", err
	}

	// Create track to recieve OpenAI audio from our backend (add before processing offer)
	openaiBackendTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000, Channels: 2},
		"openai-to-browser", "pion-openai",
	)
	if err != nil {
		slog.Error("error creating OpenAI to browser track", "userID", userID, "error", err)
		return "", err
	}

	// Add the track to browser connection BEFORE setting remote description
	if _, err := browserToBackendPC.AddTrack(openaiBackendTrack); err != nil {
		slog.Error("error adding OpenAI to backend track", "userID", userID, "error", err)
		return "", err
	}

	// OnTrack handler to receive audio FROM browser
	browserToBackendPC.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		slog.Info("Received track from browser", "trackID", track.ID(), "kind", track.Kind(), "userID", userID)
		if track.Kind() == webrtc.RTPCodecTypeAudio {
			slog.Info("Audio track received from browser", "userID", userID, "SSRC", track.SSRC())

			// Store this track for copying to OpenAI later
			if userConnections[userID] != nil && userConnections[userID].backendToOpenAITrack != nil {
				slog.Info("Copying audio track from browser to OpenAI", "userID", userID)
				copyAudioTrack(track, userConnections[userID].backendToOpenAITrack, userID, "Browser->OpenAI")
			}
		}
	})

	if err := browserToBackendPC.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offer.Offer,
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

	// Initialize connection structure
	userConnections[userID] = &PeerConnection{
		browserConnection:    browserToBackendPC,
		openaiConnection:     nil,
		backendToOpenAITrack: nil,
		openaiBackendTrack:   openaiBackendTrack,
	}

	go connectToOpenai(userID)
	return answerForBrowser.SDP, nil
}

// peer connection to openai
func connectToOpenai(userID string) error {
	backendToOpenAIpc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		slog.Error("error creating backend to OpenAI PC", "userID", userID, "error", err)
		return err
	}

	// Create track to send browser audio TO OpenAI
	backendToOpenAITrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 44000, Channels: 1},
		"browser-to-openai", "pion-browser",
	)
	if err != nil {
		slog.Error("error creating backend to OpenAI mic track", "userID", userID, "error", err)
		return err
	}

	// Add track to OpenAI connection
	if _, err := backendToOpenAIpc.AddTrack(backendToOpenAITrack); err != nil {
		slog.Error("error adding backend to OpenAI mic track", "userID", userID, "error", err)
		return err
	}

	// Store the track for copying from browser
	userConnections[userID].backendToOpenAITrack = backendToOpenAITrack
	userConnections[userID].openaiConnection = backendToOpenAIpc

	// OnTrack handler to receive audio FROM OpenAI
	backendToOpenAIpc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		slog.Info("Received track from OpenAI", "trackID", track.ID(), "kind", track.Kind(), "userID", userID)
		if track.Kind() == webrtc.RTPCodecTypeAudio {
			slog.Info("Audio track received from OpenAI", "userID", userID, "SSRC", track.SSRC())

			// Copy OpenAI audio to browser
			if userConnections[userID] != nil && userConnections[userID].openaiBackendTrack != nil {
				slog.Info("Copying audio track from OpenAI to browser", "userID", userID)
				copyAudioTrack(track, userConnections[userID].openaiBackendTrack, userID, "OpenAI->Browser")
			}
		}
	})

	// Create offer for OpenAI
	offerForOpenAI, err := backendToOpenAIpc.CreateOffer(nil)
	if err != nil {
		slog.Error("Error creating OpenAI offer", "userID", userID, "error", err)
		return err
	}

	backendToOpenAIpc.SetLocalDescription(offerForOpenAI)
	slog.Info("SDP Offer for OpenAI created with %d characters", "userID", userID, "length", len(offerForOpenAI.SDP))

	// Get ephemeral ephemeralToken
	ephemeralToken, err := getEphemeralToken()
	if err != nil {
		slog.Error("Ephemeral Token error", "userID", userID, "error", err)
		return err
	}

	openAISDPAnswer, err := getOpenAISDPAnswer(offerForOpenAI, ephemeralToken)
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

// gets the ephemral token from OpenAI
func getEphemeralToken() (string, error) {
	config := map[string]interface{}{
		"session": map[string]interface{}{
			"type":  "realtime",
			"model": "gpt-4o-mini-realtime-preview",
			"audio": map[string]interface{}{
				"output": map[string]interface{}{"voice": "alloy"},
				"input":  map[string]interface{}{"turn_detection": map[string]interface{}{"type": "server_vad"}},
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

// gets the SDP answer from OpenAI
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
	slog.Info("Response Body Length: %d chars", "length", len(responseBody))

	// Accept both 200 and 201
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		slog.Error("error getting OpenAI SDP", "status", resp.StatusCode, "body", responseBody)
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, responseBody)
	}

	return responseBody, nil
}

func (s *RealtimeService) IceCandidate(candidate string, userID string) (string, error) {
	if userConnections[userID].browserConnection == nil {
		slog.Error("client peer connection not initialized", "userID", userID)
		return "", fmt.Errorf("client peer connection not initialized")
	}

	// Check if remote description is set
	if userConnections[userID].browserConnection.RemoteDescription() == nil {
		slog.Error("remote description not set, cannot add ICE candidate", "userID", userID)
		return "", fmt.Errorf("remote description not set, cannot add ICE candidate")
	}

	if err := userConnections[userID].browserConnection.AddICECandidate(webrtc.ICECandidateInit{
		Candidate: candidate,
	}); err != nil {
		slog.Error("error adding ICE candidate", "userID", userID, "error", err)
		return "", err
	}
	return "connected", nil
}

// Generic function to copy audio from source track to destination track
func copyAudioTrack(sourceTrack *webrtc.TrackRemote, destTrack *webrtc.TrackLocalStaticRTP, userID, direction string) {
	go func() {
		buffer := make([]byte, 1400)
		slog.Info("Started goroutine to copy audio track", "direction", direction, "userID", userID)
		for {
			n, _, err := sourceTrack.Read(buffer)
			if err != nil {
				slog.Error("Error reading from source track", "direction", direction, "error", err)
				return
			}

			// Copy to destination track
			if _, err := destTrack.Write(buffer[:n]); err != nil {
				slog.Error("Error writing to destination track", "direction", direction, "error", err)
				return
			}
			slog.Debug("Copied audio data", "direction", direction, "bytes", n, "userID", userID)
		}
	}()
}
