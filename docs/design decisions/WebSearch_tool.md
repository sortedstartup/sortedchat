# WebSearch Tool

## Overview

We added agentic chat support inside the normal chat flow to enable web search when the selected model supports tool calling.

The flow works as follows:

- In `service.go`, the `Chat` function checks the selected model's capabilities.
- For now, model capabilities are seeded in `shared_models_metadata`.
- If the current message does not contain images, model supports tool calling and websearch api key is present, we route the request through the agentic chat flow.
- We do not currently support images in `sortedagents`, so image-based chat requests continue to use the normal LLM path.

## Chat Agent Behavior

The chat agent is responsible for deciding whether web search is needed for a given user request.

- If web search is needed, the agent calls the web search tool.
- The tool uses the Brave Search API.
- If web search is not needed, the agent responds directly without calling the tool.
- If the tool call or agentic flow fails, the request falls back to the normal LLM chat path.
- The maximum number of turns for the agent loop is `4`.

## Settings

We store Brave Search configuration in the `settings` table.

- Setting name: `tool.websearch.brave`
- Setting value:

```json
{
  "apiUrl": "https://api.search.brave.com/res/v1/web/search",
  "apiKey": "...apikey..."
}
```

The API key and API URL are provided by the user through the settings page.

We also store the default chat agent prompt in the `settings` table.

- Setting name: `chat.default_prompt`
- Setting value: plain text prompt

## Current Limitations

- `sortedagents` does not yet support image input.
- In normal `chat_messages`, we do not currently persist web search tool-call history or tool-call failure details.
- Brave Search API usage cost is not yet tracked.

## Future Improvements

- Add image support in `sortedagents`.
- Persist tool-call activity and failures for normal chat flows.
- Add Brave Search cost tracking and reporting.
- We could add RAG as a Tool
- we can entirely remove, direct llm call, if the model does not support function/Tool calling, simply start teh agent with no tools
- add scraping tool to get better result