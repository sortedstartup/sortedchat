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
  GetSettingRequest,
  SetSettingRequest,
} from "../../proto/chatservice";
import { Struct, Value } from "../../proto/google/protobuf/struct";
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
export const $webSearchKey = atom<string>("");
export const $webSearchApiUrl = atom<string>("https://api.search.brave.com/res/v1/web/search");
export const $webSearchCost = atom<string>("0.005");
export const $isLoadingWebSearch = atom<boolean>(false);
export const $cloudflareScrapeApiUrl = atom<string>("");
export const $cloudflareScrapeApiKey = atom<string>("");
export const $isLoadingCloudflareScrape = atom<boolean>(false);
export const $agenticSystemPrompt = atom<string>("");
export const $agenticMaxTurns = atom<string>("4");
export const $isLoadingAgenticSettings = atom<boolean>(false);

// Onboarding state
export const $onboardingStep = atom<number>(0);
export const $onboardingData = atom<{
  OPENAI_API_KEY: string;
  CLAUDE_API_KEY: string;
  GEMINI_API_KEY: string;
  FIREWORKS_API_KEY: string;
  CLAUDE_API_URL: string;
  OPENAI_API_URL: string;
  GEMINI_API_URL: string;
  FIREWORKS_API_URL: string;
  OLLAMA_URL: string;
}>({
  OPENAI_API_KEY: "",
  GEMINI_API_KEY: "",
  CLAUDE_API_KEY: "",
  FIREWORKS_API_KEY: "",
  CLAUDE_API_URL: "https://api.anthropic.com/v1/chat/completions",
  OPENAI_API_URL: "https://api.openai.com/v1/chat/completions",
  GEMINI_API_URL: "https://generativelanguage.googleapis.com/v1beta/openai/chat/completions",
  FIREWORKS_API_URL: "https://api.fireworks.ai/inference/v1/chat/completions",
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

  setFireworksApiKey: (key: string) => {
    const data = $onboardingData.get();
    $onboardingData.set({ ...data, FIREWORKS_API_KEY: key });
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

  setFireworksApiUrl: (url: string) => {
    const data = $onboardingData.get();
    $onboardingData.set({ ...data, FIREWORKS_API_URL: url });
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
      settings.set('fireworks', new ProviderSettings({
        api_key: data.FIREWORKS_API_KEY,
        api_url: data.FIREWORKS_API_URL,
        is_enabled: true
      }));

      await getClient().SetAllProviderSettings(new SetAllProviderSettingsRequest({ settings }), {});

      // Delay navigation to allow settings to propagate
      setTimeout(() => {
        window.location.replace("/?welcome=true");
      }, 500);
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

    const settings = res.settings;
    if (settings instanceof Map) {
      $providerSettings.set(settings);
    } else {
      $providerSettings.set(new Map(Object.entries(settings || {})));
    }
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

function objectToStruct(data: Record<string, string>): Struct {
  const fields = new Map<string, Value>();
  for (const [key, value] of Object.entries(data)) {
    fields.set(key, new Value({ string_value: value }));
  }
  return new Struct({ fields });
}

function structToStringRecord(settings?: Struct): Record<string, string> {
  const result: Record<string, string> = {};
  if (!settings) {
    return result;
  }

  for (const [key, value] of settings.fields.entries()) {
    if (value.kind === "string_value") {
      result[key] = value.string_value;
    }
  }

  return result;
}

export const GetSetting = async (name: string): Promise<Record<string, string>> => {
  try {
    const req = new GetSettingRequest({ name });
    const res = await getClient().GetSetting(req, {});
    return structToStringRecord(res.settings);
  } catch (error) {
    console.error("Failed to fetch setting:", error);
    throw new Error("Failed to fetch setting");
  }
};

export const SetSetting = async (name: string, settings: Record<string, string>): Promise<string> => {
  try {
    const req = new SetSettingRequest({
      name,
      settings: objectToStruct(settings),
    });
    const res = await getClient().SetSetting(req, {});
    return res.message ?? "Settings saved successfully";
  } catch (error) {
    console.error("Failed to save setting:", error);
    throw new Error("Failed to save setting");
  }
};

export const GetWebSearchSetting = async (): Promise<void> => {
  $isLoadingWebSearch.set(true);
  try {
    const settings = await GetSetting("tool.websearch.brave");
    $webSearchApiUrl.set(
      settings.apiUrl ?? "https://api.search.brave.com/res/v1/web/search",
    );
    $webSearchCost.set(settings.cost ?? "0.005");
    $webSearchKey.set(settings.apiKey ?? "");
  } catch (error) {
    console.error("Failed to fetch websearch setting:", error);
    throw error;
  } finally {
    $isLoadingWebSearch.set(false);
  }
};

export const GetCloudflareScrapeSetting = async (): Promise<void> => {
  $isLoadingCloudflareScrape.set(true);
  try {
    const settings = await GetSetting("tool.scrape.cloudflare");
    $cloudflareScrapeApiUrl.set(settings.apiUrl ?? "");
    $cloudflareScrapeApiKey.set(settings.apiKey ?? "");
  } catch (error) {
    console.error("Failed to fetch cloudflare scrape setting:", error);
    throw error;
  } finally {
    $isLoadingCloudflareScrape.set(false);
  }
};

export const GetAgenticSettings = async (): Promise<void> => {
  $isLoadingAgenticSettings.set(true);
  try {
    const [promptSettings, maxTurnsSettings] = await Promise.all([
      GetSetting("chat.default_system_prompt"),
      GetSetting("chat.agentic_max_turns"),
    ]);
    $agenticSystemPrompt.set(promptSettings.value ?? "");
    $agenticMaxTurns.set(maxTurnsSettings.value ?? "4");
  } catch (error) {
    console.error("Failed to fetch agentic settings:", error);
    throw error;
  } finally {
    $isLoadingAgenticSettings.set(false);
  }
};

onMount($providerSettings, () => {
  GetAllProviderSettings();
  return () => { };
});

onMount($webSearchKey, () => {
  GetWebSearchSetting().catch((error) => {
    console.error("Failed to load websearch setting on mount:", error);
  });
  return () => { };
});

onMount($cloudflareScrapeApiKey, () => {
  GetCloudflareScrapeSetting().catch((error) => {
    console.error("Failed to load cloudflare scrape setting on mount:", error);
  });
  return () => { };
});

onMount($agenticSystemPrompt, () => {
  GetAgenticSettings().catch((error) => {
    console.error("Failed to load agentic settings on mount:", error);
  });
  return () => { };
});
