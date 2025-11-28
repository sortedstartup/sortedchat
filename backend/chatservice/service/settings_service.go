package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"sortedstartup/chatservice/dao"
	"sortedstartup/chatservice/events"
	pb "sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/queue"
	settings "sortedstartup/chatservice/settings"
)

type SettingService struct {
	dao   dao.SettingsDAO
	queue queue.Queue
}

func NewSettingService(queue queue.Queue, daoFactory dao.DAOFactory) *SettingService {
	slog.Info("settings_service:NewSettingService")
	settingsDAO, err := daoFactory.CreateSettingsDAO()
	if err != nil {
		slog.Error("settings_service:NewSettingService, failed to create settings DAO", "error", err)
		return nil
	}
	return &SettingService{dao: settingsDAO, queue: queue}
}

func (s *SettingService) Init() {
	slog.Info("settings_service:Init", "settingService", s)
	// since right now the Setting is in chatservice so chatservice handles migrations
	isFirstBoot, err := s.IsFirstBoot()
	if err != nil {
		slog.Error("settings_service:Init", "step", "failed to check if this is first boot", "error", err)
		return
	}

	if isFirstBoot {
		// Save default settings but DON'T mark onboarding as complete
		// User must complete onboarding wizard to set is_first_boot = 1
		s.setSettingWithoutCompletingOnboarding(settings.DefaultSettings.ToProto())
	}

	// Note: FirstBootComplete() is now called only after onboarding wizard completion
}

func (s *SettingService) FirstBootComplete() {
	slog.Info("settings_service:FirstBootComplete", "settingService", s)
	err := s.dao.SetSettingValue("is_first_boot", "1")
	if err != nil {
		slog.Error("settings_service:FirstBootComplete", "message", "failed to set is_first_boot setting", "error", err)
	}
}

func (s *SettingService) GetSetting(ctx context.Context) (*pb.Settings, error) {
	slog.Info("settings_service:GetSetting", "settingService", s)
	settingsString, err := s.dao.GetSettingValue("settings")
	if err != nil {
		slog.Error("settings_service:GetSetting", "step", "failed to get settings", "error", err)
		return nil, fmt.Errorf("failed to get settings")
	}

	//json decode the settings
	var settingsObj settings.Settings
	err = json.Unmarshal([]byte(settingsString), &settingsObj)
	if err != nil {
		slog.Error("settings_service:GetSetting", "step", "failed to unmarshal settings", "error", err)
		return nil, fmt.Errorf("failed to get settings")
	}

	return settingsObj.ToProto(), nil
}

// saveSettings is the internal implementation for saving settings
// If completeOnboarding is true, sets is_first_boot = 1
func (s *SettingService) saveSettings(settingsProto *pb.Settings, completeOnboarding bool) error {
	slog.Info("settings_service:saveSettings", "settingService", s, "completeOnboarding", completeOnboarding)
	// Load existing settings from DB to support merge behavior
	existingSettingsStr, err := s.dao.GetSettingValue("settings")
	if err != nil {
		slog.Error("settings_service:saveSettings", "step", "failed to load existing settings for merge", "error", err)
		// Continue with empty existing settings on error; we'll still write incoming
	}

	var existing settings.Settings
	if existingSettingsStr != "" {
		if err := json.Unmarshal([]byte(existingSettingsStr), &existing); err != nil {
			slog.Error("settings_service:saveSettings", "step", "failed to unmarshal existing settings", "error", err)
		}
	}

	// Build incoming settings from proto
	incoming := settings.FromProto(settingsProto)

	// Merge: if incoming fields are empty strings, retain existing values
	if incoming.OpenAIAPIKey == "" {
		incoming.OpenAIAPIKey = existing.OpenAIAPIKey
	}
	if incoming.OpenAIAPIURL == "" {
		incoming.OpenAIAPIURL = existing.OpenAIAPIURL
	}
	if incoming.OllamaURL == "" {
		incoming.OllamaURL = existing.OllamaURL
	}

	settingsJSON, err := json.Marshal(incoming)
	if err != nil {
		slog.Error("settings_service:saveSettings", "step", "failed to set settings", "error", err)
		return fmt.Errorf("failed to set settings")
	}

	err = s.dao.SetSettingValue("settings", string(settingsJSON))
	if err != nil {
		slog.Error("settings_service:saveSettings", "step", "failed to set settings", "error", err)
		return fmt.Errorf("failed to set settings")
	}

	// Optionally mark onboarding as complete
	if completeOnboarding {
		err = s.dao.SetSettingValue("is_first_boot", "1")
		if err != nil {
			slog.Error("settings_service:saveSettings", "step", "failed to set is_first_boot", "error", err)
			// Don't fail the whole operation if this fails, just log it
		}
	}

	slog.Info("publishing settings change event", "event", events.SETTINGS_CHANGED_EVENT)
	// publish an event, any subscriber now need to reload settings from the database
	s.queue.Publish(context.Background(), events.SETTINGS_CHANGED_EVENT, []byte(""))

	return nil
}

