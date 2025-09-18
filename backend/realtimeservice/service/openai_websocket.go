package service

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"os"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v3"
	"gopkg.in/hraban/opus.v2"
)

// OpenAIRealtime handles WebSocket communication with OpenAI Realtime API
type OpenAIRealtime struct {
	userID             string
	apiKey             string
	ws                 *websocket.Conn
	opusEncoder        *opus.Encoder
	opusDecoder        *opus.Decoder
	outboundTrack      *webrtc.TrackLocalStaticRTP
	sequenceNumber     uint16
	timestamp          uint32
	connected          bool
	mu                 sync.RWMutex
	dataChannelManager *DataChannelManager
}

// OpenAI message types based on the actual API response format
type OpenAIMessage struct {
	EventID  string                  `json:"event_id,omitempty"`
	Type     string                  `json:"type"`
	Session  *OpenAISession          `json:"session,omitempty"`
	Audio    string                  `json:"audio,omitempty"`
	Response *OpenAIResponseConfig   `json:"response,omitempty"`
	Item     *OpenAIConversationItem `json:"item,omitempty"`
}

// Session structure based on the actual API response format from your logs
type OpenAISession struct {
	Instructions            string                    `json:"instructions,omitempty"`
	Voice                   string                    `json:"voice,omitempty"`
	InputAudioFormat        string                    `json:"input_audio_format,omitempty"`
	OutputAudioFormat       string                    `json:"output_audio_format,omitempty"`
	InputAudioTranscription *OpenAIAudioTranscription `json:"input_audio_transcription,omitempty"`
	TurnDetection           *OpenAITurnDetection      `json:"turn_detection,omitempty"`
	MaxResponseOutputTokens interface{}               `json:"max_response_output_tokens,omitempty"`
	Modalities              []string                  `json:"modalities,omitempty"`
	Speed                   float64                   `json:"speed,omitempty"`
}

type OpenAIAudioTranscription struct {
	Model string `json:"model"`
}

// Turn detection based on actual API response structure
type OpenAITurnDetection struct {
	Type              string      `json:"type"`
	Threshold         float64     `json:"threshold,omitempty"`
	PrefixPaddingMs   int         `json:"prefix_padding_ms,omitempty"`
	SilenceDurationMs int         `json:"silence_duration_ms,omitempty"`
	CreateResponse    bool        `json:"create_response,omitempty"`
	InterruptResponse bool        `json:"interrupt_response,omitempty"`
	IdleTimeoutMs     interface{} `json:"idle_timeout_ms,omitempty"`
}

type OpenAIResponseConfig struct {
	Modalities []string `json:"modalities,omitempty"`
}

type OpenAIConversationItem struct {
	Type    string        `json:"type"`
	Role    string        `json:"role,omitempty"`
	Content []interface{} `json:"content,omitempty"`
}

type ResponseDoneEvent struct {
	Type     string         `json:"type"`
	EventID  string         `json:"event_id"`
	Response ResponseDetail `json:"response"`
}

type ResponseDetail struct {
	Object        string       `json:"object"`
	ID            string       `json:"id"`
	Status        string       `json:"status"`
	StatusDetails interface{}  `json:"status_details"`
	Output        []OutputItem `json:"output"`
	Usage         UsageDetail  `json:"usage"`
	Metadata      interface{}  `json:"metadata"`
}

