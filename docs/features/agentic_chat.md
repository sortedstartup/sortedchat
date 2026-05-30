# WebSearch Tool

## Overview

We added agentic chat support inside the normal chat flow to enable web search when the selected model supports tool calling.

The flow works as follows:

- In `service.go`, the `Chat` function checks the selected model's capabilities.
- For now, model capabilities are seeded in `shared_models_metadata`.
- If the current message supports tool calling, websearch api key is present, and provider settings are set, we route the request through the agentic chat flow.

## Chat Agent Behavior

The chat agent is responsible for deciding whether web search is needed for a given user request.

- If web search is needed, the agent calls the web search tool.
- The tool uses the Brave Search API.
- If web search is not needed, the agent responds directly without calling the tool.
- If the tool call or agentic flow fails, the request falls back to the normal LLM chat path.
- The default maximum number of turns for the agent loop is `4`.
- If scrape tool setting are not set, even then agentic chat work but with scrape tool, if scrape tool settings are set, then only we pass the scrape tool to agent


## Settings

We store Brave Search configuration and the default chat agent prompt in the `settings` table, but normal reads during chat do not hit the DB every time anymore.

- Setting name: `tool.websearch.brave`
- Setting value:

```json
{
  "apiUrl": "https://api.search.brave.com/res/v1/web/search",
  "apiKey": "...apikey..."
}
```

- Setting name: `chat.default_system_prompt`
- Setting value:

```json
{
  "value": "You are SortedChat's default assistant..."
}
```

- Setting name: `tool.scrape.cloudlfare`
- Setting value:

```json
{
  "apiUrl:"",
  "apiKey":""
}
```

### Runtime Behavior

- DB is still the source of truth for both settings.
- `SettingsManager` keeps an in-memory copy of:
  - `tool.websearch.brave`
  - `chat.default_prompt`
- Chat reads these values from `SettingsManager`, not directly from `settingsDAO`, during request execution.

### Fake Message Bus Pattern

- When a setting is updated, `SettingService` writes to DB first.
- After the write, it publishes `settings.changed` on the in-memory queue.
- `SettingsManager` subscribes to that event and reloads the cached settings from DB.


## Sortedagents Image Support
- In `sortedagents.Message`, `Content` was changed from `string` to `MessageContent`.
- This matches the LLM API contract, where message content can be either a plain text string or an array of structured content parts.
- Example :
```
type ContentPart struct {
	Type     string    `json:"type"`                // "text" or "image_url"
	Text     string    `json:"text,omitempty"`      // Populated if Type is "text"
	ImageURL *ImageURL `json:"image_url,omitempty"` // Populated if Type is "image_url"
}

type ImageURL struct {
	URL string `json:"url"`
}
```
- `MessageContent` is kept as an interface so both supported representations can be handled through the same field.
- The interface is implemented by `TextContent` and `ContentParts`.
- If a message contains only text, we send `TextContent`. If it includes an image, we send `ContentParts`.


## Added Sources List per message
- Added new column in `chat_messages` table,  `metadata TEXT`
- This column stores assistant-message level metadata as JSON.
- Current stored shape:

```json
{
  "websearches": [
    { "query": "latest openai api pricing" }
  ],
  "sources": [
    { "url": "https://platform.openai.com/docs/pricing" },
    { "url": "https://openai.com/api/" }
  ]
}
```

### How metadata is collected

- In `agentic_chat.go`, during one agent run we keep two in-memory lists:
  - `webSearchQueries []string`
  - `sourceURLs []string`
- These lists exist only for the lifetime of the current assistant response generation.

### Collection points

- On `ToolCallStartEvent`:
  - if tool name is `web_search`, we read `args["query"]` and append it to `webSearchQueries`
  - if tool name is `browser_scrape`, we read `args["url"]` and append it to `sourceURLs`

This means:
- `websearches` stores web search queries
- `sources` stores only URLs that were actually scraped through `browser_scrape`

### How metadata is converted and saved

- `NewChatMessageMetadata(webSearchQueries, sourceURLs)` builds a `ChatMessageMetadata`.
- `ChatMessageMetadata.ToJSON()` serializes it for DB storage in `chat_messages.metadata`.
- `ChatMessageMetadata.ToProto()` converts it into `AssistantMessageMetadata` for gRPC responses.


### How metadata is returned

- DAO returns raw `metadata` as string from `chat_messages.metadata`.
- In `service.go`, before sending response/history:
  - if metadata string is non-empty, we `json.Unmarshal` it into `ChatMessageMetadata`
  - then call `ToProto()`
- This is used in:
  - `ResponseSummary.metadata` for the current streamed assistant message
  - `ChatMessage.metadata` in `GetHistory`
## Current Limitations

- In normal `chat_messages`, we do not currently persist web search tool-call history or tool-call failure details.
- Brave Search API usage cost is not yet tracked.

## Future Improvements

- Persist tool-call activity and failures for normal chat flows.
- Add Brave Search cost tracking and reporting.
- We could add RAG as a Tool
- we can entirely remove, direct llm call, if the model does not support function/Tool calling, simply start teh agent with no tools
- add scraping tool to get better result
