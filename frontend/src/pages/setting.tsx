import { useStore } from "@nanostores/react";
import { useState } from "react";
import { useTheme } from "@/components/theme-provider";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { $isLoadingWebSearch, $webSearchKey, $webSearchApiUrl, SetSetting } from "@/store/setting";
import { toast } from "sonner";

const Settings = () => {
  const { theme, setTheme } = useTheme();
  const webSearchKey = useStore($webSearchKey);
  const webSearchApiUrl = useStore($webSearchApiUrl);
  const isLoadingWebSearch = useStore($isLoadingWebSearch);
  const [isSavingWebSearch, setIsSavingWebSearch] = useState(false);

  const saveWebSearchSettings = async () => {
    setIsSavingWebSearch(true);
    try {
      const braveSearchApiUrl = webSearchApiUrl.trim() || "https://api.search.brave.com/res/v1/web/search";
      const braveSearchApiKey = webSearchKey.trim();
      const message = await SetSetting("tool.websearch.brave", {
        apiUrl: braveSearchApiUrl,
        apiKey: braveSearchApiKey,
      });
      $webSearchApiUrl.set(braveSearchApiUrl);
      $webSearchKey.set(braveSearchApiKey);
      toast.success(message);
    } catch (error) {
      console.error("Failed to save websearch settings:", error);
      toast.error("Failed to save web search settings");
    } finally {
      setIsSavingWebSearch(false);
    }
  };

  return (
    <div className="h-full overflow-y-auto bg-background">
      <div className="container mx-auto px-4 py-8 max-w-4xl">
        <h1 className="text-3xl font-bold text-foreground mb-6">Settings</h1>

        <div className="space-y-6">
          <Card>
            <CardContent>
              <div className="flex items-center justify-between">
                <div className="space-y-1">
                  <Label className="text-base">Theme</Label>
                  <p className="text-sm text-muted-foreground">
                    Select your preferred theme mode.
                  </p>
                </div>
                <div className="w-[200px]">
                  <select
                    value={theme}
                    onChange={(e) => setTheme(e.target.value as any)}
                    className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    <option value="light">Light</option>
                    <option value="dark">Dark</option>
                  </select>
                </div>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardContent className="space-y-4">
              <div className="space-y-1">
                <Label className="text-base">Web Search</Label>
                <p className="text-sm text-muted-foreground">
                  Save your Brave Search API URL and API key for web-enabled chat responses.
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="brave-search-api-url">Brave Search API URL</Label>
                <Input
                  id="brave-search-api-url"
                  type="text"
                  value={webSearchApiUrl}
                  onChange={(e) => $webSearchApiUrl.set(e.target.value)}
                  placeholder="Enter Brave Search API URL"
                  disabled={isLoadingWebSearch || isSavingWebSearch}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="brave-search-api-key">Brave Search API Key</Label>
                <Input
                  id="brave-search-api-key"
                  type="password"
                  value={webSearchKey}
                  onChange={(e) => $webSearchKey.set(e.target.value)}
                  placeholder="Enter Brave Search API key"
                  disabled={isLoadingWebSearch || isSavingWebSearch}
                />
              </div>

              <div className="flex justify-end">
                <Button onClick={saveWebSearchSettings} disabled={isLoadingWebSearch || isSavingWebSearch}>
                  {isSavingWebSearch ? "Saving..." : "Save"}
                </Button>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  );
};

export default Settings;
