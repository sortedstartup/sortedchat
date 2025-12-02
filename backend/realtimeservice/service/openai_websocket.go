package service

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
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
	cancel             context.CancelFunc
	service            *RealtimeService
}

// OpenAI message types based on the actual API response format
type OpenAIMessage struct {
	Type    string         `json:"type"`
	Session *OpenAISession `json:"session,omitempty"`
	Audio   string         `json:"audio,omitempty"`
}

// Session structure based on the actual API response format from your logs
type OpenAISession struct {
	Instructions            string               `json:"instructions,omitempty"`
	Voice                   string               `json:"voice,omitempty"`
	InputAudioFormat        string               `json:"input_audio_format,omitempty"`
	OutputAudioFormat       string               `json:"output_audio_format,omitempty"`
	TurnDetection           *OpenAITurnDetection `json:"turn_detection,omitempty"`
	MaxResponseOutputTokens interface{}          `json:"max_response_output_tokens,omitempty"`
	Modalities              []string             `json:"modalities,omitempty"`
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
func NewOpenAIRealtime(userID string, outboundTrack *webrtc.TrackLocalStaticRTP, dataChannelManager *DataChannelManager, service *RealtimeService) (*OpenAIRealtime, error) {

	apiKey := OPENAI_API_KEY
	if apiKey == "" {
		slog.Error("RealtimeService:openai_websocket:NewOpenAIRealtime", "message", "OPENAI_API_KEY environment variable is required")
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is required")
	}

	// // Initialize Opus encoder/decoder for 24kHz (OpenAI requirement)
	opusEncoder, err := opus.NewEncoder(48000, 1, opus.AppVoIP)
	if err != nil {
		slog.Error("RealtimeService:openai_websocket:NewOpenAIRealtime", "message", "failed to create Opus encoder", "error", err)
		return nil, fmt.Errorf("failed to create Opus encoder: %v", err)
	}

	opusDecoder, err := opus.NewDecoder(48000, 1)
	if err != nil {
		slog.Error("RealtimeService:openai_websocket:NewOpenAIRealtime", "message", "failed to create Opus decoder", "error", err)
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
		service:            service,
	}, nil
}

// SetDataChannelManager sets the data channel manager (used when it becomes available later)
func (o *OpenAIRealtime) SetDataChannelManager(dcm *DataChannelManager) {
	slog.Info("RealtimeService:openai_websocket:SetDataChannelManager")
	o.mu.Lock()
	defer o.mu.Unlock()
	o.dataChannelManager = dcm
}

// sendDataChannelMessage safely sends a message via data channel if available
func (o *OpenAIRealtime) sendDataChannelMessage(messageType string, model string, data interface{}) {
	slog.Info("RealtimeService:openai_websocket:sendDataChannelMessage")
	o.mu.RLock()
	dcm := o.dataChannelManager
	o.mu.RUnlock()

	if dcm != nil {
		dcm.sendMessageWithData(messageType, model, data)
	} else {
		slog.Debug("Data channel manager not available, skipping message", "userID", o.userID, "messageType", messageType)
	}
}

// Connect establishes WebSocket connection to OpenAI Realtime API
func (o *OpenAIRealtime) Connect(model string) error {
	slog.Info("Connecting to OpenAI Realtime API", "userID", o.userID)

	// Use the correct URL format
	wsURL := "wss://api.openai.com/v1/realtime?model=" + model

	headers := make(map[string][]string)
	headers["Authorization"] = []string{"Bearer " + o.apiKey}
	headers["OpenAI-Beta"] = []string{"realtime=v1"}

	var err error
	o.ws, _, err = websocket.DefaultDialer.Dial(wsURL, headers)
	if err != nil {
		slog.Error("Failed to connect to OpenAI", "userID", o.userID, "error", err)
		return err
	}

	if o.ws == nil {
		slog.Error("Failed to connect to OpenAI", "userID", o.userID)
		return fmt.Errorf("failed to connect to OpenAI")
	}

	slog.Info("Successfully connected to OpenAI", "userID", o.userID)

	o.mu.Lock()
	o.connected = true
	o.mu.Unlock()

	// Create context for this connection
	ctx, cancel := context.WithCancel(context.Background())
	o.cancel = cancel // Store cancel function

	// Start handling responses with context
	go func() {
		defer cancel() // Always cancel when goroutine exits

		err := o.handleResponses(ctx)
		if err != nil {
			slog.Error("handleResponses failed", "userID", o.userID, "error", err)
		}
	}()

	// Monitor context
	go func() {
		<-ctx.Done() // Wait for context cancellation-blockign
		slog.Info("Context cancelled - triggering service cleanup", "userID", o.userID)

		// Trigger full service cleanup
		if o.service != nil {
			o.service.Cleanup(o.userID)
		}
	}()

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
// converting opus/48000 to pcm/24000 and sending to OpenAI
func (o *OpenAIRealtime) HandleAudioTrack(track *webrtc.TrackRemote) {
	slog.Info("Starting audio track handling for OpenAI", "userID", o.userID)

	o.sendDataChannelMessage("Client_audio", "openai", nil) //custom event

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
			slog.Error("RealtimeService:openai_websocket:HandleAudioTrack", "message", "failed to unmarshal Opus data", "userID", o.userID, "error", err)
			continue
		}
		if opusData == nil {
		}
		// Decode Opus to PCM 20ms at 48kHz
		pcmData := make([]int16, 960) // 20ms at 48kHz
		n, err := o.opusDecoder.Decode(opusData, pcmData)
		if err != nil {
			slog.Error("RealtimeService:openai_websocket:HandleAudioTrack", "message", "failed to decode Opus data", "userID", o.userID, "error", err)
			continue
		}

		// Accumulate PCM data
		pcmBuffer = append(pcmBuffer, pcmData[:n]...)

		if len(pcmBuffer) >= 4800 {
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

	o.sendDataChannelMessage("sent_audio", "openai", nil) //custom event

	if err := o.ws.WriteJSON(msg); err != nil {
		slog.Error("Error sending to OpenAI", "userID", o.userID, "error", err)
	}
}

// handleResponses processes incoming messages from OpenAI
func (o *OpenAIRealtime) handleResponses(ctx context.Context) error {
	for {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			slog.Info("handleResponses cancelled by context", "userID", o.userID)
			return ctx.Err()
		default:
			// Continue processing
		}

		o.mu.RLock()
		connected := o.connected //connected to OpenAI status
		o.mu.RUnlock()

		if !connected {
			slog.Error("RealtimeService:openai_websocket:handleResponses", "message", "OpenAI connection closed", "userID", o.userID)
			return fmt.Errorf("OpenAI connection closed")
		}

		var response map[string]interface{}
		if err := o.ws.ReadJSON(&response); err != nil {
			slog.Error("OpenAI connection closed", "userID", o.userID, "error", err)
			o.mu.Lock()
			o.connected = false //set connected to false if connection is closed
			o.mu.Unlock()
			return err // This will exit function → defer cancel() → context cancelled → cleanup triggered
		}

		// Handle different event types
		eventType, ok := response["type"].(string)
		if !ok {
			continue
		}

		switch eventType {
		case "session.created":
			o.sendDataChannelMessage("OpenAI:session created", "session created", nil)
			slog.Info("OpenAI session created", "userID", o.userID)

		case "session.updated":
			o.sendDataChannelMessage("OpenAI:session updated", "session updated", nil)
			slog.Info("OpenAI session updated", "userID", o.userID)

		case "response.audio.delta":
			// This is the key event for receiving audio chunks
			if delta, ok := response["delta"].(string); ok && delta != "" {
				o.sendDataChannelMessage("OpenAI:audio chunks recieved from llm", "delta", nil)
				slog.Info("Received audio delta from OpenAI", "userID", o.userID, "chars", len(delta))
				o.sendAudioToClient(delta)
			}

		case "response.audio.done":
			o.sendDataChannelMessage("OpenAI:complete audio response from llm", "done", nil)
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
			o.sendDataChannelMessage("OpenAI:speech started", "speech started", nil)
			slog.Info("OpenAI detected speech start", "userID", o.userID)

		case "input_audio_buffer.speech_stopped":
			o.sendDataChannelMessage("OpenAI:speech stopped", "speech stopped", nil)
			slog.Info("OpenAI detected speech stop", "userID", o.userID)

		case "input_audio_buffer.committed":
			o.sendDataChannelMessage("OpenAI:audio buffer committed", "audio buffer committed", nil)
			slog.Info("OpenAI audio buffer committed", "userID", o.userID)

		case "conversation.item.created":
			o.sendDataChannelMessage("OpenAI:conversation item created", "conversation item created", nil)
			slog.Info("OpenAI conversation item created", "userID", o.userID)

		case "conversation.item.input_audio_transcription.completed":
			o.sendDataChannelMessage("OpenAI:input transcription completed", "input transcription completed", nil)
			if transcript, ok := response["transcript"].(string); ok {
				slog.Info("OpenAI input transcription", "userID", o.userID, "text", transcript)
			}

		case "response.created":
			o.sendDataChannelMessage("OpenAI:response created", "response created", nil)
			slog.Info("OpenAI response created", "userID", o.userID)

		case "response.output_item.added":
			slog.Info("OpenAI response output item added", "userID", o.userID)

		case "response.content_part.added":
			slog.Info("OpenAI response content part added", "userID", o.userID)

		case "response.done":
			o.sendDataChannelMessage("response_completed", "openai", nil)

			responseBytes, err := json.Marshal(response)
			if err != nil {
				slog.Error("Failed to marshal response", "error", err)
				return err
			}
			var responseDone ResponseDoneEvent
			if err := json.Unmarshal(responseBytes, &responseDone); err != nil {
				slog.Error("Failed to unmarshal response.done event", "openai", err)
				return err
			}

			// Now you have type-safe access to usage data
			usage := responseDone.Response.Usage

			// Send basic usage info as structured data
			o.sendDataChannelMessage("OpenAI:usage", "openai", map[string]interface{}{
				"total_tokens":  usage.TotalTokens,
				"input_tokens":  usage.InputTokens,
				"output_tokens": usage.OutputTokens,
			})

			// Send detailed token breakdown as structured data
			inputDetails := usage.InputTokenDetails
			o.sendDataChannelMessage("OpenAI:input_details", "openai", map[string]interface{}{
				"text_tokens":   inputDetails.TextTokens,
				"audio_tokens":  inputDetails.AudioTokens,
				"cached_tokens": inputDetails.CachedTokens,
			})

			outputDetails := usage.OutputTokenDetails
			o.sendDataChannelMessage("OpenAI:output_details", "openai", map[string]interface{}{
				"text_tokens":  outputDetails.TextTokens,
				"audio_tokens": outputDetails.AudioTokens,
			})

			slog.Info("OpenAI response completed", "userID", o.userID)

		case "rate_limits.updated":
			slog.Debug("OpenAI rate limits updated", "userID", o.userID)

		case "error":
			if errorData, ok := response["error"]; ok {
				slog.Error("OpenAI API error", "userID", o.userID, "error", errorData)
				return fmt.Errorf("OpenAI API error: %v", errorData)
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
		if frame == nil {
		}

		// Encode to Opus
		opusData := make([]byte, 4800)
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

	// Cancel context first to stop all goroutines
	if o.cancel != nil {
		o.cancel()
	}

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
	slog.Info("RealtimeService:openai_websocket:IsConnected")
	o.mu.RLock()
	defer o.mu.RUnlock()
	return o.connected
}

// Helper functions

// downsample48to24 converts 48kHz PCM to 24kHz (2:1 ratio)
// might need to change this logic and use github.com/zaf/resample for better quality
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
