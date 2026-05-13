# Add Provider

## Overview 
Currently, we have hardcoded the provider and models in code. We need to add a way to add providers and models anytime user wants to add a new provider or model.

## Condition
Provider and model, user is trying to add should have openai chat completions compatible api endpoint.

## User Experience 
- User has to go to Models section in sidebar
- there above list of provider in left panel, should have a button of "Add Provider", basically put an add icon.
- when clicked on add icon, it should open a modal to add provider details(provider name,api_url, api_key, is_enabled(true/false, default true)).
- in provider's list, they should be able to see provider and on click of that right panel should show list of models(currently 0 models for new provider) for that provider.
- then there should be a button to add model for that provider, which should open a modal to add model details(model id(mandatory), input_token_cost(optional), output_token_cost(optional), cached_token_cost(optional),is_embedding_model(true/false, default false))


## Technical Implementation
### Backend changes

#### Request Templates
- When calling an LLM provider, the system uses a request template from `backend/chatservice/llm/templates/`. 
- If a provider-specific template is not found, it defaults to the `openai` template.

#### Adding Provider
There is already a  rpc's for setting a provider, we can use that to set provider.
```proto
    rpc SetProviderSetting(SetProviderSettingRequest) returns (SetProviderSettingResponse);
    message SetProviderSettingRequest {
        string name = 1;
        ProviderSettings settings = 2;
    }

    message SetProviderSettingResponse {
        string message = 1;
    }
```
- We just have to implement the UI for adding provider.

#### Adding Model
We added a new rpc to add a model for a provider.
```proto
    rpc AddModel(AddModelRequest) returns (AddModelResponse);
    message AddModelRequest {
        string provider_name = 1;
        string model_id = 2;
        string model_name = 3;
        string url = 4;
        float input_token_cost = 5;
        float output_token_cost = 6;
        float cached_token_cost = 7;
        bool is_embedding_model = 8;
    }
    message AddModelResponse {
        string message = 1;
    }
```

#### Database Queries
We have a table shared_models_metadata, there we have to add model details.
```sql
-- Add model for sqlite
INSERT INTO shared_models_metadata (provide, id, name, url, input_token_cost, output_token_cost, cached_token_cost, is_embedding_model)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);
```
```sql
-- Add model for postgres
INSERT INTO shared_models_metadata (provide, id, name, url, input_token_cost, output_token_cost, cached_token_cost, is_embedding_model)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);
```

### Frontend changes
- use store/setting.ts to get and set provider details.
    - in provider's modal, ask for model name,url, api key, is_enabled(true/false, default true)
- add a new function for integrating, AddModel rpc in store/chat.ts
    - in model's modal, ask for model id(mandatory), input_token_cost(optional), output_token_cost(optional), cached_token_cost(optional),is_embedding_model(true/false, default false)

### Frontend implementation decisions

#### API usage
- Provider APIs are integrated in `setting.ts`.
- `GetAllProviderSettings()` loads all saved provider configs into the frontend store.
- `SetProviderSetting(providerName, settings)` is used by `add-provider-dialog.tsx` and `provider-view.tsx` to create or update provider settings.
- Model API is integrated in `chat.ts` as `addModel(req)`.
- Before calling `AddModel`, the store reads the selected provider's `api_url` from provider settings and sets `req.url`.

#### Nanostore usage
- `$providerSettings` in `setting.ts` is the source of truth for provider config on the frontend.
- After `SetProviderSetting`, the store refreshes `$providerSettings` by calling `GetAllProviderSettings()` again.
- After `addModel(...)`, the store refreshes model data using `fetchAvailableModels()` and `ListLLMModels()` so UI updates from store state instead of manual component mutation.

#### Adding default provider
- Default providers like `openai`, `claude`, `gemini`, `fireworks`, and `openrouter` are not seeded in DB from backend.
- On onboarding, frontend prepares provider settings for these providers and sends them using `SetAllProviderSettings(...)`. 
    - Don't know why we have done it this way
- Backend saves them in DB as provider settings entries like `provider.openai`, `provider.claude`, `provider.fireworks`, etc.
