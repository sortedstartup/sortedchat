package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sortedstartup/chatservice/proto"
	"sortedstartup/realtimeservice/dao"
	"sync"
	"time"

	"github.com/pion/webrtc/v3"
)

type RealtimeService struct {
	dao            dao.DAO
	settingsClient proto.SettingServiceClient
}

func NewRealtimeService(daoFactory dao.DAOFactory, settingsClient proto.SettingServiceClient) *RealtimeService {
	slog.Info("RealtimeService: NewRealtimeService")

	if settingsClient == nil {
		slog.Error("RealtimeService: NewRealtimeService", "message", "settingsClient cannot be nil")
		return nil
	}

	daoInstance, err := daoFactory.CreateDAO()
	if err != nil {
		slog.Error("RealtimeService: NewRealtimeService", "message", "Failed to create DAO", "error", err)
		return nil
	}
	return &RealtimeService{dao: daoInstance, settingsClient: settingsClient}
}

func (s *RealtimeService) Init(config *dao.Config) {
	slog.Info("RealtimeService: Init")
}

var OPENAI_API_KEY string
var GEMINI_API_KEY string

type PeerConnection struct {
	browserConnection    *webrtc.PeerConnection      //backend-browser peer connection
	openaiConnection     *webrtc.PeerConnection      // backend-openai peer connection
	geminiRealtime       *GeminiRealtime             //gemini websocket realtime
	openaiRealtime       *OpenAIRealtime             //openai websocket realtime
	backendToOpenAITrack *webrtc.TrackLocalStaticRTP //backend-openai track
	aiBackendTrack       *webrtc.TrackLocalStaticRTP //ai-backend track
	dataChannelManager   *DataChannelManager         //data channel manager between backend and browser
	audioChatDbID        string                      //audio chat database id
}

var (
	userConnections      = make(map[string]*PeerConnection)
	userConnectionsMutex sync.RWMutex
)

