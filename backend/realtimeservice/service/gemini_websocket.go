package service

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3"
)

// GeminiRealtime handles WebSocket communication with Gemini Live API
type GeminiRealtime struct {
	userID string
	apiKey string
	ws     *websocket.Conn
	// opusEncoder        *opus.Encoder
	// opusDecoder        *opus.Decoder
	outboundTrack      *webrtc.TrackLocalStaticRTP
	sequenceNumber     uint16
	timestamp          uint32
	connected          bool
	mu                 sync.RWMutex
	dataChannelManager *DataChannelManager
}

const opusPayloadType = 111
const ssrc = 1234

// Gemini message types based on API docs
type GeminiMessage struct {
	Setup         *GeminiSetup         `json:"setup,omitempty"`
	RealtimeInput *GeminiRealtimeInput `json:"realtimeInput,omitempty"`
}

type GeminiSetup struct {
	Model            string                  `json:"model"`
	GenerationConfig *GeminiGenerationConfig `json:"generationConfig,omitempty"`
}

type GeminiGenerationConfig struct {
	ResponseModalities []string `json:"responseModalities"`
}

type GeminiRealtimeInput struct {
	MediaChunks []GeminiMediaChunk `json:"mediaChunks"`
}

type GeminiMediaChunk struct {
	MimeType string `json:"mimeType"`
	Data     string `json:"data"`
}

// NewGeminiRealtime creates a new GeminiRealtime instance
func NewGeminiRealtime(userID string, outboundTrack *webrtc.TrackLocalStaticRTP, dataChannelManager *DataChannelManager) (*GeminiRealtime, error) {
	slog.Info("RealtimeService:gemini_websocket:NewGeminiRealtime")
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		slog.Error("RealtimeService:gemini_websocket:NewGeminiRealtime", "message", "GEMINI_API_KEY environment variable is required")
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is required")
	}

	// // Initialize Opus encoder/decoder
	// opusEncoder, err := opus.NewEncoder(48000, 1, opus.AppVoIP)
	// if err != nil {
	// 	slog.Error("RealtimeService:gemini_websocket:NewGeminiRealtime", "message", "failed to create Opus encoder", "error", err)
	// 	return nil, fmt.Errorf("failed to create Opus encoder: %v", err)
	// }

	// opusDecoder, err := opus.NewDecoder(48000, 1)
	// if err != nil {
	// 	slog.Error("RealtimeService:gemini_websocket:NewGeminiRealtime", "message", "failed to create Opus decoder", "error", err)
	// 	return nil, fmt.Errorf("failed to create Opus decoder: %v", err)
	// }

	return &GeminiRealtime{
		userID: userID,
		apiKey: apiKey,
		// opusEncoder:        opusEncoder,
		// opusDecoder:        opusDecoder,
		outboundTrack:      outboundTrack,
		sequenceNumber:     1,
		timestamp:          1,
		dataChannelManager: dataChannelManager,
	}, nil
}

func (g *GeminiRealtime) SetDataChannelManager(dcm *DataChannelManager) {
	slog.Info("RealtimeService:gemini_websocket:SetDataChannelManager")
	g.mu.Lock()
	defer g.mu.Unlock()
	g.dataChannelManager = dcm
}

// sendDataChannelMessage safely sends a message via data channel if available
func (g *GeminiRealtime) sendDataChannelMessage(messageType string, model string, data interface{}) {
	slog.Info("RealtimeService:gemini_websocket:sendDataChannelMessage")
	g.mu.RLock()
	dcm := g.dataChannelManager
	g.mu.RUnlock()

	if dcm != nil {
		dcm.sendMessageWithData(messageType, model, data)
	} else {
		slog.Debug("Data channel manager not available, skipping message", "userID", g.userID, "messageType", messageType)
	}
}

