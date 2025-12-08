package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"sortedstartup/chatservice/dao"
	"sortedstartup/chatservice/events"
	"sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/queue"
	"sync"
)

var SQLITE_DB_URL = "db.sqlite"

/*
- TODO: To think : settings have to be app level and then broken down to the service level
*/
type Settings struct {
	OpenAIAPIKey string `koanf:"openai_api_key" json:"openai_api_key"`
	GeminiAPIKey string `koanf:"gemini_api_key" json:"gemini_api_key"`
	ClaudeAPIKey string `koanf:"claude_api_key" json:"claude_api_key"`
	ClaudeAPIUrl string `koanf:"claude_api_url" json:"claude_api_url"`
	GeminiAPIUrl string `koanf:"gemini_api_url" json:"gemini_api_url"`
	OpenaiAPIUrl string `koanf:"openai_api_url" json:"openai_api_url"`
	OllamaURL    string `koanf:"ollama_url" json:"ollama_url"`
}

var DefaultSettings = &Settings{
	OpenAIAPIKey: "",
	GeminiAPIKey: "",
	ClaudeAPIKey: "",
	ClaudeAPIUrl: "",
	GeminiAPIUrl: "",
	OpenaiAPIUrl: "",
	OllamaURL:    "",
}

func (s *Settings) ToProto() *proto.Settings {
	return &proto.Settings{
		OPENAI_API_KEY: s.OpenAIAPIKey,
		GEMINI_API_KEY: s.GeminiAPIKey,
		CLAUDE_API_KEY: s.ClaudeAPIKey,
		CLAUDE_API_URL: s.ClaudeAPIUrl,
		GEMINI_API_URL: s.GeminiAPIUrl,
		OPENAI_API_URL: s.OpenaiAPIUrl,
		OLLAMA_URL:     s.OllamaURL,
	}
}

func FromProto(protoSettings *proto.Settings) *Settings {
	return &Settings{
		OpenAIAPIKey: protoSettings.OPENAI_API_KEY,
		GeminiAPIKey: protoSettings.GEMINI_API_KEY,
		ClaudeAPIKey: protoSettings.CLAUDE_API_KEY,
		ClaudeAPIUrl: protoSettings.CLAUDE_API_URL,
		GeminiAPIUrl: protoSettings.GEMINI_API_URL,
		OpenaiAPIUrl: protoSettings.OPENAI_API_URL,
		OllamaURL:    protoSettings.OLLAMA_URL,
	}
}

// Application should use settings from here, not directly from the database
// This monitors the database for changes and reloads the settings
type SettingsManager struct {
	settings *Settings
	mu       sync.RWMutex
	queue    queue.Queue
	dao      dao.SettingsDAO
}

func NewSettingsManager(queue queue.Queue, daoFactory dao.DAOFactory) *SettingsManager {
	slog.Debug("settings:NewSettingsManager")
	settingsDAO, err := daoFactory.CreateSettingsDAO()
	if err != nil {
		log.Fatalf("Failed to create settings DAO: %v", err)
	}

	cm := &SettingsManager{
		settings: &Settings{},
		queue:    queue,
		dao:      settingsDAO,
	}

	cm.StartSettingsChangedSubscriber()
	slog.Info("settings:NewSettingsManager", "settingsManager", cm)
	return cm
}

// Legacy constructor for backward compatibility
func NewSettingsManagerWithSQLite(queue queue.Queue) *SettingsManager {
	// Create a simple SQLite settings DAO directly
	settingsDAO := dao.NewSQLiteSettingsDAO(SQLITE_DB_URL)

	cm := &SettingsManager{
		settings: &Settings{},
		queue:    queue,
		dao:      settingsDAO,
	}

	cm.StartSettingsChangedSubscriber()
	return cm
}

func (cm *SettingsManager) LoadSettingsFromProto(protoSettings *proto.Settings) error {

	cm.settings = &Settings{
		OpenAIAPIKey: protoSettings.OPENAI_API_KEY,
		GeminiAPIKey: protoSettings.GEMINI_API_KEY,
		ClaudeAPIKey: protoSettings.CLAUDE_API_KEY,
		ClaudeAPIUrl: protoSettings.CLAUDE_API_URL,
		GeminiAPIUrl: protoSettings.GEMINI_API_URL,
		OpenaiAPIUrl: protoSettings.OPENAI_API_URL,
		OllamaURL:    protoSettings.OLLAMA_URL,
	}

	cm.LoadSettings(cm.settings)
	return nil
}

func (cm *SettingsManager) LoadSettings(settings_ *Settings) error {

	slog.Info("settings:LoadSettings")

	// The lock prevents race conditions when loading settings from the database
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Create new config struct
	// clone the settings_, Later this will be replaced by koanf
	newSettings := *settings_

	// Replace the config atomically
	cm.settings = &newSettings
	return nil
}

func (cm *SettingsManager) GetSettings() *Settings {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.settings
}

func (s *SettingsManager) StartSettingsChangedSubscriber() {
	slog.Info("settings:StartSettingsChangedSubscriber")
	go func() {
		sub, err := s.queue.Subscribe(context.Background(), events.SETTINGS_CHANGED_EVENT)
		if err != nil {
			slog.Error("settings:StartSettingsChangedSubscriber", "step", "failed to subscribe to settings changed event", "error", err)
			return
		}
		for msg := range sub {
			slog.Info("settings:StartSettingsChangedSubscriber", "message", events.SETTINGS_CHANGED_EVENT, "payload_bytes", len(msg.Data))
			// reload settings from the database
			slog.Info("settings:StartSettingsChangedSubscriber", "action", "Reloading settings from the database")
			s.LoadSettingsFromDB()

		}
	}()
}

func (s *SettingsManager) LoadSettingsFromDB() error {
	slog.Info("settings:LoadSettingsFromDB")
	settingsString, err := s.dao.GetSettingValue("settings")
	if err != nil {
		slog.Error("settings:LoadSettingsFromDB", "message", "failed to get settings value", "error", err)
		return fmt.Errorf("failed to get settings value")
	}

	//json decode the settings
	var settings Settings
	err = json.Unmarshal([]byte(settingsString), &settings)
	if err != nil {
		slog.Error("settings:LoadSettingsFromDB", "message", "failed to unmarshal settings", "error", err)
		return fmt.Errorf("failed to unmarshal settings")
	}

	return s.LoadSettings(&settings)
}
