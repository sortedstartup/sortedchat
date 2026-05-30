package settings

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"sortedstartup/chatservice/dao"
	"sortedstartup/chatservice/events"
	"sortedstartup/chatservice/proto"
	"sortedstartup/chatservice/queue"
	"strconv"
	"sync"
)

var SQLITE_DB_URL = "db.sqlite"

const (
	WEBSEARCH_SETTINGS_KEY       = "tool.websearch.brave"
	CHAT_DEFAULT_PROMPT_KEY      = "chat.default_system_prompt"
	DEFAULT_BRAVE_SEARCH_API_URL = "https://api.search.brave.com/res/v1/web/search"
	DEFAULT_BRAVE_SEARCH_COST    = "0"
	DEFAULT_CHAT_PROMPT          = `You are SortedChat’s default assistant.

Answer from your own knowledge and reasoning by default. Use web_search only when fresh, external, verifiable, or source-backed information is needed, such as news, prices, laws, schedules, product details, live data, recent updates, or explicit search requests.
When you search, ground the answer in the results and include relevant source URLs.
`
	CLOUDFLARE_SCRAPE_SETTINGS_KEY = "tool.scrape.cloudflare"
	DEFAULT_AGENCIC_MAX_TURNS      = 4
	DEFAULT_AGENCIC_MAX_TURNS_KEY  = "chat.agentic_max_turns"
)

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

type WebSearchSettings struct {
	APIURL string `json:"apiUrl"`
	APIKey string `json:"apiKey"`
	Cost   string `json:"cost"`
}

type CloudflareScrapeSettings struct {
	APIURL string `json:"apiUrl"`
	APIKey string `json:"apiKey"`
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
	settings               *Settings
	webSearchSettings      *WebSearchSettings
	chatDefaultPrompt      string
	defaultAgenticMaxTurns int
	scrapeSettings         *CloudflareScrapeSettings
	mu                     sync.RWMutex
	queue                  queue.Queue
	dao                    dao.SettingsDAO
}

func NewSettingsManager(queue queue.Queue, daoFactory dao.DAOFactory) *SettingsManager {
	slog.Debug("settings:NewSettingsManager")
	settingsDAO, err := daoFactory.CreateSettingsDAO()
	if err != nil {
		log.Fatalf("Failed to create settings DAO: %v", err)
	}

	cm := &SettingsManager{
		settings:               &Settings{},
		webSearchSettings:      &WebSearchSettings{APIURL: DEFAULT_BRAVE_SEARCH_API_URL, Cost: DEFAULT_BRAVE_SEARCH_COST},
		chatDefaultPrompt:      DEFAULT_CHAT_PROMPT,
		defaultAgenticMaxTurns: DEFAULT_AGENCIC_MAX_TURNS,
		scrapeSettings:         &CloudflareScrapeSettings{APIURL: "", APIKey: ""},
		queue:                  queue,
		dao:                    settingsDAO,
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
		settings:          &Settings{},
		webSearchSettings: &WebSearchSettings{APIURL: DEFAULT_BRAVE_SEARCH_API_URL, Cost: DEFAULT_BRAVE_SEARCH_COST},
		scrapeSettings:    &CloudflareScrapeSettings{APIURL: "", APIKey: ""},
		chatDefaultPrompt: DEFAULT_CHAT_PROMPT,
		queue:             queue,
		dao:               settingsDAO,
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

func (cm *SettingsManager) LoadWebSearchSettings(settings_ *WebSearchSettings) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	newSettings := *settings_
	if newSettings.APIURL == "" {
		newSettings.APIURL = DEFAULT_BRAVE_SEARCH_API_URL
	}
	if newSettings.Cost == "" {
		newSettings.Cost = DEFAULT_BRAVE_SEARCH_COST
	}
	cm.webSearchSettings = &newSettings
}

func (cm *SettingsManager) LoadScrapeSettings(settings_ *CloudflareScrapeSettings) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	newSettings := *settings_
	cm.scrapeSettings = &newSettings
}

func (cm *SettingsManager) LoadChatDefaultPrompt(prompt string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if prompt == "" {
		prompt = DEFAULT_CHAT_PROMPT
	}
	cm.chatDefaultPrompt = prompt
}

func (cm *SettingsManager) LoadAgenticMaxTurns(maxTurns int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if maxTurns <= 0 {
		maxTurns = DEFAULT_AGENCIC_MAX_TURNS
	}
	cm.defaultAgenticMaxTurns = maxTurns
}

func (cm *SettingsManager) GetSettings() *Settings {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.settings
}

func (cm *SettingsManager) GetWebSearchSettings() *WebSearchSettings {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	newSettings := *cm.webSearchSettings
	return &newSettings
}

func (cm *SettingsManager) GetAgenticMaxTurns() int {
	// This is a temporary function to get the agentic max turns setting, which will be used in the agentic chat service
	// In the future, this should be replaced by a more generic function to get any setting
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.defaultAgenticMaxTurns
}
func (cm *SettingsManager) GetCloudflareScrapeSettings() *CloudflareScrapeSettings {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	newSettings := *cm.scrapeSettings
	return &newSettings
}
func (cm *SettingsManager) GetChatDefaultPrompt() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.chatDefaultPrompt
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

	if err := s.LoadSettings(&settings); err != nil {
		return err
	}

	if err := s.LoadWebSearchSettingsFromDB(); err != nil {
		return err
	}

	if err := s.LoadScrapeSettingsFromDB(); err != nil {
		return err
	}

	if err := s.LoadChatDefaultPromptFromDB(); err != nil {
		return err
	}

	if err := s.LoadAgenticMaxTurnsFromDB(); err != nil {
		return err
	}

	return nil
}

