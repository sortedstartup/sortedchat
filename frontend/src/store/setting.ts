// store/setting.ts
import {
  Settings,
  SettingServiceClient,
  IsFirstBootRequest,
  TestConnectionRequest,
  TestConnectionResponse,
  ConnectionType,
  GetAllProviderSettingsRequest,
  GetAllProviderSettingsResponse,
  SetAllProviderSettingsRequest,
  SetProviderSettingRequest,
  SetProviderSettingResponse,
  ProviderSettings,
} from "../../proto/chatservice";
import { atom, onMount } from "nanostores";
import { createAuthenticatedClientOptions } from "../lib/auth";
import { getUIConfig } from "../lib/config";

let _settingClient: SettingServiceClient | undefined;

function getClient(): SettingServiceClient {
  if (!_settingClient) {
    const config = getUIConfig();
    if (!config) {
      throw new Error("UI config not loaded, cannot initialize chat client.");
    }
    _settingClient = new SettingServiceClient(
      config.API_URL,
      {},
      createAuthenticatedClientOptions(),
    );
  }
  return _settingClient;
}

export const $settings = atom<Settings>(new Settings({}));
export const $providerSettings = atom<Map<string, ProviderSettings>>(new Map());
export const $isLoadingProviderSettings = atom<boolean>(false);

// Onboarding state
export const $onboardingStep = atom<number>(0);
export const $onboardingData = atom<{
  OPENAI_API_KEY: string;
  CLAUDE_API_KEY: string;
  GEMINI_API_KEY: string;
  CLAUDE_API_URL: string;
  OPENAI_API_URL: string;
  GEMINI_API_URL: string;
  OLLAMA_URL: string;
}>({
  OPENAI_API_KEY: "",
  GEMINI_API_KEY: "",
  CLAUDE_API_KEY: "",
  CLAUDE_API_URL: "https://api.anthropic.com/v1/chat/completions",
  OPENAI_API_URL: "https://api.openai.com/v1/chat/completions",
  GEMINI_API_URL: "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
  OLLAMA_URL: "http://localhost:8081/v1/embeddings",
});

// Onboarding actions
export const onboardingActions = {
  setOpenaiApiKey: (key: string) => {
    const data = $onboardingData.get();
    $onboardingData.set({ ...data, OPENAI_API_KEY: key });
  },

  setGeminiApiKey: (key: string) => {
    const data = $onboardingData.get();
    $onboardingData.set({ ...data, GEMINI_API_KEY: key });
  },

  setClaudeApiKey: (key: string) => {
    const data = $onboardingData.get();
    $onboardingData.set({ ...data, CLAUDE_API_KEY: key });
  },

  setClaudeApiUrl: (url: string) => {
    const data = $onboardingData.get();
    $onboardingData.set({ ...data, CLAUDE_API_URL: url });
  },

  setOpenaiApiUrl: (url: string) => {
    const data = $onboardingData.get();
    $onboardingData.set({ ...data, OPENAI_API_URL: url });
  },

  setGeminiApiUrl: (url: string) => {
    const data = $onboardingData.get();
    $onboardingData.set({ ...data, GEMINI_API_URL: url });
  },

  setOllamaUrl: (url: string) => {
    const data = $onboardingData.get();
    $onboardingData.set({ ...data, OLLAMA_URL: url });
  },

  nextStep: () => {
    const current = $onboardingStep.get();
    if (current < 2) {
      $onboardingStep.set(current + 1);
    }
  },

  prevStep: () => {
    const current = $onboardingStep.get();
    if (current > 0) {
      $onboardingStep.set(current - 1);
    }
  },

  testConnection: async (
    url: string,
    type: ConnectionType,
  ): Promise<TestConnectionResponse> => {
    try {
      const req = new TestConnectionRequest({ url, connection_type: type });
      const res = await getClient().TestConnection(req, {});
      return res;
    } catch (error) {
      console.error("Failed to test connection:", error);
      throw new Error("Failed to test connection");
    }
  },

  completeOnboarding: async (): Promise<void> => {
    try {
      const data = $onboardingData.get();

      const settings = new Map<string, ProviderSettings>();
      settings.set('openai', new ProviderSettings({
        api_key: data.OPENAI_API_KEY,
        api_url: data.OPENAI_API_URL,
        is_enabled: true
      }));
      settings.set('claude', new ProviderSettings({
        api_key: data.CLAUDE_API_KEY,
        api_url: data.CLAUDE_API_URL,
        is_enabled: true
      }));
      settings.set('gemini', new ProviderSettings({
        api_key: data.GEMINI_API_KEY,
        api_url: data.GEMINI_API_URL,
        is_enabled: true
      }));

      await getClient().SetAllProviderSettings(new SetAllProviderSettingsRequest({ settings }), {});

      window.location.replace("/");
    } catch (error) {
      console.error("Failed to complete onboarding:", error);
      throw new Error("Failed to complete onboarding");
    }
  },
};
export const GetIsFirstBootStatus = async (): Promise<boolean> => {
  try {
    const req = new IsFirstBootRequest({});
    const res = await getClient().IsFirstBoot(req, {});
    return res.is_first_boot;
  } catch (error) {
    console.error("Failed to check if first boot:", error);
    throw error;
  }
};

export const GetAllProviderSettings = async () => {
  $isLoadingProviderSettings.set(true);
  try {
    const req = new GetAllProviderSettingsRequest({});
    const res: GetAllProviderSettingsResponse = await getClient().GetAllProviderSettings(req, {});

    $providerSettings.set(res.settings as Map<string, ProviderSettings>);
    $isLoadingProviderSettings.set(false);
  } catch (error) {
    console.error("Failed to fetch provider settings:", error);
    $isLoadingProviderSettings.set(false);
  }
};

export const SetProviderSetting = async (
  providerName: string,
  settings: ProviderSettings
): Promise<string> => {
  try {
    const req = new SetProviderSettingRequest({
      name: providerName,
      settings: settings,
    });
    const res: SetProviderSettingResponse = await getClient().SetProviderSetting(req, {});

    // Refresh all provider settings after successful save
    await GetAllProviderSettings();

    return res.message ?? "Provider settings saved successfully";
  } catch (error) {
    console.error("Failed to save provider settings:", error);
    throw new Error("Failed to save provider settings");
  }
};

onMount($providerSettings, () => {
  GetAllProviderSettings();
  return () => { };
});
