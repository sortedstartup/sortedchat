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
}

// userID -> PeerConnection (browserConnection and openaiConnection)
var userConnections = make(map[string]*PeerConnection)

func (s *RealtimeService) Offer(offer *pb.OfferRequest, userID string) (string, error) {

	browserToBackendPC, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return "", err
	}

	browserToBackendPC.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		slog.Info("Received track from browser", "trackID", track.ID(), "kind", track.Kind(), "userID", userID)

		if track.Kind() == webrtc.RTPCodecTypeAudio {

			// Start reading audio data in a goroutine
			go func() {
				buffer := make([]byte, 1400)
				for {
					n, _, err := track.Read(buffer)
					if err != nil {
						slog.Error("Error reading audio data", "error", err)
						return
					}
					// Log that we received audio data
					slog.Info("Received audio data", "bytes", n, "userID", userID)
				}
			}()
		}
	})

	if err := browserToBackendPC.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offer.Offer,
	}); err != nil {
		slog.Error("error setting backendPC.remote description", "error", err)
		return "", err
	}

	answerForBrowser, err := browserToBackendPC.CreateAnswer(nil)
	if err != nil {
		slog.Error("error creating backendPC.answer", "error", err)
		return "", err
	}

	if err := browserToBackendPC.SetLocalDescription(answerForBrowser); err != nil {
		slog.Error("error setting backendPC.local description", "error", err)
		return "", err
	}

	userConnections[userID] = &PeerConnection{
		browserConnection: browserToBackendPC,
		openaiConnection:  nil,
	}

	go connectToOpenai(userID)

	return answerForBrowser.SDP, nil
}

// peer connection to openai
func connectToOpenai(userID string) error {

	backendToOpenAIPC, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		return err
	}

	// add empty audio track to backendToOpenAIPC
	emptyAudioTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 44000, Channels: 1},
		"openai-audio", "pion-openai",
	)
	if err != nil {
		slog.Error("error creating empty audio track", "error", err)
		return err
	}
	if _, err := backendToOpenAIPC.AddTrack(emptyAudioTrack); err != nil {
		slog.Error("error adding empty audio track to backendPC", "error", err)
		return err
	}

	// create offer for openai
	offerForOpenAI, err := backendToOpenAIPC.CreateOffer(nil)
	if err != nil {
		fmt.Printf("Error creating OpenAI offer: %v\n", err)
		return err
	}

	backendToOpenAIPC.SetLocalDescription(offerForOpenAI)
	fmt.Printf("OpenAI Offer created with %d characters\n", len(offerForOpenAI.SDP))

	// Get ephemral token
	token, err := getEphemeralToken()
	if err != nil {
		fmt.Printf("Token error: %v\n", err)
		return err
	}

	responseBody, err := getOpenaiSDP(offerForOpenAI, token)
	if err != nil {
		fmt.Printf("OpenAI error: %v\n", err)
		return err
	}

	backendToOpenAIPC.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeAnswer,
		SDP:  responseBody,
	})
	userConnections[userID].openaiConnection = backendToOpenAIPC
	log.Println("Connected to OpenAI ")
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
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if value, ok := result["value"].(string); ok {
		return value, nil
	}
	return "", fmt.Errorf("no token received")
}

// gets the SDP answer from OpenAI
func getOpenaiSDP(offer webrtc.SessionDescription, token string) (string, error) {
	req, _ := http.NewRequest("POST",
		"https://api.openai.com/v1/realtime/calls?model=gpt-4o-mini-realtime-preview",
		bytes.NewBufferString(offer.SDP))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/sdp")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	responseBody := buf.String()

	fmt.Printf("Response Status: %d\n", resp.StatusCode)
	fmt.Printf("Response Body Length: %d chars\n", len(responseBody))

	// Accept both 200 and 201
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, responseBody)
	}

	return responseBody, nil
}

func (s *RealtimeService) IceCandidate(candidate string, userID string) (string, error) {
	if userConnections[userID].browserConnection == nil {
		return "", fmt.Errorf("client peer connection not initialized")
	}

	// Check if remote description is set
	if userConnections[userID].browserConnection.RemoteDescription() == nil {
		return "", fmt.Errorf("remote description not set, cannot add ICE candidate")
	}

	if err := userConnections[userID].browserConnection.AddICECandidate(webrtc.ICECandidateInit{
		Candidate: candidate,
	}); err != nil {
		slog.Error("error adding ICE candidate", "error", err)
		return "", err
	}
	return "connected", nil
}
