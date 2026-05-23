import { useStore } from "@nanostores/react";
import { useState } from "react";
import { useTheme } from "@/components/theme-provider";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  $cloudflareScrapeApiKey,
  $cloudflareScrapeApiUrl,
  $isLoadingCloudflareScrape,
  $isLoadingWebSearch,
  $webSearchApiUrl,
  $webSearchCost,
  $webSearchKey,
  SetSetting,
} from "@/store/setting";
import { toast } from "sonner";

const Settings = () => {
  const { theme, setTheme } = useTheme();
  const webSearchKey = useStore($webSearchKey);
  const webSearchApiUrl = useStore($webSearchApiUrl);
  const webSearchCost = useStore($webSearchCost);
  const cloudflareScrapeApiUrl = useStore($cloudflareScrapeApiUrl);
  const cloudflareScrapeApiKey = useStore($cloudflareScrapeApiKey);
  const isLoadingWebSearch = useStore($isLoadingWebSearch);
  const isLoadingCloudflareScrape = useStore($isLoadingCloudflareScrape);
  const [isSavingWebSearch, setIsSavingWebSearch] = useState(false);
  const [isSavingCloudflareScrape, setIsSavingCloudflareScrape] = useState(false);

  const saveWebSearchSettings = async () => {
    setIsSavingWebSearch(true);
    try {
      const braveSearchApiUrl = webSearchApiUrl.trim() || "https://api.search.brave.com/res/v1/web/search";
      const braveSearchApiKey = webSearchKey.trim();
      const braveSearchCost = webSearchCost.trim() || "0.005";
      const message = await SetSetting("tool.websearch.brave", {
        apiUrl: braveSearchApiUrl,
        apiKey: braveSearchApiKey,
        cost: braveSearchCost,
      });
      $webSearchApiUrl.set(braveSearchApiUrl);
      $webSearchKey.set(braveSearchApiKey);
      $webSearchCost.set(braveSearchCost);
      toast.success(message);
    } catch (error) {
      console.error("Failed to save websearch settings:", error);
      toast.error("Failed to save web search settings");
    } finally {
      setIsSavingWebSearch(false);
    }
  };

  const saveCloudflareScrapeSettings = async () => {
    setIsSavingCloudflareScrape(true);
    try {
      const apiUrl = cloudflareScrapeApiUrl.trim();
      const apiKey = cloudflareScrapeApiKey.trim();
      const message = await SetSetting("tool.scrape.cloudflare", {
        apiUrl,
        apiKey,
      });
      $cloudflareScrapeApiUrl.set(apiUrl);
      $cloudflareScrapeApiKey.set(apiKey);
      toast.success(message);
    } catch (error) {
      console.error("Failed to save cloudflare scrape settings:", error);
      toast.error("Failed to save Cloudflare scrape settings");
    } finally {
      setIsSavingCloudflareScrape(false);
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

              <div className="space-y-2">
                <Label htmlFor="brave-search-cost">Brave Search Cost Per Request (USD)</Label>
                <Input
                  id="brave-search-cost"
                  type="number"
                  min="0"
                  value={webSearchCost}
                  onChange={(e) => $webSearchCost.set(e.target.value)}
                  placeholder="Enter cost per request"
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

          <Card>
            <CardContent className="space-y-4">
              <div className="space-y-1">
                <Label className="text-base">Cloudflare Scrape</Label>
                <p className="text-sm text-muted-foreground">
                  Save your Cloudflare Browser Rendering API URL and API key to enable page scraping for agentic chat.
                </p>
              </div>

              <div className="space-y-2">
                <Label htmlFor="cloudflare-scrape-api-url">Cloudflare Scrape API URL</Label>
                <Input
                  id="cloudflare-scrape-api-url"
                  type="text"
                  value={cloudflareScrapeApiUrl}
                  onChange={(e) => $cloudflareScrapeApiUrl.set(e.target.value)}
                  placeholder="https://api.cloudflare.com/client/v4/accounts/<accountId>/browser-rendering/markdown"
                  disabled={isLoadingCloudflareScrape || isSavingCloudflareScrape}
                />
              </div>

              <div className="space-y-2">
                <Label htmlFor="cloudflare-scrape-api-key">Cloudflare Scrape API Key</Label>
                <Input
                  id="cloudflare-scrape-api-key"
                  type="password"
                  value={cloudflareScrapeApiKey}
                  onChange={(e) => $cloudflareScrapeApiKey.set(e.target.value)}
                  placeholder="Enter Cloudflare Browser Rendering API key"
                  disabled={isLoadingCloudflareScrape || isSavingCloudflareScrape}
                />
              </div>

              <div className="flex justify-end">
                <Button onClick={saveCloudflareScrapeSettings} disabled={isLoadingCloudflareScrape || isSavingCloudflareScrape}>
                  {isSavingCloudflareScrape ? "Saving..." : "Save"}
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
