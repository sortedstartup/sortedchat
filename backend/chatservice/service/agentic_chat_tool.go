package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

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
		apiURL:     defaultBraveSearchAPIURL,
	}
}

func NewBraveSearchToolWithConfig(apiURL, apiKey string) *BraveSearchTool {
	if strings.TrimSpace(apiURL) == "" {
		apiURL = defaultBraveSearchAPIURL
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

// BrowserScrapeTool scrapes web pages using Cloudflare Browser Rendering API
type BrowserScrapeTool struct {
	httpClient *http.Client
	apiURL     string
	apiKey     string
}

func NewBrowserScrapeToolWithConfig(apiURL, apiKey string) *BrowserScrapeTool {
	return &BrowserScrapeTool{
		httpClient: http.DefaultClient,
		apiURL:     strings.TrimSpace(apiURL),
		apiKey:     strings.TrimSpace(apiKey),
	}
}

func (t *BrowserScrapeTool) Name() string {
	return "browser_scrape"
}

func (t *BrowserScrapeTool) Description() string {
	return "Scrape a web page and convert it to markdown using Cloudflare Browser Rendering API"
}

func (t *BrowserScrapeTool) Parameters() *sortedagents.JSONSchema {
	return &sortedagents.JSONSchema{
		Type: "object",
		Properties: map[string]sortedagents.JSONSchema{
			"url": {
				Type:        "string",
				Description: "The URL to scrape",
			},
		},
		Required: []string{"url"},
	}
}

type BrowserScrapeResult struct {
	Success bool   `json:"success"`
	Result  string `json:"result"`
}

const scrapeUsageSecondsResultKey = "_scrape_api_usage_seconds"

func (t *BrowserScrapeTool) Execute(ctx context.Context, args map[string]any) (any, error) {
	slog.Info("executing browser scrape tool")

	urlStr, ok := args["url"].(string)
	if !ok || urlStr == "" {
		slog.Error("BrowserScrapeTool:Execute", "message", "url parameter is required")
		return nil, fmt.Errorf("url parameter is required")
	}

	if t.apiURL == "" || t.apiKey == "" {
		slog.Error("BrowserScrapeTool:Execute", "message", "cloudflare scrape api url and api key must be configured")
		return nil, fmt.Errorf("cloudflare scrape api url and api key must be configured")
	}

	// Create JSON body
	requestBody := map[string]string{"url": urlStr}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		slog.Error("BrowserScrapeTool:Execute", "message", "failed to marshal request body", "error", err)
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		slog.Error("BrowserScrapeTool:Execute", "message", "failed to create request", "error", err)
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.httpClient.Do(req)
	if err != nil {
		slog.Error("BrowserScrapeTool:Execute", "message", "scrape request failed", "error", err)
		return nil, fmt.Errorf("scrape request failed: %v", err)
	}
	defer resp.Body.Close()

	// Handle rate limiting (429)
	if resp.StatusCode == 429 {
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			slog.Info("Rate limited. Waiting before retry", "retry_after_seconds", retryAfter)
			// Parse retry-after as seconds
			var waitSeconds int
			fmt.Sscanf(retryAfter, "%d", &waitSeconds)
			if waitSeconds > 0 && waitSeconds <= 60 { // Max 60 seconds wait
				time.Sleep(time.Duration(waitSeconds) * time.Second)

				// Retry the request
				retryReq, err := http.NewRequestWithContext(ctx, "POST", t.apiURL, bytes.NewBuffer(jsonBody))
				if err != nil {
					return nil, fmt.Errorf("failed to create retry request: %v", err)
				}
				retryReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", t.apiKey))
				retryReq.Header.Set("Content-Type", "application/json")

				resp, err = t.httpClient.Do(retryReq)
				if err != nil {
					return nil, fmt.Errorf("retry request failed: %v", err)
				}
				defer resp.Body.Close()
			}
		}
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("BrowserScrapeTool:Execute", "message", "failed to read response body", "error", err)
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	scrapeUsageSeconds, err := parseScrapeUsageSeconds(resp.Header.Get("X-Browser-Ms-Used"))
	if err != nil {
		slog.Warn("BrowserScrapeTool:Execute", "message", "failed to parse scrape usage header", "error", err, "header", resp.Header.Get("X-Browser-Ms-Used"))
		scrapeUsageSeconds = 0
	}

	slog.Debug("Cloudflare API response", "status", resp.StatusCode)

	// Parse the response
	var result BrowserScrapeResult
	if err := json.Unmarshal(body, &result); err != nil {
		slog.Error("BrowserScrapeTool:Execute", "message", "failed to parse response", "error", err, "response_body", string(body))
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	slog.Info("Parsed result", "success", result.Success, "result_length", len(result.Result))

	return map[string]interface{}{
		"url":                       urlStr,
		"markdown":                  result.Result,
		scrapeUsageSecondsResultKey: scrapeUsageSeconds,
	}, nil
}

func parseScrapeUsageSeconds(browserMsUsed string) (float64, error) {
	browserMsUsed = strings.TrimSpace(browserMsUsed)
	if browserMsUsed == "" {
		return 0, nil
	}

	milliseconds, err := strconv.ParseFloat(browserMsUsed, 64)
	if err != nil {
		return 0, fmt.Errorf("parse browser ms used: %w", err)
	}
	if milliseconds < 0 {
		return 0, fmt.Errorf("browser ms used must be non-negative")
	}

	return milliseconds / 1000.0, nil
}