func (s *RealtimeService) Offer(offer string, provider string, model string, userID string) (string, error) {

	if provider != "gemini" && provider != "openai" {
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
	if model == "" {
		return "", fmt.Errorf("model cannot be empty")
	}

	settings, err := s.settingsClient.GetSetting(context.Background(), &proto.GetSettingRequest{})
	if err != nil {
		slog.Error("RealtimeService: Init", "message", "failed to get setting", "error", err)
		return "", err
	}

	OPENAI_API_KEY = settings.Settings.OPENAI_API_KEY
	GEMINI_API_KEY = settings.Settings.GEMINI_API_KEY

	slog.Info("RealtimeService: Offer", "offer", offer, "provider", provider, "model", model, "userID", userID)

	browserToBackendPC, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		slog.Error("RealtimeService: Offer", "message", "error creating browser to backend PC", "userID", userID, "error", err)
		return "", err
	}

	aiBackendTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: "audio/opus", ClockRate: 48000, Channels: 2},
		"ai-backend-track", "pion-ai",
	)
	if err != nil {
		slog.Error("RealtimeService: Offer", "message", "error creating AI to backend track", "userID", userID, "error", err)
		return "", err
	}
	if _, err := browserToBackendPC.AddTrack(aiBackendTrack); err != nil {
		slog.Error("RealtimeService: Offer", "message", "error adding AI to backend track", "userID", userID, "error", err)
		return "", err
	}

	// Check for existing connection and cleanup if needed
	userConnectionsMutex.Lock()
	existingUserConn := userConnections[userID]
	if existingUserConn != nil {
		s.Cleanup(userID)
	}
	userConnections[userID] = &PeerConnection{
		browserConnection:    browserToBackendPC,
		openaiConnection:     nil,
		geminiRealtime:       nil,
		backendToOpenAITrack: nil,
		aiBackendTrack:       aiBackendTrack,
		dataChannelManager:   nil,
	}
	userConnectionsMutex.Unlock()

	// OnTrack handler
	browserToBackendPC.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		slog.Info("Received track from browser", "trackID", track.ID(), "kind", track.Kind(), "userID", userID, "model", model)

		if track.Kind() == webrtc.RTPCodecTypeAudio {
			slog.Info("Audio track received from browser", "userID", userID, "SSRC", track.SSRC(), "model", model)

			userConnectionsMutex.RLock()
			userConn := userConnections[userID]
			userConnectionsMutex.RUnlock()

			if userConn == nil {
				slog.Error("User connection not found", "userID", userID)
				return
			}

			if provider == "gemini" {
				if userConn.geminiRealtime != nil {
					slog.Info("Handling audio track with Gemini", "userID", userID)
					go userConn.geminiRealtime.HandleAudioTrack(track)
				}
			} else {
				slog.Info("Copying audio track from browser to OpenAI", "userID", userID)
				if userConn.openaiRealtime != nil {
					slog.Info("Handling audio track with OpenAI", "userID", userID)
					go userConn.openaiRealtime.HandleAudioTrack(track)
				}
			}
		}
	})

	// Handle data channel creation
	browserToBackendPC.OnDataChannel(func(dc *webrtc.DataChannel) {
		slog.Info("Data channel created", "userID", userID)

		userConnectionsMutex.RLock()
		userConn := userConnections[userID]
		userConnectionsMutex.RUnlock()

		if userConn == nil {
			slog.Error("User connection not found when setting up data channel", "userID", userID)
			return
		}

		userConn.dataChannelManager = NewDataChannelManager(userID, dc, s)

		if userConn.openaiRealtime != nil {
			userConn.openaiRealtime.SetDataChannelManager(userConn.dataChannelManager)
		}

		if userConn.geminiRealtime != nil {
			userConn.geminiRealtime.SetDataChannelManager(userConn.dataChannelManager)
		}
	})

	if err := browserToBackendPC.SetRemoteDescription(webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  offer,
	}); err != nil {
		slog.Error("error setting browserToBackendPC.remote description", "userID", userID, "error", err)
		return "", err
	}

	answerForBrowser, err := browserToBackendPC.CreateAnswer(nil)
	if err != nil {
		slog.Error("error creating browserToBackendPC.answer", "userID", userID, "error", err)
		return "", err
	}
	if err := browserToBackendPC.SetLocalDescription(answerForBrowser); err != nil {
		slog.Error("error setting browserToBackendPC.local description", "userID", userID, "error", err)
		return "", err
	}

	if provider == "gemini" {
		go s.connectToGemini(userID, model)
	} else {
		go s.connectToOpenai(userID, model)
	}

	return answerForBrowser.SDP, nil
}

func (s *RealtimeService) connectToGemini(userID string, model string) error {
	userConnectionsMutex.RLock()
	userConn := userConnections[userID]
	userConnectionsMutex.RUnlock()

	if userConn == nil {
		slog.Error("User connection not found for Gemini setup", "userID", userID)
		return fmt.Errorf("user connection not found")
	}

	geminiRealtime, err := NewGeminiRealtime(userID, userConn.aiBackendTrack, userConn.dataChannelManager)
	if err != nil {
		slog.Error("Failed to create Gemini realtime instance", "userID", userID, "error", err)
		return err
	}

	userConn.geminiRealtime = geminiRealtime

	if err := geminiRealtime.Connect(model); err != nil {
		slog.Error("Failed to connect to Gemini", "userID", userID, "error", err)
		return err
	}

	if userConn.dataChannelManager != nil {
		userConn.dataChannelManager.sendMessageWithData("connected", "gemini", nil)
	} else {
		slog.Debug("Data channel manager not yet available, skipping connected message", "userID", userID)
	}

	slog.Info("Successfully connected to Gemini", "userID", userID)
	now := time.Now().Format(time.RFC3339)
	id, err := s.dao.CreateAudioChat(userID, model, now, now) //check time
	if err != nil {
		slog.Error("Failed to create audio chat", "userID", userID, "error", err)
		return err
	}
	userConn.audioChatDbID = id

	return nil
}

