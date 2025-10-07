## Onboarding Flow – First Boot Setup (Desktop App)

This document describes the full implementation plan for the first-run onboarding experience. Users will be guided through configuring API access and embeddings provider before entering the app.

### Goals
- Show a multi-step onboarding wizard on first launch only
- Step 1: Configure OpenAI API Key or LiteLLM endpoint (with validation)
- Step 2: Configure Ollama URL for embeddings (with validation)
- Step 3: Finish screen and mark onboarding complete
- Persist settings via `SettingService` and only then disable onboarding

### Current State
Backend has `is_first_boot` stored in settings. Behavior today:

```50:57:backend/chatservice/service/settings_service.go
func (s *SettingService) FirstBootComplete() {
    slog.Info("settings_service:FirstBootComplete", "settingService", s)
    err := s.dao.SetSettingValue("is_first_boot", "1")
    if err != nil {
        slog.Error("settings_service:FirstBootComplete", "message", "failed to set is_first_boot setting", "error", err)
    }
}
```

```101:123:backend/chatservice/service/settings_service.go
func (s *SettingService) IsFirstBoot() (bool, error) {
    value, err := s.dao.GetSettingValue("is_first_boot")
    if err != nil {
        if err == sql.ErrNoRows { return true, nil }
        return false, fmt.Errorf("error getting is_first_boot setting: %w", err)
    }
    intValue, err := strconv.Atoi(value)
    if err != nil { return true, nil }
    return intValue == 0, nil
}
```

And in init:

```34:48:backend/chatservice/service/settings_service.go
func (s *SettingService) Init() {
    isFirstBoot, err := s.IsFirstBoot()
    if isFirstBoot {
        s.SetSetting(context.Background(), settings.DefaultSettings.ToProto())
    }
    s.FirstBootComplete()
}
```

Issue: calling `FirstBootComplete()` during init flips `is_first_boot` before the UI can run onboarding. We will change this to complete onboarding only after the user finishes the wizard.

---

## Backend Changes

### 1) Adjust initialization logic
- Remove the unconditional call to `FirstBootComplete()` from `SettingService.Init()`.
- Keep default settings seeding on true first boot.

### 2) Extend proto to expose onboarding controls
- Add two RPCs in `SettingService`:
  - `IsFirstBoot(IsFirstBootRequest) -> IsFirstBootResponse { bool is_first_boot }`
  - `CompleteOnboarding(CompleteOnboardingRequest) -> CompleteOnboardingResponse`

Proposed proto additions in `proto/chatservice.proto`:

```proto
service SettingService {
  rpc GetSetting(GetSettingRequest) returns (GetSettingResponse);
  rpc SetSetting(SetSettingRequest) returns (SetSettingResponse);
  // New
  rpc IsFirstBoot(IsFirstBootRequest) returns (IsFirstBootResponse);
  rpc CompleteOnboarding(CompleteOnboardingRequest) returns (CompleteOnboardingResponse);
}

message IsFirstBootRequest {}
message IsFirstBootResponse { bool is_first_boot = 1; }

message CompleteOnboardingRequest {}
message CompleteOnboardingResponse { string message = 1; }
```

Implementation notes:
- `IsFirstBoot` should just call `service.IsFirstBoot()` and return the boolean.
- `CompleteOnboarding` should call `service.FirstBootComplete()`.
- In `service.Init()`, remove `FirstBootComplete()` so first boot stays true until onboarding finishes.

### 3) Code generation
- From `backend/chatservice/` run:

```bash
go generate
```

This regenerates Go code at `backend/chatservice/proto/` and web client code at `frontend/proto/`.

### 4) Monolith wiring
`backend/mono/main.go` already registers `SettingService`. After adding RPCs, no extra wiring is required beyond codegen and implementing the API server methods.

---

## Frontend Changes

### 1) Routing and entry
- On app start, query `IsFirstBoot` via gRPC-web client (generated in `frontend/proto/`).
- If `is_first_boot == true`, route to `/onboarding` wizard; otherwise proceed to the main app.

### 2) UI – shadcn components
- Use shadcn/ui for a clean multi-step wizard:
  - Stepper or Tabs for steps
  - Inputs, Textarea, Select, Button, Alert, Form
