import { useTheme } from "@/components/theme-provider";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { UpdateModelsList } from "@/store/setting";
import { useState } from "react";

const Settings = () => {
  const { theme, setTheme } = useTheme();
  const [isSyncingModels, setIsSyncingModels] = useState(false);

  const handleSyncModels = async () => {
    setIsSyncingModels(true);
    try {
      await UpdateModelsList();
    } catch (error) {
      console.error(error);
    } finally {
      setIsSyncingModels(false);
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
            <CardContent>
              <div className="flex items-center justify-between gap-4">
                <div className="space-y-1">
                  <Label className="text-base">Models List</Label>
                  <p className="text-sm text-muted-foreground">
                    Fetch the latest hosted models catalog and update your local models list.
                  </p>
                </div>
                <Button onClick={handleSyncModels} disabled={isSyncingModels}>
                  {isSyncingModels ? "Syncing..." : "Sync Models List"}
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
