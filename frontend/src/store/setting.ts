// store/setting.ts
import {
  Settings,
  GetSettingRequest,
  GetSettingResponse,
  SetSettingRequest,
  SetSettingResponse,
  SettingServiceClient,
} from "../../proto/chatservice";
import { atom, onMount } from "nanostores";

const client = new SettingServiceClient(import.meta.env.VITE_API_URL);

export const $settings = atom<Settings>(new Settings({}));

export const saveSettings = async (formData: Record<string, string>): Promise<string> => {
  try {
    const settings = new Settings(formData);
    
    const req = new SetSettingRequest({ settings });
    const res: SetSettingResponse = await client.SetSetting(req, {});
    
    $settings.set(settings);
    
    return res.message ?? "Settings saved successfully";
  } catch (error) {
    console.error("Failed to save settings:", error);
    throw new Error("Failed to save settings");
  }
};

const getSetting = async () => {
  try {
    const req = new GetSettingRequest({});
    const res: GetSettingResponse = await client.GetSetting(req, {});
    if (res.settings) {
      $settings.set(res.settings);
    }
  } catch (error) {
    console.error("Failed to fetch settings:", error);
  }
};

onMount($settings, () => {
  getSetting();
  return () => {};
});