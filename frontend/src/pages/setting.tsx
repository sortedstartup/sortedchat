import { useState, useEffect } from "react";
import { useStore } from "@nanostores/react";
import { $settings, saveSettings } from "../store/setting";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
import { Button } from "@/components/ui/button";
import { toast } from "sonner";

const Settings = () => {
  const settings = useStore($settings);
  const [formData, setFormData] = useState<Record<string, string>>({});

  useEffect(() => {
    if (settings) {
      const settingsObj = settings.toObject?.() ?? {};
      setFormData(settingsObj);
    }
  }, [settings]);

  const handleFieldChange = (fieldName: string, value: string) => {
    setFormData(prev => ({
      ...prev,
      [fieldName]: value
    }));
  };

  const handleSave = async () => {
    try {
      const message = await saveSettings(formData);
      toast.success(message);
      
    } catch (error) {
      const errorMessage = error instanceof Error 
        ? error.message 
        : "An unexpected error occurred while saving settings";
        
      toast.error(errorMessage);
      
      console.error("Save failed:", error);
    } 
  };

  return (
    <div className="container mx-auto py-6 px-4 max-w-2xl">
      <h1 className="text-2xl font-semibold mb-4">Settings</h1>

      <Card>
        <CardHeader>
          <CardTitle>Configuration</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          {Object.keys(formData).map(fieldName => (
            <div key={fieldName} className="space-y-2">
              <Label htmlFor={fieldName}>{fieldName}</Label>
              <Input
                id={fieldName}
                value={formData[fieldName] || ""}
                onChange={(e) => handleFieldChange(fieldName, e.target.value)}
              />
            </div>
          ))}
          
          <div className="pt-2">
            <Button 
              onClick={handleSave} 
            >
                Save
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
};

export default Settings;