package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	settings "sortedstartup/chatservice/settings"
	"sortedstartup/chatservice/sortedagents"
)

type BraveSearchTool struct {
	httpClient *http.Client
	apiURL     string
	apiKey     string
}

type braveSearchArgs struct {
	Query string `json:"query" description:"The web search query to execute"`
}

type braveSearchResponse struct {
	Web struct {
		Results []struct {
			Title       string `json:"title"`
			URL         string `json:"url"`
			Description string `json:"description"`
		} `json:"results"`
	} `json:"web"`
}

func NewBraveSearchTool() *BraveSearchTool {
	return &BraveSearchTool{
		httpClient: http.DefaultClient,
		apiURL:     settings.DEFAULT_BRAVE_SEARCH_API_URL,
	}
}

func NewBraveSearchToolWithConfig(apiURL, apiKey string) *BraveSearchTool {
	if strings.TrimSpace(apiURL) == "" {
		apiURL = settings.DEFAULT_BRAVE_SEARCH_API_URL
	}
	return &BraveSearchTool{
		httpClient: http.DefaultClient,
		apiURL:     apiURL,
		apiKey:     apiKey,
	}
}

func (t *BraveSearchTool) Name() string {
	return "web_search"
}

func (t *BraveSearchTool) Description() string {
	return "Search the web for fresh information and return a concise list of relevant results with titles, URLs, and snippets."
}

func (t *BraveSearchTool) Parameters() *sortedagents.JSONSchema {
	schema, _ := sortedagents.GenerateSchema[braveSearchArgs]()
	return schema
}

func (t *BraveSearchTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	slog.Info("BraveSearchTool:Execute", "message", "executing brave search tool", "args", args)
	if strings.TrimSpace(t.apiKey) == "" {
		slog.Error("BraveSearchTool:Execute", "message", "brave search api key is not configured")
		return nil, fmt.Errorf("brave search api key is not configured")
	}

	query, _ := args["query"].(string)
	query = strings.TrimSpace(query)
	if query == "" {
		slog.Error("BraveSearchTool:Execute", "message", "query parameter is required")
		return nil, fmt.Errorf("query parameter is required")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.apiURL+"?q="+url.QueryEscape(query), nil)
	if err != nil {
		slog.Error("BraveSearchTool:Execute", "message", "failed to create brave search request", "error", err)
		return nil, fmt.Errorf("failed to create brave search request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", t.apiKey)

	resp, err := t.httpClient.Do(req)
	if err != nil {
		slog.Error("BraveSearchTool:Execute", "message", "brave search request failed", "error", err)
		return nil, fmt.Errorf("brave search request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("BraveSearchTool:Execute", "message", "failed to read brave search response", "error", err)
		return nil, fmt.Errorf("failed to read brave search response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		slog.Error("BraveSearchTool:Execute", "message", "brave search failed", "statusCode", resp.StatusCode, "body", strings.TrimSpace(string(body)))
		return nil, fmt.Errorf("brave search failed with status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var payload braveSearchResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		slog.Error("BraveSearchTool:Execute", "message", "failed to parse brave search response", "error", err)
		return nil, fmt.Errorf("failed to parse brave search response: %w", err)
	}

	limit := len(payload.Web.Results)
	if limit > 5 {
		limit = 5
	}
	results := make([]map[string]string, 0, limit)
	for _, item := range payload.Web.Results {
		if len(results) == 5 {
			break
		}
		results = append(results, map[string]string{
			"title":   item.Title,
			"url":     item.URL,
			"snippet": item.Description,
		})
	}

	slog.Info("BraveSearchTool:Execute", "message", "brave search executed successfully", "query", query, "resultsCount", len(results))

	return map[string]any{
		"query":   query,
		"results": results,
	}, nil
}
