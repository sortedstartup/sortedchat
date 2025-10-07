import { useState } from 'react';
import { useStore } from '@nanostores/react';
import { Button } from '@/components/ui/button';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2, CheckCircle, AlertCircle } from 'lucide-react';
import { onboardingStore, onboardingActions } from '@/stores/onboardingStore';

// Import the generated gRPC client
import { SettingServiceClient, SetSettingRequest, Settings, CompleteOnboardingRequest } from '../../../../proto/chatservice';

export function StepFinish() {
  const state = useStore(onboardingStore);
  const [isCompleting, setIsCompleting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isCompleted, setIsCompleted] = useState(false);
  
  const handleFinish = async () => {
    setIsCompleting(true);
    setError(null);
    
    try {
      // Create gRPC client
      const client = new SettingServiceClient(
        import.meta.env.VITE_API_URL || window.location.origin
      );
      
      // Prepare settings
      const settings = new Settings({
        OPENAI_API_KEY: state.OPENAI_API_KEY,
        OPENAI_API_URL: state.OPENAI_API_URL,
        OLLAMA_URL: state.OLLAMA_URL,
      });
      
      // Save settings
      const setSettingRequest = new SetSettingRequest({ settings });
      await client.SetSetting(setSettingRequest, {});
      
      // Complete onboarding
      const completeRequest = new CompleteOnboardingRequest({});
      await client.CompleteOnboarding(completeRequest, {});
      
      setIsCompleted(true);
      
      // Redirect to main app after a short delay
      setTimeout(() => {
        window.location.href = '/';
      }, 2000);
      
    } catch (error) {
      console.error('Failed to complete onboarding:', error);
      setError(error instanceof Error ? error.message : 'Failed to save settings');
    } finally {
      setIsCompleting(false);
    }
  };
  
  const handleBack = () => {
    onboardingActions.prevStep();
  };
  
  if (isCompleted) {
    return (
      <div className="text-center space-y-6">
        <div className="flex justify-center">
          <CheckCircle className="h-16 w-16 text-green-500" />
        </div>
        
        <div>
          <h3 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">
            Setup Complete!
          </h3>
          <p className="text-gray-600 dark:text-gray-300">
            Redirecting you to the main application...
          </p>
        </div>
      </div>
    );
  }
  
  return (
    <div className="space-y-6">
      <div className="space-y-4">
        <h3 className="text-lg font-semibold">Review Your Configuration</h3>
        
        <div className="bg-gray-50 dark:bg-gray-800 rounded-lg p-4 space-y-3">
          <div>
            <span className="font-medium">API Key:</span>{' '}
            <span className="font-mono text-sm">
              {state.OPENAI_API_KEY.substring(0, 8)}...
            </span>
          </div>
          {state.OPENAI_API_URL && (
            <div>
              <span className="font-medium">API URL:</span>{' '}
              <span className="font-mono text-sm">{state.OPENAI_API_URL}</span>
            </div>
          )}
          <div>
            <span className="font-medium">Ollama URL:</span>{' '}
            <span className="font-mono text-sm">{state.OLLAMA_URL}</span>
          </div>
        </div>
      </div>
      
      {error && (
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            {error}
          </AlertDescription>
        </Alert>
      )}
      
      <Alert>
        <CheckCircle className="h-4 w-4" />
        <AlertDescription>
          Your settings will be saved securely and you can modify them anytime in the Settings page.
        </AlertDescription>
      </Alert>
      
      <div className="flex justify-between">
        <Button
          variant="outline"
          onClick={handleBack}
          disabled={isCompleting}
        >
          Back
        </Button>
        
        <Button
          onClick={handleFinish}
          disabled={isCompleting}
          className="min-w-[120px]"
        >
          {isCompleting ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Finishing...
            </>
          ) : (
            'Complete Setup'
          )}
        </Button>
      </div>
    </div>
  );
}
