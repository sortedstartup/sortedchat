// store/setting.ts
import {
  Settings,
  GetSettingRequest,
  GetSettingResponse,
  SetSettingRequest,
  SetSettingResponse,
  SettingServiceClient,
  IsFirstBootRequest,
  TestConnectionRequest,
  TestConnectionResponse,
  ConnectionType,
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

// Onboarding state
export const $onboardingStep = atom<number>(0);
export const $onboardingData = atom<{
  OPENAI_API_KEY: string;
  OPENAI_API_URL: string;
  OLLAMA_URL: string;
}>({
  OPENAI_API_KEY: "",
  OPENAI_API_URL: "https://api.openai.com/v1/chat/completions",
  OLLAMA_URL: "http://localhost:11434/v1/embeddings",
});

export const saveSettings = async (
  formData: Record<string, string>,
): Promise<string> => {
  try {
    const settings = new Settings(formData);

    const req = new SetSettingRequest({ settings });
    const res: SetSettingResponse = await getClient().SetSetting(req, {});

    $settings.set(settings);

    return res.message ?? "Settings saved successfully";
  } catch (error) {
    console.error("Failed to save settings:", error);
    throw new Error("Failed to save settings");
  }
};

// Onboarding actions
export const onboardingActions = {
  setApiKey: (key: string) => {
    const data = $onboardingData.get();
    $onboardingData.set({ ...data, OPENAI_API_KEY: key });
  },

  setApiUrl: (url: string) => {
    const data = $onboardingData.get();
    $onboardingData.set({ ...data, OPENAI_API_URL: url });
  },

  setOllamaUrl: (url: string) => {
    const data = $onboardingData.get();
    $onboardingData.set({ ...data, OLLAMA_URL: url });
  },

  nextStep: () => {
    const current = $onboardingStep.get();
    if (current < 1) {
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
      const settings = new Settings(data);

      // Save settings
      const setReq = new SetSettingRequest({ settings });
      await getClient().SetSetting(setReq, {});

      $settings.set(settings);

      // Force a full page reload to ensure isFirstBoot check runs fresh
      window.location.replace("/");
    } catch (error) {
      console.error("Failed to complete onboarding:", error);
      throw new Error("Failed to complete onboarding");
    }
  },
};

const getSetting = async () => {
  try {
    const req = new GetSettingRequest({});
    const res: GetSettingResponse = await getClient().GetSetting(req, {});
    if (res.settings) {
      $settings.set(res.settings);
    }
  } catch (error) {
    console.error("Failed to fetch settings:", error);
  }
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

onMount($settings, () => {
  getSetting();
  return () => {};
});