func (s *RealtimeService) connectToOpenai(userID string, model string) error {
	slog.Info("RealtimeService: connectToOpenai", "userID", userID, "model", model)

	userConnectionsMutex.RLock()
	userConn := userConnections[userID]
	userConnectionsMutex.RUnlock()

	if userConn == nil {
		slog.Error("User connection not found for OpenAI setup", "userID", userID)
		return fmt.Errorf("user connection not found")
	}

	openaiRealtime, err := NewOpenAIRealtime(userID, userConn.aiBackendTrack, userConn.dataChannelManager)
	if err != nil {
		slog.Error("Failed to create OpenAI realtime instance", "userID", userID, "error", err)
		return err
	}

	userConn.openaiRealtime = openaiRealtime

	if err := openaiRealtime.Connect(model); err != nil {
		slog.Error("Failed to connect to OpenAI", "userID", userID, "error", err)
		return err
	}

	if userConn.dataChannelManager != nil {
		userConn.dataChannelManager.sendMessageWithData("connected", "openai", nil)
	} else {
		slog.Debug("Data channel manager not yet available, skipping connected message", "userID", userID)
	}

	slog.Info("Successfully connected to OpenAI", "userID", userID)
	now := time.Now().Format(time.RFC3339)
	id, err := s.dao.CreateAudioChat(userID, model, now, now)
	if err != nil {
		slog.Error("Failed to create audio chat", "userID", userID, "error", err)
		return err
	}
	userConn.audioChatDbID = id

	return nil
}

func (s *RealtimeService) Cleanup(userID string) error {
	slog.Info("Cleaning up user connection", "userID", userID)

	userConnectionsMutex.Lock()
	userConn := userConnections[userID]
	if userConn == nil {
		userConnectionsMutex.Unlock()
		slog.Error("User connection not found for cleanup", "userID", userID)
		return nil
	}
	delete(userConnections, userID)
	userConnectionsMutex.Unlock()

	// Close AI connections
	if userConn.geminiRealtime != nil {
		userConn.geminiRealtime.Close()
	}
	if userConn.openaiRealtime != nil {
		userConn.openaiRealtime.Close()
	}

	if userConn.dataChannelManager != nil {
		userConn.dataChannelManager.Close()
	}

	if userConn.browserConnection != nil {
		userConn.browserConnection.Close()
	}

	if userConn.audioChatDbID != "" {
		err := s.dao.UpdateAudioChat(userID, userConn.audioChatDbID, time.Now().Format(time.RFC3339))
		if err != nil {
			slog.Error("Failed to update audio chat", "userID", userID, "error", err)
		}
	}
	delete(userConnections, userID)
	slog.Info("Cleaned up user connection", "userID", userID)
	return nil
}

func (s *RealtimeService) IceCandidate(candidate string, userID string) (string, error) {
	userConnectionsMutex.RLock()
	userConn := userConnections[userID]
	userConnectionsMutex.RUnlock()

	if userConn == nil {
		slog.Error("user connection not found", "userID", userID)
		return "", fmt.Errorf("user connection not found for userID: %s", userID)
	}

	if userConn.browserConnection == nil {
		slog.Error("browser peer connection not initialized", "userID", userID)
		return "", fmt.Errorf("browser peer connection not initialized for userID: %s", userID)
	}

	if userConn.browserConnection.RemoteDescription() == nil {
		slog.Error("remote description not set, cannot add ICE candidate", "userID", userID)
		return "", fmt.Errorf("remote description not set, cannot add ICE candidate for userID: %s", userID)
	}

	var iceCandidateInit webrtc.ICECandidateInit
	if err := json.Unmarshal([]byte(candidate), &iceCandidateInit); err != nil {
		slog.Error("error unmarshaling ICE candidate", "userID", userID, "error", err)
		return "", err
	}

	if err := userConn.browserConnection.AddICECandidate(iceCandidateInit); err != nil {
		slog.Error("error adding ICE candidate", "userID", userID, "error", err)
		return "", err
	}

	return "connected", nil
}
