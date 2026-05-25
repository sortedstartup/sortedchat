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

	"sortedstartup/chatservice/constants"
	"sortedstartup/chatservice/dao"
	"sortedstartup/chatservice/events"
	pb "sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/queue"
	settings "sortedstartup/chatservice/settings"

	"google.golang.org/protobuf/types/known/structpb"
)

type SettingService struct {
	dao             dao.SettingsDAO
	queue           queue.Queue
	settingsManager *settings.SettingsManager
}

func NewSettingService(queue queue.Queue, settingsManager *settings.SettingsManager, daoFactory dao.DAOFactory) *SettingService {
	slog.Debug("settings_service:NewSettingService")
	settingsDAO, err := daoFactory.CreateSettingsDAO()
	if err != nil {
		slog.Error("settings_service:NewSettingService, failed to create settings DAO", "error", err)
		return nil
	}
	return &SettingService{dao: settingsDAO, queue: queue, settingsManager: settingsManager}
}

func (s *SettingService) Init() {
	// since right now the Setting is in chatservice so chatservice handles migrations
	isFirstBoot, err := s.IsFirstBoot()
	if err != nil {
		slog.Error("settings_service:Init", "step", "failed to check if this is first boot", "error", err)
		return
	}

	// Check if local provider settings exist, if not seed them (for backward compatibility)
	localProviderSettings, err := s.GetProviderSetting(context.Background(), constants.LOCAL_PROVIDER)
	if err != nil && err != sql.ErrNoRows {
		slog.Error("settings_service:Init", "step", "failed to check local provider settings", "error", err)
	} else if localProviderSettings == nil || localProviderSettings.ApiUrl == "" {
		slog.Info("settings_service:Init", "step", "seeding default local provider settings")
		defaultLocalSettings := &pb.ProviderSettings{
			ApiUrl:    constants.LOCAL_LLAMA_PROXY_BASE_URL + "/v1/chat/completions",
			ApiKey:    "NOT_NEEDED",
			IsEnabled: true,
		}
		if err := s.SetProviderSetting(context.Background(), constants.LOCAL_PROVIDER, defaultLocalSettings, false); err != nil {
			slog.Error("settings_service:Init", "step", "failed to seed local provider settings", "error", err)
		}
	}

	if isFirstBoot {
		// Save default settings but DON'T mark onboarding as complete
		// User must complete onboarding wizard to set is_first_boot = 1

		// Convert DefaultSettings to map for structpb
		defaultSettingsMap := map[string]interface{}{
			"OPENAI_API_KEY": settings.DefaultSettings.OpenAIAPIKey,
			"GEMINI_API_KEY": settings.DefaultSettings.GeminiAPIKey,
			"CLAUDE_API_KEY": settings.DefaultSettings.ClaudeAPIKey,
			"CLAUDE_API_URL": settings.DefaultSettings.ClaudeAPIUrl,
			"GEMINI_API_URL": settings.DefaultSettings.GeminiAPIUrl,
			"OPENAI_API_URL": settings.DefaultSettings.OpenaiAPIUrl,
			"OLLAMA_URL":     settings.DefaultSettings.OllamaURL,
		}

		st, err := structpb.NewStruct(defaultSettingsMap)
		if err != nil {
			slog.Error("settings_service:Init", "error", "failed to create default settings struct", "err", err)
			return
		}

		s.setSettingWithoutCompletingOnboarding("settings", st)
	}

	webSearchDefaults, err := json.Marshal(webSearchSettings{
		APIURL: settings.DEFAULT_BRAVE_SEARCH_API_URL,
		APIKey: "",
	})
	if err != nil {
		slog.Error("settings_service:Init", "step", "failed to marshal default web search settings", "error", err)
	} else {
		s.ensureDefaultSetting(settings.WEBSEARCH_SETTINGS_KEY, string(webSearchDefaults))
	}
	defaultPromptStruct, err := structpb.NewStruct(map[string]interface{}{"value": settings.DEFAULT_CHAT_PROMPT})
	if err != nil {
		slog.Error("settings_service:Init", "step", "failed to create default chat prompt struct", "error", err)
	} else {
		bytes, marshalErr := json.Marshal(defaultPromptStruct.AsMap())
		if marshalErr != nil {
			slog.Error("settings_service:Init", "step", "failed to marshal default chat prompt", "error", marshalErr)
		} else {
			s.ensureDefaultSetting(settings.CHAT_DEFAULT_PROMPT_KEY, string(bytes))
		}
	}

	// Note: FirstBootComplete() is now called only after onboarding wizard completion
}

