// service/datachannel.go

package service

import (
	"encoding/json"
	"log/slog"

	"github.com/pion/webrtc/v3"
)

// DataChannelManager handles WebRTC data channel communication
type DataChannelManager struct {
	userID      string
	dataChannel *webrtc.DataChannel
	service     *RealtimeService
}

// DataChannelMessage represents messages sent through data channel
type DataChannelMessage struct {
	Type  string      `json:"type"`
	Model string      `json:"model,omitempty"` // Only used for switch_model
	Data  interface{} `json:"data,omitempty"`  // For structured payloads
}

// NewDataChannelManager creates a new data channel manager
func NewDataChannelManager(userID string, dataChannel *webrtc.DataChannel, service *RealtimeService) *DataChannelManager {
	dcm := &DataChannelManager{
		userID:      userID,
		dataChannel: dataChannel,
		service:     service,
	}

	dcm.setupDataChannel()
	return dcm
}

// setupDataChannel configures the data channel event handlers
func (dcm *DataChannelManager) setupDataChannel() {
	dcm.dataChannel.OnOpen(func() {
		slog.Info("Data channel opened", "userID", dcm.userID)
	})

	dcm.dataChannel.OnMessage(func(msg webrtc.DataChannelMessage) {
		var message DataChannelMessage
		if err := json.Unmarshal(msg.Data, &message); err != nil {
			slog.Error("Failed to parse data channel message", "userID", dcm.userID, "error", err)
			return
		}

		slog.Info("Received data channel message", "userID", dcm.userID, "type", message.Type)

		switch message.Type {
		case "disconnect":
			dcm.handleDisconnect()
		case "switch_model":
			dcm.handleSwitchModel(message.Model)
		default:
			slog.Warn("Unknown message type", "userID", dcm.userID, "type", message.Type)
		}
	})

	dcm.dataChannel.OnClose(func() {
		slog.Info("Data channel closed", "userID", dcm.userID)
	})
}

// handleDisconnect handles disconnect requests
func (dcm *DataChannelManager) handleDisconnect() {
	slog.Info("Disconnect requested", "userID", dcm.userID)

	go func() {
		err := dcm.service.Cleanup(dcm.userID)
		if err != nil {
			slog.Error("Failed to cleanup", "userID", dcm.userID, "error", err)
			dcm.sendMessage("error", "failed to cleanup", nil)
		}
	}()
	dcm.sendMessage("disconnected", "", nil)
}

// handleSwitchModel handles model switching
func (dcm *DataChannelManager) handleSwitchModel(model string) {
	if model == "" {
		slog.Error("Model parameter missing", "userID", dcm.userID)
		dcm.sendMessage("error", "model parameter required", nil)
		return
	}

	if model != "openai" && model != "gemini" {
		slog.Error("Unsupported model", "userID", dcm.userID, "model", model)
		dcm.sendMessage("error", "unsupported model: "+model, nil)
		return
	}

	slog.Info("Model switch requested", "userID", dcm.userID, "model", model)

	userConn := userConnections[dcm.userID]
	if userConn == nil {
		slog.Error("User connection not found", "userID", dcm.userID)
		return
	}

	// Close current AI connections
	if userConn.geminiRealtime != nil {
		userConn.geminiRealtime.Close()
		userConn.geminiRealtime = nil
	}
	if userConn.openaiRealtime != nil {
		userConn.openaiRealtime.Close()
		userConn.openaiRealtime = nil
	}

	// Connect to new model
	go func() {
		var err error
		if model == "gemini" {
			err = dcm.service.connectToGemini(dcm.userID)
		} else {
			err = dcm.service.connectToOpenai(dcm.userID)
		}

		if err != nil {
			slog.Error("Failed to switch model", "userID", dcm.userID, "model", model, "error", err)
			dcm.sendMessage("error", "failed to switch to "+model, nil)
		} else {
			dcm.sendMessage("model_switched", model, nil)
		}
	}()
}

// sendMessage sends a message to the browser
func (dcm *DataChannelManager) sendMessage(messageType string, model string, data interface{}) {
	dcm.sendMessageWithData(messageType, model, nil)
}

// sendMessageWithData sends a message to the browser with structured data
func (dcm *DataChannelManager) sendMessageWithData(messageType string, model string, data interface{}) {
	msg := DataChannelMessage{
		Type:  messageType,
		Model: model,
		Data:  data,
	}

	msgBytes, err := json.Marshal(msg)
	if err != nil {
		slog.Error("Failed to marshal message", "userID", dcm.userID, "error", err)
		return
	}

	if err := dcm.dataChannel.Send(msgBytes); err != nil {
		slog.Error("Failed to send data channel message", "userID", dcm.userID, "error", err)
	}
}

// Close closes the data channel
func (dcm *DataChannelManager) Close() error {
	if dcm.dataChannel != nil {
		return dcm.dataChannel.Close()
	}
	return nil
}
