import {
  Settings,
  GetConfigRequest,
  GetConfigResponse,
  SetConfigRequest,
  SetConfigResponse,
  ConfigServiceClient,
} from "../../proto/chatservice";
import { atom, onMount } from "nanostores";

const client = new ConfigServiceClient(import.meta.env.VITE_API_URL);

export const $settings = atom<Settings>(new Settings({}));

export const saveSettings = async (apiKey: string, apiUrl: string): Promise<string> => {
  try {
    const settings = new Settings({
      OPENAI_API_KEY: apiKey,
      OPENAI_API_URL: apiUrl,
    });
    
    const req = new SetConfigRequest({ settings });
    const res: SetConfigResponse = await client.SetConfig(req, {});
    
    $settings.set(settings);
    
    return res.message
  } catch (error) {
    console.error("Failed to save settings:", error);
    throw new Error("Failed to save settings");
  }
};

const getConfig = async () => {
  try {
    const req = new GetConfigRequest({});
    const res: GetConfigResponse = await client.GetConfig(req, {});
    if (res.settings) {
      $settings.set(res.settings);
    }
  } catch (error) {
    console.error("Failed to fetch settings:", error);
  }
};

onMount($settings, () => {
  getConfig();
  return () => {};
});