- Create `OnboardingWizard` with three steps:
  1. OpenAI or LiteLLM
     - Radio or Segmented control: “OpenAI” vs “LiteLLM”
     - If OpenAI: input `OPENAI_API_KEY`, optional `OPENAI_API_URL`
     - If LiteLLM: input `OPENAI_API_URL` (LiteLLM proxy), optional API key if needed
     - Validate on Next (see validation rules below)
  2. Embeddings provider
     - Input `OLLAMA_URL` (e.g., http://localhost:11434)
     - Validate connectivity and presence of required embeddings model
  3. Finish
     - Summary of saved values
     - “Finish” button to persist and then call `CompleteOnboarding`

Suggested file structure:
- `frontend/src/routes/onboarding/index.tsx` – page wrapper
- `frontend/src/components/onboarding/Wizard.tsx`
- `frontend/src/components/onboarding/steps/StepApi.tsx`
- `frontend/src/components/onboarding/steps/StepEmbeddings.tsx`
- `frontend/src/components/onboarding/steps/StepFinish.tsx`
- `frontend/src/lib/validators/onboarding.ts`
 - `frontend/src/stores/onboardingStore.ts` (Nanostores)

### 3) State and persistence (Nanostores)
- Use Nanostores for wizard state in `frontend/src/stores/onboardingStore.ts`.
- Store shape (example):

```ts
type OnboardingState = {
  provider: 'openai' | 'litellm';
  OPENAI_API_KEY: string;
  OPENAI_API_URL: string;
  OLLAMA_URL: string;
  step: 0 | 1 | 2;
};
```

- Expose actions: `setProvider`, `setKey`, `setApiUrl`, `setOllamaUrl`, `next`, `back`, `reset`.
- Components subscribe via `useStore` and update via actions.
- On each step, validate inputs. Persist only on Finish:
  - Build `Settings` message: `{ OPENAI_API_KEY, OPENAI_API_URL, OLLAMA_URL }`
  - Call `SetSetting` via gRPC-web.
  - On success, call `CompleteOnboarding`.
  - Route to app root.

### 4) Validation rules and checks
- Step 1 (OpenAI/LiteLLM):
  - If OpenAI: require `OPENAI_API_KEY` non-empty.
  - If LiteLLM: require `OPENAI_API_URL` non-empty and HTTP(S).
  - Connectivity validation options:
    - Attempt a lightweight model list call to backend proxy if available (`ListModel`) after saving temp config locally, or add a temporary “validate” call in the backend to ping provider based on inputs (preferred client-side: defer saving until final step; use a dedicated validate endpoint passing candidate values).
    - Acceptable fallback: simple HEAD/GET to provided URL when LiteLLM.
- Step 2 (Ollama URL):
  - Require `OLLAMA_URL` non-empty and HTTP(S).
  - Validate by calling `${OLLAMA_URL}/api/tags` and expect 200 JSON.
- Surface errors inline with shadcn `Form` + `Alert`.

### 5) gRPC-web usage
- Use generated client under `frontend/proto/` to call:
  - `SettingService.IsFirstBoot({})`
  - `SettingService.SetSetting({ settings })`
  - `SettingService.CompleteOnboarding({})`

### 6) Access control
- `IsFirstBoot` and onboarding calls should be allowed without auth for desktop first-run. If auth is enforced, mark these methods as skipped in the auth interceptor/middleware similarly to existing health/static rules.

---

## Backend Implementation Steps
1. Update `proto/chatservice.proto` with new RPCs and messages.
2. Implement server handlers in `backend/chatservice/api/setting_service.go`:
   - `IsFirstBoot` -> call `service.IsFirstBoot()` and return response
   - `CompleteOnboarding` -> call `service.FirstBootComplete()`
3. Update `backend/chatservice/service/settings_service.go`:
   - Remove `s.FirstBootComplete()` from `Init()`
   - Keep default settings seeding on first boot
4. Run codegen: `go generate` from `backend/chatservice/`
5. Build and run monolith: from `backend/`:

```bash
go run ./mono
```

---

## Frontend Implementation Steps
1. Generate latest proto web stubs via backend `go generate` (outputs to `frontend/proto/`).
2. Add onboarding route and shadcn-based wizard components.
3. On app bootstrap, call `IsFirstBoot` and route accordingly.
4. Implement validation functions in `frontend/src/lib/validators/onboarding.ts`.
5. On Finish: call `SetSetting`, then `CompleteOnboarding`, then navigate to main.

---

## UX Copy (suggested)
- Step 1 title: “Connect your LLM provider”
  - Description: “Use an OpenAI key or a LiteLLM proxy URL.”
- Step 2 title: “Enable embeddings (Ollama)”
  - Description: “Provide your local Ollama URL to power retrieval and embeddings.”
- Step 3 title: “All set!”
  - Description: “Your configuration is saved. You can change it anytime in Settings.”

---

## Testing
- Fresh install
  - Ensure `is_first_boot` is absent or 0 -> onboarding shows
  - Complete wizard -> settings saved, `is_first_boot` set to 1, next launch skips onboarding
- Partial inputs
  - Invalid OpenAI key -> shows error, cannot proceed
  - Invalid LiteLLM URL -> shows error
  - Invalid Ollama URL -> shows error
- Regression
  - Existing users with `is_first_boot=1` never see onboarding

---

## Rollout Notes
- This change is backward compatible for existing users.
- Ensure auth skip-list includes new onboarding RPCs if auth is enabled at startup.
- Document environment defaults so validation can run offline.