// setSettingWithoutCompletingOnboarding saves settings without marking onboarding as complete
func (s *SettingService) setSettingWithoutCompletingOnboarding(settingsProto *pb.Settings) error {
	return s.saveSettings(settingsProto, false)
}

// SetSetting saves settings and marks onboarding as complete
func (s *SettingService) SetSetting(ctx context.Context, settingsProto *pb.Settings) error {
	return s.saveSettings(settingsProto, true)
}

// IsFirstBoot checks if this is the first boot by looking for the 'is_first_boot' setting
// Returns true if the setting doesn't exist or is 0, false otherwise
// Returns an error if there's a database error (except for sql.ErrNoRows)
func (s *SettingService) IsFirstBoot() (bool, error) {
	value, err := s.dao.GetSettingValue("is_first_boot")
	if err != nil {
		// If the setting doesn't exist, consider it first boot
		if err == sql.ErrNoRows {
			return true, nil
		}
		// For other database errors, return the error
		return false, fmt.Errorf("error getting is_first_boot setting: %w", err)
	}

	// Try to parse the value as an integer
	intValue, err := strconv.Atoi(value)
	if err != nil {
		// If we can't parse it, consider it first boot
		log.Printf("Error parsing is_first_boot value '%s': %v", value, err)
		return true, nil
	}

	// Return true if value is 0, false otherwise
	return intValue == 0, nil
}

func (s *SettingService) TestConnection(ctx context.Context, req *pb.TestConnectionRequest) (*pb.TestConnectionResponse, error) {
	slog.Info("settings_service:TestConnection", "url", req.Url, "type", req.ConnectionType)

	client := &http.Client{Timeout: 10 * time.Second}
	var resp *http.Response
	var err error
	var serviceName string

	// Prepare HTTP request
	httpReq, reqErr := http.NewRequest("HEAD", req.Url, nil)
	if reqErr != nil {
		slog.Error("settings_service:TestConnection", "step", "failed to create HTTP request", "error", reqErr, "url", req.Url)
		return &pb.TestConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Invalid URL: %v", reqErr),
		}, nil
	}

	// Determine service name
	switch req.ConnectionType {
	case pb.ConnectionType_OLLAMA:
		serviceName = "Ollama"
	case pb.ConnectionType_OPENAI:
		serviceName = "OpenAI API"
	default:
		return &pb.TestConnectionResponse{
			Success: false,
			Message: "Unsupported connection type",
		}, nil
	}

	// Send request
	resp, err = client.Do(httpReq)
	if err != nil {
		return &pb.TestConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("Failed to connect to %s: %v", serviceName, err),
		}, nil
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")

	// Handle Ollama
	if req.ConnectionType == pb.ConnectionType_OLLAMA {
		if resp.StatusCode == 200 {
			return &pb.TestConnectionResponse{
				Success: true,
				Message: fmt.Sprintf("%s connection successful", serviceName),
			}, nil
		}
		return &pb.TestConnectionResponse{
			Success: false,
			Message: fmt.Sprintf("%s returned status %d", serviceName, resp.StatusCode),
		}, nil
	}

	// Handle OpenAI
	if req.ConnectionType == pb.ConnectionType_OPENAI {
		switch resp.StatusCode {
		case 200:
			return &pb.TestConnectionResponse{
				Success: true,
				Message: fmt.Sprintf("%s connection successful", serviceName),
			}, nil
		case 404:
			if strings.Contains(contentType, "application/json") {
				// 404 from OpenAI backend — endpoint reachable
				return &pb.TestConnectionResponse{
					Success: true,
					Message: fmt.Sprintf("%s connection successful", serviceName),
				}, nil
			}
			// 404 from Cloudflare / invalid URL
			return &pb.TestConnectionResponse{
				Success: false,
				Message: fmt.Sprintf("%s endpoint not found (invalid URL)", serviceName),
			}, nil
		default:
			if resp.StatusCode >= 500 {
				return &pb.TestConnectionResponse{
					Success: false,
					Message: fmt.Sprintf("%s server error: %d", serviceName, resp.StatusCode),
				}, nil
			}
			return &pb.TestConnectionResponse{
				Success: false,
				Message: fmt.Sprintf("%s returned status %d", serviceName, resp.StatusCode),
			}, nil
		}
	}

	return &pb.TestConnectionResponse{
		Success: false,
		Message: fmt.Sprintf("Unknown error connecting to %s", serviceName),
	}, nil
}