func (s *SettingService) ensureDefaultSetting(name string, defaultValue string) {
	value, err := s.dao.GetSettingValue(name)
	if err != nil && err != sql.ErrNoRows {
		slog.Error("settings_service:ensureDefaultSetting", "step", "failed to load setting", "name", name, "error", err)
		return
	}

	if strings.TrimSpace(value) != "" {
		return
	}

	if err := s.dao.SetSettingValue(name, defaultValue); err != nil {
		slog.Error("settings_service:ensureDefaultSetting", "step", "failed to seed setting", "name", name, "error", err)
	}
}

func (s *SettingService) FirstBootComplete() {
	err := s.dao.SetSettingValue("is_first_boot", "1")
	if err != nil {
		slog.Error("settings_service:FirstBootComplete", "message", "failed to set is_first_boot setting", "error", err)
	}
}

func (s *SettingService) GetSetting(ctx context.Context, name string) (*structpb.Struct, error) {
	switch name {
	case settings.WEBSEARCH_SETTINGS_KEY:
		webSearchSettings := s.settingsManager.GetWebSearchSettings()
		return structpb.NewStruct(map[string]interface{}{
			"apiUrl": webSearchSettings.APIURL,
			"apiKey": webSearchSettings.APIKey,
		})
	case settings.CHAT_DEFAULT_PROMPT_KEY:
		return structpb.NewStruct(map[string]interface{}{
			"value": s.settingsManager.GetChatDefaultPrompt(),
		})
	}

	settingsString, err := s.dao.GetSettingValue(name)
	if err != nil {
		if err == sql.ErrNoRows {
			return &structpb.Struct{}, nil
		}
		slog.Error("settings_service:GetSetting", "step", "failed to get settings", "error", err, "name", name)
		return nil, fmt.Errorf("failed to get settings")
	}

	if settingsString == "" {
		return &structpb.Struct{}, nil
	}

	var settingsMap map[string]interface{}
	err = json.Unmarshal([]byte(settingsString), &settingsMap)
	if err != nil {
		slog.Error("settings_service:GetSetting", "step", "failed to unmarshal settings", "error", err, "name", name)
		return nil, fmt.Errorf("failed to get settings")
	}

	return structpb.NewStruct(settingsMap)
}

// saveSettings is the internal implementation for saving settings
// If completeOnboarding is true, sets is_first_boot = 1
func (s *SettingService) saveSettings(name string, settingsStruct *structpb.Struct, completeOnboarding bool) error {
	// Load existing settings from DB to support merge behavior
	existingSettingsStr, err := s.dao.GetSettingValue(name)
	if err != nil && err != sql.ErrNoRows {
		slog.Error("settings_service:saveSettings", "step", "failed to load existing settings for merge", "error", err)
		// Continue with empty existing settings on error; we'll still write incoming
	}

	existingMap := make(map[string]interface{})
	if existingSettingsStr != "" {
		if err := json.Unmarshal([]byte(existingSettingsStr), &existingMap); err != nil {
			slog.Error("settings_service:saveSettings", "step", "failed to unmarshal existing settings", "error", err)
		}
	}

	incomingMap := settingsStruct.AsMap()

	// Merge: overwrite existing with incoming
	for k, v := range incomingMap {
		existingMap[k] = v
	}

	settingsJSON, err := json.Marshal(existingMap)
	if err != nil {
		slog.Error("settings_service:saveSettings", "step", "failed to marshal settings", "error", err)
		return fmt.Errorf("failed to set settings")
	}

	err = s.dao.SetSettingValue(name, string(settingsJSON))
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

	switch name {
	case settings.WEBSEARCH_SETTINGS_KEY:
		if err := s.settingsManager.LoadWebSearchSettingsFromDB(); err != nil {
			return err
		}
	case settings.CHAT_DEFAULT_PROMPT_KEY:
		if err := s.settingsManager.LoadChatDefaultPromptFromDB(); err != nil {
			return err
		}
	}

	return nil
}

