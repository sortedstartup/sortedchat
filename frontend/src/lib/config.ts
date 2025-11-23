export type UIConfig = {
  API_URL: string;
  API_UPLOAD_URL: string;
  GOOGLE_CLIENT_ID: string;
  GOOGLE_OAUTH_URL: string;
  GOOGLE_REDIRECT_URL: string;
};

let config: UIConfig | null = null;

export async function loadUIConfig(): Promise<UIConfig> {
  if (config) {
    console.log("UI config already loaded", JSON.stringify(config));
    return config;
  }

  try {
    const res = await fetch("/ui-config.json");
    if (!res.ok) {
      throw new Error(
        `Failed to load UI config: ${res.status} ${res.statusText}`,
      );
    }
    config = await res.json();
    if (!config) {
      throw new Error("UI Config is null after loading");
    }
    console.log("UI config loaded", JSON.stringify(config));
    return config;
  } catch (error) {
    console.error("Failed to load UI config:", error);
    throw error;
  }
}

export function getUIConfig(): UIConfig | null {
  return config;
}
