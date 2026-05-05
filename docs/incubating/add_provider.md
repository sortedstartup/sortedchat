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
There are already two rpc's for provider getting and setting, we can use them to get and set provider.
```proto
    rpc GetProviderSetting(GetProviderSettingRequest) returns (GetProviderSettingResponse);
    rpc SetProviderSetting(SetProviderSettingRequest) returns (SetProviderSettingResponse);
    message GetProviderSettingRequest {
        string name = 1;
    }

    message GetProviderSettingResponse {
        ProviderSettings settings = 1;
    }

    message SetProviderSettingRequest {
        string name = 1;
        ProviderSettings settings = 2;
    }

    message SetProviderSettingResponse {
        string message = 1;
    }
```
- So, No changes required for adding provider in backend
- We just have to implement the UI for adding provider.

#### Adding Model
We have to add a new rpc to add a model for a provider.
```proto
    rpc AddModel(AddModelRequest) returns (AddModelResponse);
    message AddModelRequest {
        string provider_name = 1;
        string model_id = 2;
        string model_name = 3;
        string url = 4;
        int32 input_token_cost = 5;
        int32 output_token_cost = 6;
        int32 cached_token_cost = 7;
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
INSERT INTO shared_models_metadata (provider_name, model_id, model_name, url, input_token_cost, output_token_cost, cached_token_cost, is_embedding_model)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);
```
```sql
-- Add model for postgres
INSERT INTO shared_models_metadata (provider_name, model_id, model_name, url, input_token_cost, output_token_cost, cached_token_cost, is_embedding_model)
VALUES (?, ?, ?, ?, ?, ?, ?, ?);
```

### Frontend changes
- use store/setting.ts to get and set provider details.
    - in provider's modal, ask for model name,url, api key, is_enabled(true/false, default true)
- add a new function for integrating, AddModel rpc in store/chat.ts
    - in model's modal, ask for model id(mandatory), input_token_cost(optional), output_token_cost(optional), cached_token_cost(optional),is_embedding_model(true/false, default false)
- look for modal component in components folder and try to reuse the same for adding provider and model.
- keep code changes minimal
- all business logic should be stores only and stores should be used by components to get and set data.