// setSettingWithoutCompletingOnboarding saves settings without marking onboarding as complete
func (s *SettingService) setSettingWithoutCompletingOnboarding(name string, settingsStruct *structpb.Struct) error {
	return s.saveSettings(name, settingsStruct, false)
}

// SetSetting saves settings and marks onboarding as complete
func (s *SettingService) SetSetting(ctx context.Context, name string, settingsStruct *structpb.Struct) error {
	return s.saveSettings(name, settingsStruct, true)
}

func (s *SettingService) GetProviderSetting(ctx context.Context, name string) (*pb.ProviderSettings, error) {
	key := "provider." + name
	val, err := s.dao.GetSettingValue(key)
	if err != nil {
		if err == sql.ErrNoRows {
			return &pb.ProviderSettings{}, nil
		}
		return nil, err
	}

	if val == "" {
		return &pb.ProviderSettings{}, nil
	}

	var ps pb.ProviderSettings
	if err := json.Unmarshal([]byte(val), &ps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provider settings: %w", err)
	}
	return &ps, nil
}

func (s *SettingService) SetProviderSetting(ctx context.Context, name string, settings *pb.ProviderSettings, completeOnboarding bool) error {
	key := "provider." + name
	bytes, err := json.Marshal(settings)
	if err != nil {
		return fmt.Errorf("failed to marshal provider settings: %w", err)
	}

	if err := s.dao.SetSettingValue(key, string(bytes)); err != nil {
		return err
	}

	if completeOnboarding {
		err = s.dao.SetSettingValue("is_first_boot", "1")
		if err != nil {
			slog.Error("settings_service:SetProviderSetting", "step", "failed to set is_first_boot", "error", err)
		}
	}

	s.queue.Publish(context.Background(), events.SETTINGS_CHANGED_EVENT, []byte(""))
	return nil
}

func (s *SettingService) GetAllProviderSettings(ctx context.Context) (map[string]*pb.ProviderSettings, error) {
	settingsMap, err := s.dao.GetSettingsByPrefix("provider.")
	if err != nil {
		return nil, err
	}

	result := make(map[string]*pb.ProviderSettings)
	for k, v := range settingsMap {
		// k is like "provider.openai", we want just "openai"
		name := strings.TrimPrefix(k, "provider.")
		var ps pb.ProviderSettings
		if err := json.Unmarshal([]byte(v), &ps); err != nil {
			slog.Error("failed to unmarshal provider setting", "key", k, "error", err)
			continue
		}
		result[name] = &ps
	}
	return result, nil
}

func (s *SettingService) SetAllProviderSettings(ctx context.Context, settingsMap map[string]*pb.ProviderSettings, completeOnboarding bool) error {
	for name, ps := range settingsMap {
		if err := s.SetProviderSetting(ctx, name, ps, true); err != nil {
			return err
		}
	}
	return nil
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

type webSearchSettings struct {
	APIURL string `json:"apiUrl"`
	APIKey string `json:"apiKey"`
}

func (s *ChatService) getWebSearchSettings() (*webSearchSettings, error) {
	settings := s.settingsManager.GetWebSearchSettings()
	return &webSearchSettings{
		APIURL: settings.APIURL,
		APIKey: settings.APIKey,
	}, nil
}

func (s *ChatService) getChatDefaultPrompt() (string, error) {
	return s.settingsManager.GetChatDefaultPrompt(), nil
}