type OutputItem struct {
	Object    string `json:"object"`
	ID        string `json:"id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Name      string `json:"name,omitempty"`
	CallID    string `json:"call_id,omitempty"`
	Arguments string `json:"arguments,omitempty"`
}

type UsageDetail struct {
	TotalTokens        int                `json:"total_tokens"`
	InputTokens        int                `json:"input_tokens"`
	OutputTokens       int                `json:"output_tokens"`
	InputTokenDetails  InputTokenDetails  `json:"input_token_details"`
	OutputTokenDetails OutputTokenDetails `json:"output_token_details"`
}

type InputTokenDetails struct {
	TextTokens          int                `json:"text_tokens"`
	AudioTokens         int                `json:"audio_tokens"`
	CachedTokens        int                `json:"cached_tokens"`
	CachedTokensDetails CachedTokenDetails `json:"cached_tokens_details"`
}

type OutputTokenDetails struct {
	TextTokens  int `json:"text_tokens"`
	AudioTokens int `json:"audio_tokens"`
}

type CachedTokenDetails struct {
	TextTokens  int `json:"text_tokens"`
	AudioTokens int `json:"audio_tokens"`
}

// NewOpenAIRealtime creates a new OpenAIRealtime instance
func NewOpenAIRealtime(userID string, outboundTrack *webrtc.TrackLocalStaticRTP, dataChannelManager *DataChannelManager) (*OpenAIRealtime, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	// Initialize Opus encoder/decoder for 24kHz (OpenAI requirement)
	opusEncoder, err := opus.NewEncoder(48000, 1, opus.AppVoIP)
	if err != nil {
		return nil, fmt.Errorf("failed to create Opus encoder: %v", err)
	}
	opusEncoder.SetBitrate(64000) // 64kbps

	opusDecoder, err := opus.NewDecoder(48000, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to create Opus decoder: %v", err)
	}

	return &OpenAIRealtime{
		userID:             userID,
		apiKey:             apiKey,
		opusEncoder:        opusEncoder,
		opusDecoder:        opusDecoder,
		outboundTrack:      outboundTrack,
		sequenceNumber:     1,
		timestamp:          1,
		dataChannelManager: dataChannelManager,
	}, nil
}

// SetDataChannelManager sets the data channel manager (used when it becomes available later)
func (o *OpenAIRealtime) SetDataChannelManager(dcm *DataChannelManager) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.dataChannelManager = dcm
}

// sendDataChannelMessage safely sends a message via data channel if available
func (o *OpenAIRealtime) sendDataChannelMessage(messageType, model string) {
	o.mu.RLock()
	dcm := o.dataChannelManager
	o.mu.RUnlock()

	if dcm != nil {
		dcm.sendMessage(messageType, model)
	} else {
		slog.Debug("Data channel manager not available, skipping message", "userID", o.userID, "messageType", messageType)
	}
}

// Connect establishes WebSocket connection to OpenAI Realtime API
func (o *OpenAIRealtime) Connect() error {
	slog.Info("Connecting to OpenAI Realtime API", "userID", o.userID)

	// Use the correct URL format
	wsURL := "wss://api.openai.com/v1/realtime?model=gpt-realtime"

	headers := make(map[string][]string)
	headers["Authorization"] = []string{"Bearer " + o.apiKey}
	headers["OpenAI-Beta"] = []string{"realtime=v1"}

	var err error
	o.ws, _, err = websocket.DefaultDialer.Dial(wsURL, headers)
	if err != nil {
		slog.Error("Failed to connect to OpenAI", "userID", o.userID, "error", err)
		return err
	}

	slog.Info("Successfully connected to OpenAI", "userID", o.userID)

	o.mu.Lock()
	o.connected = true
	o.mu.Unlock()

	// Start handling responses first
	go o.handleResponses()

	sessionMsg := OpenAIMessage{
		Type: "session.update",
		Session: &OpenAISession{
			Instructions:      "Speak clearly and briefly. Respond naturally and conversationally.",
			Voice:             "alloy",
			InputAudioFormat:  "pcm16",
			OutputAudioFormat: "pcm16",
			Modalities:        []string{"text", "audio"},
			TurnDetection: &OpenAITurnDetection{
				Type:              "server_vad",
				Threshold:         0.5,
				PrefixPaddingMs:   300,
				SilenceDurationMs: 200,
				CreateResponse:    true,
				InterruptResponse: true,
				IdleTimeoutMs:     nil,
			},
			MaxResponseOutputTokens: "inf",
		},
	}

	if err := o.ws.WriteJSON(sessionMsg); err != nil {
		slog.Error("Failed to send session update", "userID", o.userID, "error", err)
		return err
	}

	slog.Info("OpenAI session update sent", "userID", o.userID)

	return nil
}

// HandleAudioTrack processes incoming audio from WebRTC track
func (o *OpenAIRealtime) HandleAudioTrack(track *webrtc.TrackRemote) {
	log.Println("Starting audio track handling for OpenAI", "userID", o.userID)
	slog.Info("Starting audio track handling for OpenAI", "userID", o.userID)

	opusPacket := &codecs.OpusPacket{}
	pcmBuffer := make([]int16, 0, 48000) // Buffer for 1 second at 48kHz

	for {
		rtpPacket, _, err := track.ReadRTP()
		if err != nil {
			slog.Error("Error reading RTP", "userID", o.userID, "error", err)
			return
		}

		// Extract Opus payload
		opusData, err := opusPacket.Unmarshal(rtpPacket.Payload)
		if err != nil {
			continue
		}

		// Decode Opus to PCM
		pcmData := make([]int16, 960) // 20ms at 48kHz
		n, err := o.opusDecoder.Decode(opusData, pcmData)
		if err != nil {
			continue
		}

		// Accumulate PCM data
		pcmBuffer = append(pcmBuffer, pcmData[:n]...)

		if len(pcmBuffer) >= 24000 {
			// Downsample to 24kHz for OpenAI (OpenAI uses 24kHz PCM16)
			pcm24kHz := o.downsample48to24(pcmBuffer)

			// Convert to bytes (little-endian 16-bit)
			pcmBytes := make([]byte, len(pcm24kHz)*2)
			for i, sample := range pcm24kHz {
				binary.LittleEndian.PutUint16(pcmBytes[i*2:], uint16(sample))
			}

			o.SendAudio(pcmBytes)
			pcmBuffer = pcmBuffer[:0] // Clear buffer
		}
	}
}

// SendAudio sends audio data to OpenAI using input_audio_buffer.append
func (o *OpenAIRealtime) SendAudio(audioData []byte) {
	o.mu.RLock()
	connected := o.connected
	o.mu.RUnlock()

	if !connected || o.ws == nil || len(audioData) == 0 {
		return
	}

	encodedAudio := base64.StdEncoding.EncodeToString(audioData)

	// Use the correct event format
	msg := OpenAIMessage{
		Type:  "input_audio_buffer.append",
		Audio: encodedAudio,
	}

	if err := o.ws.WriteJSON(msg); err != nil {
		slog.Error("Error sending to OpenAI", "userID", o.userID, "error", err)
	} else {
		slog.Debug("Sent audio to OpenAI", "userID", o.userID, "bytes", len(audioData))
	}
}

// handleResponses processes incoming messages from OpenAI
func (o *OpenAIRealtime) handleResponses() {
	for {
		o.mu.RLock()
		connected := o.connected
		o.mu.RUnlock()

		if !connected {
			break
		}

		var response map[string]interface{}
		if err := o.ws.ReadJSON(&response); err != nil {
			slog.Error("OpenAI connection closed", "userID", o.userID, "error", err)
			o.mu.Lock()
			o.connected = false
			o.mu.Unlock()
			return
		}

		// Handle different event types
		eventType, ok := response["type"].(string)
		if !ok {
			continue
		}

		switch eventType {
		case "session.created":
			o.sendDataChannelMessage("OpenAI:session created", "session created")
			slog.Info("OpenAI session created", "userID", o.userID)

		case "session.updated":
			o.sendDataChannelMessage("OpenAI:session updated", "session updated")
			slog.Info("OpenAI session updated", "userID", o.userID)

		case "response.audio.delta":
			// This is the key event for receiving audio chunks
			if delta, ok := response["delta"].(string); ok && delta != "" {
				o.sendDataChannelMessage("OpenAI:audio chunks recieved from llm", "delta")
				slog.Info("Received audio delta from OpenAI", "userID", o.userID, "chars", len(delta))
				o.sendAudioToClient(delta)
			}

		case "response.audio.done":
			o.sendDataChannelMessage("OpenAI:complete audio response from llm", "done")
			slog.Info("OpenAI audio response completed", "userID", o.userID)

		case "response.audio_transcript.delta":
			if delta, ok := response["delta"].(string); ok {
				slog.Info("OpenAI audio transcript delta", "userID", o.userID, "text", delta)
			}

		case "response.audio_transcript.done":
			if transcript, ok := response["transcript"].(string); ok {
				slog.Info("OpenAI audio transcript done", "userID", o.userID, "text", transcript)
			}

		case "input_audio_buffer.speech_started":
			o.sendDataChannelMessage("OpenAI:speech started", "speech started")
			slog.Info("OpenAI detected speech start", "userID", o.userID)

		case "input_audio_buffer.speech_stopped":
			o.sendDataChannelMessage("OpenAI:speech stopped", "speech stopped")
			slog.Info("OpenAI detected speech stop", "userID", o.userID)

		case "input_audio_buffer.committed":
			o.sendDataChannelMessage("OpenAI:audio buffer committed", "audio buffer committed")
			slog.Info("OpenAI audio buffer committed", "userID", o.userID)

		case "conversation.item.created":
			o.sendDataChannelMessage("OpenAI:conversation item created", "conversation item created")
			slog.Info("OpenAI conversation item created", "userID", o.userID)

		case "conversation.item.input_audio_transcription.completed":
			o.sendDataChannelMessage("OpenAI:input transcription completed", "input transcription completed")
			if transcript, ok := response["transcript"].(string); ok {
				slog.Info("OpenAI input transcription", "userID", o.userID, "text", transcript)
			}

		case "response.created":
			o.sendDataChannelMessage("OpenAI:response created", "response created")
			slog.Info("OpenAI response created", "userID", o.userID)

		case "response.output_item.added":
			slog.Info("OpenAI response output item added", "userID", o.userID)

		case "response.content_part.added":
			slog.Info("OpenAI response content part added", "userID", o.userID)

		case "response.done":
			o.sendDataChannelMessage("OpenAI:response completed", "response completed")

			// Convert the response map to JSON bytes first, then unmarshal to struct
			responseBytes, err := json.Marshal(response)
			if err != nil {
				slog.Error("Failed to marshal response", "error", err)
				break
			}

			var responseDone ResponseDoneEvent
			if err := json.Unmarshal(responseBytes, &responseDone); err != nil {
				slog.Error("Failed to unmarshal response.done event", "error", err)
				break
			}

			// Now you have type-safe access to usage data
			usage := responseDone.Response.Usage

			// Send basic usage info
			usageMsg := fmt.Sprintf("Usage - Total: %d, Input: %d, Output: %d",
				usage.TotalTokens, usage.InputTokens, usage.OutputTokens)
			o.sendDataChannelMessage("OpenAI:usage", usageMsg)

			// Send detailed token breakdown
			inputDetails := usage.InputTokenDetails
			detailMsg := fmt.Sprintf("Input Details - Text: %d, Audio: %d, Cached: %d",
				inputDetails.TextTokens, inputDetails.AudioTokens, inputDetails.CachedTokens)
			o.sendDataChannelMessage("OpenAI:input_details", detailMsg)

			outputDetails := usage.OutputTokenDetails
			outputMsg := fmt.Sprintf("Output Details - Text: %d, Audio: %d",
				outputDetails.TextTokens, outputDetails.AudioTokens)
			o.sendDataChannelMessage("OpenAI:output_details", outputMsg)

			slog.Info("OpenAI response completed", "userID", o.userID)

		case "rate_limits.updated":
			slog.Debug("OpenAI rate limits updated", "userID", o.userID)

		case "error":
			if errorData, ok := response["error"]; ok {
				slog.Error("OpenAI API error", "userID", o.userID, "error", errorData)
			}

		default:
			slog.Debug("Unhandled OpenAI event", "userID", o.userID, "type", eventType)
		}
	}
}

// sendAudioToClient sends processed audio to the WebRTC outbound track
func (o *OpenAIRealtime) sendAudioToClient(base64Audio string) {
	if o.outboundTrack == nil {
		return
	}

	// Decode base64 audio from OpenAI (24kHz PCM16)
	pcmData, err := base64.StdEncoding.DecodeString(base64Audio)
	if err != nil || len(pcmData) == 0 {
		slog.Error("Failed to decode audio", "userID", o.userID, "error", err)
		return
	}

	// Convert raw bytes to 16-bit audio samples (little-endian)
	pcmSamples := make([]int16, len(pcmData)/2)
	for i := 0; i < len(pcmSamples); i++ {
		pcmSamples[i] = int16(binary.LittleEndian.Uint16(pcmData[i*2:]))
	}

	// Upsample from 24kHz to 48kHz for WebRTC
	pcm48kHz := o.upsample24to48(pcmSamples)

	// Encode PCM to Opus in 20ms frames (960 samples at 48kHz)
	const frameSize = 960 // 20ms at 48kHz
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

		// Encode to Opus
		opusData := make([]byte, 4000)
		n, err := o.opusEncoder.Encode(frame, opusData)
		if err != nil {
			slog.Error("Opus encoding error", "userID", o.userID, "error", err)
			continue
		}

		// Create RTP packet with proper timing
		rtpPacket := &rtp.Packet{
			Header: rtp.Header{
				Version:        2,
				PayloadType:    111, // Opus
				SequenceNumber: o.sequenceNumber,
				Timestamp:      o.timestamp,
				SSRC:           1234,
			},
			Payload: opusData[:n],
		}
		o.sequenceNumber++
		o.timestamp += 960 // 20ms worth of samples at 48kHz

		if err := o.outboundTrack.WriteRTP(rtpPacket); err != nil {
			slog.Error("Error sending RTP packet", "userID", o.userID, "error", err)
			break
		}
	}

	slog.Debug("Sent audio to client", "userID", o.userID, "bytes", len(pcmData))
}

// Close terminates the OpenAI connection
func (o *OpenAIRealtime) Close() error {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.connected = false

	if o.ws != nil {
		err := o.ws.Close()
		o.ws = nil
		slog.Info("Closed OpenAI WebSocket connection", "userID", o.userID)
		return err
	}

	return nil
}

// IsConnected returns the connection status
func (o *OpenAIRealtime) IsConnected() bool {
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.connected
}

// Helper functions

// downsample48to24 converts 48kHz PCM to 24kHz (2:1 ratio)
func (o *OpenAIRealtime) downsample48to24(input []int16) []int16 {
	output := make([]int16, len(input)/2)
	for i := 0; i < len(output); i++ {
		output[i] = input[i*2] // Simple decimation
	}
	return output
}

// upsample24to48 converts 24kHz PCM to 48kHz (1:2 ratio)
func (o *OpenAIRealtime) upsample24to48(input []int16) []int16 {
	output := make([]int16, len(input)*2)
	for i := 0; i < len(input); i++ {
		output[i*2] = input[i] // Original sample
		if i*2+1 < len(output) {
			output[i*2+1] = input[i] // Duplicate sample
		}
	}
	return output
}