// Connect establishes WebSocket connection to Gemini Live API
func (g *GeminiRealtime) Connect(model string) error {
	slog.Info("RealtimeService:gemini_websocket:Connect")
	slog.Info("Connecting to Gemini Live API", "userID", g.userID)

	wsURL := "wss://generativelanguage.googleapis.com/ws/google.ai.generativelanguage.v1beta.GenerativeService.BidiGenerateContent?key=" + g.apiKey

	var err error
	g.ws, _, err = websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		slog.Error("RealtimeService:gemini_websocket:Connect", "message", "failed to connect to Gemini", "userID", g.userID, "error", err)
		return err
	}

	slog.Info("RealtimeService:gemini_websocket:Connect", "message", "Connected to Gemini WebSocket", "userID", g.userID)

	// Send setup message
	setupMsg := GeminiMessage{
		Setup: &GeminiSetup{
			Model: "models/" + model,
			GenerationConfig: &GeminiGenerationConfig{
				ResponseModalities: []string{"AUDIO"},
			},
		},
	}
	if err := g.ws.WriteJSON(setupMsg); err != nil {
		slog.Error("RealtimeService:gemini_websocket:Connect", "message", "failed to send setup", "userID", g.userID, "error", err)
		return err
	}

	// Wait for setup acknowledgment
	var setupResponse map[string]interface{}
	if err := g.ws.ReadJSON(&setupResponse); err != nil {
		slog.Error("RealtimeService:gemini_websocket:Connect", "message", "failed to read setup response", "userID", g.userID, "error", err)
		return err
	}

	g.mu.Lock()
	g.connected = true
	g.mu.Unlock()

	slog.Info("RealtimeService:gemini_websocket:Connect", "message", "Gemini setup complete", "userID", g.userID)

	// Start handling responses
	go g.handleResponses()

	return nil
}

// HandleAudioTrack processes incoming audio from WebRTC track
func (g *GeminiRealtime) HandleAudioTrack(track *webrtc.TrackRemote) {
	slog.Info("RealtimeService:gemini_websocket:HandleAudioTrack", "message", "Starting audio track handling for Gemini", "userID", g.userID)

	g.sendDataChannelMessage("Client_audio", "gemini", nil) //custom event

	opusPacket := &codecs.OpusPacket{}
	pcmBuffer := make([]int16, 0, 48000) // Buffer for 1 second at 48kHz

	for {
		rtpPacket, _, err := track.ReadRTP()
		if err != nil {
			slog.Error("Error reading RTP", "userID", g.userID, "error", err)
			return
		}

		// Extract Opus payload
		opusData, err := opusPacket.Unmarshal(rtpPacket.Payload)
		if err != nil {
			slog.Error("RealtimeService:gemini_websocket:HandleAudioTrack", "message", "failed to unmarshal Opus data", "userID", g.userID, "error", err)
			continue
		}

		if opusData == nil {
		}
		// Decode Opus to PCM
		pcmData := make([]int16, 960) // 20ms at 48kHz
		n := 10
		// n, err := g.opusDecoder.Decode(opusData, pcmData)
		// if err != nil {
		// 	slog.Error("RealtimeService:gemini_websocket:HandleAudioTrack", "message", "failed to decode Opus data", "userID", g.userID, "error", err)
		// 	continue
		// }

		// Accumulate PCM data
		pcmBuffer = append(pcmBuffer, pcmData[:n]...)

		if len(pcmBuffer) >= 4800 {
			// Downsample to 16kHz for Gemini
			pcm16kHz := g.downsample48to16(pcmBuffer)

			// Convert to bytes
			pcmBytes := make([]byte, len(pcm16kHz)*2)
			for i, sample := range pcm16kHz {
				binary.LittleEndian.PutUint16(pcmBytes[i*2:], uint16(sample))
			}

			g.SendAudio(pcmBytes)
			pcmBuffer = pcmBuffer[:0] // Clear buffer
		}
	}
}

// SendAudio sends audio data to Gemini
func (g *GeminiRealtime) SendAudio(audioData []byte) {
	g.mu.RLock()
	connected := g.connected
	g.mu.RUnlock()

	if !connected || g.ws == nil || len(audioData) == 0 {
		return
	}

	encodedAudio := base64.StdEncoding.EncodeToString(audioData)

	msg := GeminiMessage{
		RealtimeInput: &GeminiRealtimeInput{
			MediaChunks: []GeminiMediaChunk{
				{
					MimeType: "audio/pcm;rate=16000",
					Data:     encodedAudio,
				},
			},
		},
	}
	g.sendDataChannelMessage("sent_audio", "gemini", nil) //custom event

	if err := g.ws.WriteJSON(msg); err != nil {
		slog.Error("RealtimeService:gemini_websocket:SendAudio", "message", "failed to send to Gemini", "userID", g.userID, "error", err)
	}
}

