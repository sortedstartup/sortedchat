package llm

type CustomChatRequest struct {
	ModelName     string         `json:"model_name"`
	Messages      []Message      `json:"messages"`
	Stream        bool           `json:"stream"`
	StreamOptions *StreamOptions `json:"stream_options,omitempty"`
}

type StreamOptions struct {
	IncludeUsage bool `json:"include_usage"`
}

type Message struct {
	Role    string      `json:"role"`
	Content interface{} `json:"content"` // string or []ContentPart
}

type ContentPart struct {
	Type     string    `json:"type"`                // "text" or "image_url"
	Text     string    `json:"text,omitempty"`      // Populated if Type is "text"
	ImageURL *ImageURL `json:"image_url,omitempty"` // Populated if Type is "image_url"
}

type ImageURL struct {
	URL string `json:"url"`
}
