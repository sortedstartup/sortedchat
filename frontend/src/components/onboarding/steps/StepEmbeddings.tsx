import { useState } from 'react';
import { useStore } from '@nanostores/react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { onboardingStore, onboardingActions } from '@/stores/onboardingStore';

export function StepEmbeddings() {
  const state = useStore(onboardingStore);
  const [isValidating, setIsValidating] = useState(false);
  
  const handleNext = async () => {
    // Simple validation - just check if URL is filled
    if (!state.OLLAMA_URL.trim()) {
      onboardingActions.setValidationError('ollamaUrl', 'Ollama URL is required');
      return;
    }
    
    setIsValidating(true);
    
    // Complete onboarding directly without review step
    try {
      // Import the gRPC client here
      const { SettingServiceClient, SetSettingRequest, Settings, CompleteOnboardingRequest } = await import('../../../../proto/chatservice');
      const { createAuthenticatedClientOptions } = await import('@/lib/auth');
      
      // Create gRPC client with authentication
      const client = new SettingServiceClient(
        import.meta.env.VITE_API_URL || window.location.origin,
        {},
        createAuthenticatedClientOptions()
      );
      
      // Get existing settings first
      const { GetSettingRequest } = await import('../../../../proto/chatservice');
      const getRequest = new GetSettingRequest({});
      const existingResponse = await client.GetSetting(getRequest, {});
      const existingSettings = existingResponse.settings;
      
      // Prepare settings - preserve existing values and only update what's provided
      const settingsData: any = {
        OPENAI_API_KEY: state.OPENAI_API_KEY,
        OLLAMA_URL: state.OLLAMA_URL,
        // Preserve existing API URL if user didn't provide a new one
        OPENAI_API_URL: state.OPENAI_API_URL.trim() || (existingSettings?.OPENAI_API_URL || ''),
      };
      
      const settings = new Settings(settingsData);
      
      // Save settings
      const setSettingRequest = new SetSettingRequest({ settings });
      await client.SetSetting(setSettingRequest, {});
      
      // Complete onboarding
      const completeRequest = new CompleteOnboardingRequest({});
      await client.CompleteOnboarding(completeRequest, {});
      
      // Redirect to main app
      window.location.href = '/';
      
    } catch (error) {
      console.error('Failed to complete onboarding:', error);
      onboardingActions.setValidationError('ollamaUrl', 'Failed to save settings. Please try again.');
    } finally {
      setIsValidating(false);
    }
  };
  
  const handleBack = () => {
    onboardingActions.prevStep();
  };
  
  return (
    <div className="space-y-6">
      <div className="space-y-4">
        <div>
          <Label htmlFor="ollama-url">Ollama Server URL</Label>
          <Input
            id="ollama-url"
            type="url"
            placeholder="http://localhost:11434"
            value={state.OLLAMA_URL}
            onChange={(e) => onboardingActions.setOllamaUrl(e.target.value)}
            className={state.validationErrors.ollamaUrl ? 'border-red-500' : ''}
          />
          {state.validationErrors.ollamaUrl && (
            <p className="text-sm text-red-600 mt-1">{state.validationErrors.ollamaUrl}</p>
          )}
          <p className="text-sm text-gray-500 mt-1">
            URL where your Ollama server is running
          </p>
        </div>
      </div>
      
      <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
        <p className="text-sm text-blue-800 mb-2">
          <strong>Need Ollama?</strong> Download from ollama.ai and run: <code className="bg-blue-100 px-1 rounded">ollama pull nomic-embed-text</code>
        </p>
      </div>
      
      <div className="flex justify-between">
        <Button
          variant="outline"
          onClick={handleBack}
        >
          Back
        </Button>
        
        <Button
          onClick={handleNext}
          disabled={isValidating}
        >
          {isValidating ? 'Saving...' : 'Complete Setup'}
        </Button>
      </div>
    </div>
  );
}
