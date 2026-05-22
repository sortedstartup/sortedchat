# WebSearch Tool 
- Added agentic chat with normal chat 
- first we check the models capabilities
	- for now we have seeded capabilites in `shared_models_metadata`
- in service.go/Chat function - be check the capabilities of the model
	- if message does not have image and model support function/tool calling
	- we dont have image support in sortedagents
	- we run agent which have websearch tool 

## Chat Agent
- This agent first determines whether to use websearch or not based on user input
- if yes, then it call websearch tool using brave search api key
- if not then directly answer it
- if tool call fails, we route back to normal llm chat
- Max turns for this agent is 4 

## Settings
- In settings table, we added a new row with name "websearch"
- setting value, {"brave_search_api_key":"...apikey..."}
- we take this key from user from settings page


## What need to be added 
- Added image support in sortedagents
- Right now, we don't have save websearch tool call or tool call failure in normal chat_messages
