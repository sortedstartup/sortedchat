// pages/Settings.tsx
import { useState } from "react";
import { useStore } from "@nanostores/react";
import { $settings, saveSettings } from "../store/setting";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";

const Settings = () => {
  const settings = useStore($settings);
  
  const [apiKey, setApiKey] = useState(settings.OPENAI_API_KEY || "");
  const [apiUrl, setApiUrl] = useState(settings.OPENAI_API_URL || "");

  const handleSave = async () => {
    try {
      const message = await saveSettings(apiKey, apiUrl);
      console.log(message);
    } catch (error) {
      console.error("Save failed:", error);
    }
  };

  return (
    <div className="container mx-auto py-6 px-4 max-w-2xl">
      <h1 className="text-2xl font-semibold mb-4">Settings</h1>

      <Card>
        <CardHeader>
          <CardTitle>API Configuration</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="apiKey">OpenAI API Key</Label>
            <Input
              id="apiKey"
              type="password"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="apiUrl">OpenAI API URL</Label>
            <Input
              id="apiUrl"
              type="url"
              value={apiUrl}
              onChange={(e) => setApiUrl(e.target.value)}
            />
          </div>

          <div className="pt-2">
            <Button onClick={handleSave}>Save</Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default Settings;