func (s *SettingsManager) LoadWebSearchSettingsFromDB() error {
	value, err := s.dao.GetSettingValue(WEBSEARCH_SETTINGS_KEY)
	if err != nil {
		if err == sql.ErrNoRows {
			s.LoadWebSearchSettings(&WebSearchSettings{APIURL: DEFAULT_BRAVE_SEARCH_API_URL, Cost: DEFAULT_BRAVE_SEARCH_COST})
			return nil
		}
		slog.Error("settings:LoadWebSearchSettingsFromDB", "message", "failed to get settings value", "error", err)
		return fmt.Errorf("failed to get web search settings value")
	}

	if value == "" {
		s.LoadWebSearchSettings(&WebSearchSettings{APIURL: DEFAULT_BRAVE_SEARCH_API_URL, Cost: DEFAULT_BRAVE_SEARCH_COST})
		return nil
	}

	var webSearchSettings WebSearchSettings
	if err := json.Unmarshal([]byte(value), &webSearchSettings); err != nil {
		slog.Error("settings:LoadWebSearchSettingsFromDB", "message", "failed to unmarshal web search settings", "error", err)
		return fmt.Errorf("failed to unmarshal web search settings")
	}

	s.LoadWebSearchSettings(&webSearchSettings)
	return nil
}

func (s *SettingsManager) LoadScrapeSettingsFromDB() error {
	value, err := s.dao.GetSettingValue(CLOUDFLARE_SCRAPE_SETTINGS_KEY)
	if err != nil {
		if err == sql.ErrNoRows {
			s.LoadScrapeSettings(&CloudflareScrapeSettings{APIURL: "", APIKey: ""})
			return nil
		}
		slog.Error("settings:LoadScrapeSettingsFromDB", "message", "failed to get scrape settings value", "error", err)
		return fmt.Errorf("failed to get scrape settings value")
	}

	if value == "" {
		s.LoadScrapeSettings(&CloudflareScrapeSettings{APIURL: "", APIKey: ""})
		return nil
	}

	var scrapeSettings CloudflareScrapeSettings
	if err := json.Unmarshal([]byte(value), &scrapeSettings); err != nil {
		slog.Error("settings:LoadScrapeSettingsFromDB", "message", "failed to unmarshal scrape settings", "error", err)
		return fmt.Errorf("failed to unmarshal scrape settings")
	}

	s.LoadScrapeSettings(&scrapeSettings)
	return nil
}

func (s *SettingsManager) LoadChatDefaultPromptFromDB() error {
	value, err := s.dao.GetSettingValue(CHAT_DEFAULT_PROMPT_KEY)
	if err != nil {
		if err == sql.ErrNoRows {
			s.LoadChatDefaultPrompt(DEFAULT_CHAT_PROMPT)
			return nil
		}
		slog.Error("settings:LoadChatDefaultPromptFromDB", "message", "failed to get prompt value", "error", err)
		return fmt.Errorf("failed to get chat default prompt value")
	}

	s.LoadChatDefaultPrompt(extractChatDefaultPrompt(value))
	return nil
}

func (s *SettingsManager) LoadAgenticMaxTurnsFromDB() error {
	value, err := s.dao.GetSettingValue(DEFAULT_AGENCIC_MAX_TURNS_KEY)
	if err != nil {
		if err == sql.ErrNoRows {
			s.LoadAgenticMaxTurns(DEFAULT_AGENCIC_MAX_TURNS)
			return nil
		}
		slog.Error("settings:LoadAgenticMaxTurnsFromDB", "message", "failed to get max turns value", "error", err)
		return fmt.Errorf("failed to get agentic max turns value")
	}

	s.LoadAgenticMaxTurns(extractAgenticMaxTurns(value))
	return nil
}

func (cm *SettingsManager) GetProviderSetting(providerName string) (*proto.ProviderSettings, error) {
	key := "provider." + providerName
	val, err := cm.dao.GetSettingValue(key)
	if err != nil {
		slog.Error("settings:GetProviderSetting", "provider", providerName, "error", err)
		return nil, fmt.Errorf("failed to get provider setting: %w", err)
	}

	if val == "" {
		return nil, nil
	}

	// Parse the JSON
	var ps proto.ProviderSettings
	if err := json.Unmarshal([]byte(val), &ps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal provider settings: %w", err)
	}

	return &ps, nil
}

func extractChatDefaultPrompt(value string) string {
	if value == "" {
		return DEFAULT_CHAT_PROMPT
	}

	var settingsMap map[string]interface{}
	if err := json.Unmarshal([]byte(value), &settingsMap); err == nil {
		if prompt, ok := settingsMap["value"].(string); ok && prompt != "" {
			return prompt
		}
	}

	return value
}

func extractAgenticMaxTurns(value string) int {
	if value == "" {
		return DEFAULT_AGENCIC_MAX_TURNS
	}

	var settingsMap map[string]interface{}
	if err := json.Unmarshal([]byte(value), &settingsMap); err == nil {
		switch v := settingsMap["value"].(type) {
		case float64:
			if int(v) > 0 {
				return int(v)
			}
		case string:
			if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
				return parsed
			}
		}
	}

	if parsed, err := strconv.Atoi(value); err == nil && parsed > 0 {
		return parsed
	}

	return DEFAULT_AGENCIC_MAX_TURNS
}
