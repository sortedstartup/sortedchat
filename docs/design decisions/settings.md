# Settings

How we handle various kinds of settings in our app.

## User Settings
Right now all settings are bundled up together in a settings table
## System Settings
Right now all settings are bundled up together in a settings table
## Tenant Settings
Right now all settings are bundled up together in a settings table

## High level Idea
- Each "Setting" will be a nested JSON value (at proto API level it is a a 'Struct').
- Each setting has a name

examples :
- "provider.settings" (<- name) : (value ->)  { "openai":{}, "google":{}   }

API for getting and setting these settings - 
GetSetting(name) proto.Struct (https://github.com/protocolbuffers/protobuf/blob/main/src/google/protobuf/struct.proto)
SetSetting(name string, value proto.Struct)

### Special case: provider settings
For a special case of provider settings we have special helpers functions to make schema management easy
These api functions will be used in the backend and the frontend.

```proto
// will be used when chatting with specific providers
// The name format examples -
//   - 'provider.openai', 'provider.google', 'provider.openrouter'
GetProviderSetting(name string) (ProviderSettings)
SetProviderSetting(name string, ProviderSettings)

// used by ui to make it easy to read and write settings in one API call
GetAllProviderSettings(GetAllProviderSettingsRequest) (GetAllProviderSettingsResponse)
SetAllProviderSettings(SetAllProviderSettingsRequest) (SetAllProvidersSettingsResponse)
```

```proto
message ProviderSettings {
    string api_url = 1;
   string api_key = 2;
   bool is_enabled = 3;
}

message GetAllProviderSettingsResponse {
    AllProviderSettings settings = 1;
}

message GetAllProviderSettingsRequest {
}

message SetAllProviderSettings {
    map<string, ProviderSettings> settings = 1;
}

message SetAllProviderSettings {
    map<string, ProviderSettings> settings = 1;
}
```

## Settings Table
In database there is a settings table
|name|string(json)|

|provider.openai|{/*json of the type ProviderSettings*/}|
|provider.google|{}|
|provider.openrouter|{}|
