package settings

import "sortedstartup/chatservice/proto"

/*
- TODO: To think : settings have to be app level and then broken down to the service level
*/
type Settings struct {
	OpenAIAPIKey string `koanf:"openai_api_key" json:"openai_api_key"`
	OpenAIAPIURL string `koanf:"openai_api_url" json:"openai_api_url"`
	OllamaURL    string `koanf:"ollama_url" json:"ollama_url"`
}

var DefaultSettings = &Settings{
	OpenAIAPIKey: "",
	OpenAIAPIURL: "https://api.openai.com/v1/chat/completions",
	OllamaURL:    "",
}

func (s *Settings) ToProto() *proto.Settings {
	return &proto.Settings{
		OPENAI_API_KEY: s.OpenAIAPIKey,
		OPENAI_API_URL: s.OpenAIAPIURL,
		OLLAMA_URL:     s.OllamaURL,
	}
}

func FromProto(protoSettings *proto.Settings) *Settings {
	return &Settings{
		OpenAIAPIKey: protoSettings.OPENAI_API_KEY,
		OpenAIAPIURL: protoSettings.OPENAI_API_URL,
		OllamaURL:    protoSettings.OLLAMA_URL,
	}
}