// handleResponses processes incoming messages from Gemini
func (g *GeminiRealtime) handleResponses() {
	for {

		g.mu.RLock()
		connected := g.connected
		g.mu.RUnlock()

		if !connected {
			break
		}

		var response map[string]interface{}
		if err := g.ws.ReadJSON(&response); err != nil {
			slog.Error("Gemini connection closed", "userID", g.userID, "error", err)
			g.mu.Lock()
			g.connected = false
			g.mu.Unlock()
			return
		}

		// Extract audio from serverContent response
		if serverContent, ok := response["serverContent"].(map[string]interface{}); ok {
			if modelTurn, ok := serverContent["modelTurn"].(map[string]interface{}); ok {
				if parts, ok := modelTurn["parts"].([]interface{}); ok {
					for _, part := range parts {
						if partMap, ok := part.(map[string]interface{}); ok {
							if inlineData, ok := partMap["inlineData"].(map[string]interface{}); ok {
								if audioData, ok := inlineData["data"].(string); ok {
									slog.Info("Received audio from Gemini", "userID", g.userID, "chars", len(audioData))
									g.sendDataChannelMessage("recieving_audio", "gemini", nil) //custom event
									g.sendAudioToClient(audioData)
								}
							}
						}
					}
				}
			}
		}

		// Check for setup acknowledgment
		if setupComplete, ok := response["setupComplete"]; ok {
			slog.Info("Gemini setup acknowledged", "userID", g.userID, "setupComplete", setupComplete)
		}
	}
}

// sendAudioToClient sends processed audio to the WebRTC outbound track
func (g *GeminiRealtime) sendAudioToClient(base64Audio string) {
	if g.outboundTrack == nil {
		return
	}

	// Converts Gemini's base64-encoded audio back to raw bytes
	pcmData, err := base64.StdEncoding.DecodeString(base64Audio)
	if err != nil || len(pcmData) == 0 {
		slog.Error("Failed to decode audio", "userID", g.userID, "error", err)
		return
	}

	// Converts raw bytes to 16-bit audio samples using little-endian format, every 2 bytes becomes one int16 sample
	pcmSamples := make([]int16, len(pcmData)/2)
	for i := 0; i < len(pcmSamples); i++ {
		pcmSamples[i] = int16(binary.LittleEndian.Uint16(pcmData[i*2:]))
	}

	// Upsample from 24kHz to 48kHz
	pcm48kHz := g.upsample24to48(pcmSamples)

	// Encode PCM to Opus in 20ms frames (960 samples at 48kHz)
	const frameSize = 960 // 20ms at 48kHz can be changed as per need
	for i := 0; i < len(pcm48kHz); i += frameSize {
		end := i + frameSize
		if end > len(pcm48kHz) {
			// Pad last frame with zeros
			frame := make([]int16, frameSize)
			copy(frame, pcm48kHz[i:])
			pcm48kHz = append(pcm48kHz[:i], frame...)
			end = len(pcm48kHz)
		}

		frame := pcm48kHz[i:end]
		if frame == nil {

		}

		// Encode to Opus
		opusData := make([]byte, 4000)
		n := 10
		// n, err := g.opusEncoder.Encode(frame, opusData)
		// if err != nil {
		// 	slog.Error("Opus encoding error", "userID", g.userID, "error", err)
		// 	continue
		// }

		// Create RTP packet with proper timing
		rtpPacket := &rtp.Packet{
			Header: rtp.Header{
				Version:        2,
				PayloadType:    opusPayloadType, // Opus
				SequenceNumber: g.sequenceNumber,
				Timestamp:      g.timestamp,
				SSRC:           ssrc,
			},
			Payload: opusData[:n],
		}
		g.sequenceNumber++
		g.timestamp += 960 // 20ms worth of samples at 48kHz

		if err := g.outboundTrack.WriteRTP(rtpPacket); err != nil {
			slog.Error("Error sending RTP packet", "userID", g.userID, "error", err)
			break
		}
	}

}

// Close terminates the Gemini connection
func (g *GeminiRealtime) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.connected = false

	if g.ws != nil {
		err := g.ws.Close()
		g.ws = nil
		slog.Info("Closed Gemini WebSocket connection", "userID", g.userID)
		return err
	}

	return nil
}

// IsConnected returns the connection status
func (g *GeminiRealtime) IsConnected() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.connected
}

// Helper functions

// downsample48to16 converts 48kHz PCM to 16kHz (3:1 ratio)
// might need to change this logic and use github.com/zaf/resample for better quality
func (g *GeminiRealtime) downsample48to16(input []int16) []int16 {
	output := make([]int16, len(input)/3)
	for i := 0; i < len(output); i++ {
		output[i] = input[i*3] // Simple decimation
	}
	return output
}

// upsample24to48 converts 24kHz PCM to 48kHz (1:2 ratio)
func (g *GeminiRealtime) upsample24to48(input []int16) []int16 {
	output := make([]int16, len(input)*2)
	for i := 0; i < len(input); i++ {
		output[i*2] = input[i] // Original sample
		if i*2+1 < len(output) {
			output[i*2+1] = input[i] // Duplicate sample
		}
	}
	return output
}
