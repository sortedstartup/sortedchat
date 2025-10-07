import { useState } from 'react';
import { useStore } from '@nanostores/react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Loader2, AlertCircle } from 'lucide-react';
import { onboardingStore, onboardingActions, canProceed } from '@/stores/onboardingStore';
import { validateCurrentStep } from '@/lib/validators/onboarding';

export function StepApi() {
  const state = useStore(onboardingStore);
  const canProceedToNext = useStore(canProceed);
  const [isValidating, setIsValidating] = useState(false);
  
  const handleNext = async () => {
    setIsValidating(true);
    onboardingActions.setValidating(true);
    onboardingActions.clearValidationErrors();
    
    try {
      const errors = await validateCurrentStep(
        0,
        state.provider,
        state.OPENAI_API_KEY,
        state.OPENAI_API_URL,
        state.OLLAMA_URL
      );
      
      if (Object.keys(errors).length > 0) {
        Object.entries(errors).forEach(([field, error]) => {
          onboardingActions.setValidationError(field as any, error);
        });
      } else {
        onboardingActions.nextStep();
      }
    } catch (error) {
      onboardingActions.setValidationError('apiKey', 'Validation failed. Please try again.');
    } finally {
      setIsValidating(false);
      onboardingActions.setValidating(false);
    }
  };
  
  return (
    <div className="space-y-6">
      <div className="space-y-4">
        <div>
          <Label className="text-base font-medium">Choose your LLM provider</Label>
          <RadioGroup
            value={state.provider}
            onValueChange={(value) => onboardingActions.setProvider(value as 'openai' | 'litellm')}
            className="mt-2"
          >
            <div className="flex items-center space-x-2">
              <RadioGroupItem value="openai" id="openai" />
              <Label htmlFor="openai" className="cursor-pointer">
                OpenAI (Direct API)
              </Label>
            </div>
            <div className="flex items-center space-x-2">
              <RadioGroupItem value="litellm" id="litellm" />
              <Label htmlFor="litellm" className="cursor-pointer">
                LiteLLM (Proxy Server)
              </Label>
            </div>
          </RadioGroup>
        </div>
        
        {state.provider === 'openai' ? (
          <div className="space-y-4">
            <div>
              <Label htmlFor="api-key">OpenAI API Key *</Label>
              <Input
                id="api-key"
                type="password"
                placeholder="sk-..."
                value={state.OPENAI_API_KEY}
                onChange={(e) => onboardingActions.setApiKey(e.target.value)}
                className={state.validationErrors.apiKey ? 'border-red-500' : ''}
              />
              {state.validationErrors.apiKey && (
                <p className="text-sm text-red-600 mt-1">{state.validationErrors.apiKey}</p>
              )}
            </div>
            
            <div>
              <Label htmlFor="api-url">Custom API URL (optional)</Label>
              <Input
                id="api-url"
                type="url"
                placeholder="https://api.openai.com/v1"
                value={state.OPENAI_API_URL}
                onChange={(e) => onboardingActions.setApiUrl(e.target.value)}
                className={state.validationErrors.apiUrl ? 'border-red-500' : ''}
              />
              {state.validationErrors.apiUrl && (
                <p className="text-sm text-red-600 mt-1">{state.validationErrors.apiUrl}</p>
              )}
              <p className="text-sm text-gray-500 mt-1">
                Leave empty to use the default OpenAI API
              </p>
            </div>
          </div>
        ) : (
          <div>
            <Label htmlFor="litellm-url">LiteLLM Proxy URL *</Label>
            <Input
              id="litellm-url"
              type="url"
              placeholder="http://localhost:4000"
              value={state.OPENAI_API_URL}
              onChange={(e) => onboardingActions.setApiUrl(e.target.value)}
              className={state.validationErrors.apiUrl ? 'border-red-500' : ''}
            />
            {state.validationErrors.apiUrl && (
              <p className="text-sm text-red-600 mt-1">{state.validationErrors.apiUrl}</p>
            )}
            <p className="text-sm text-gray-500 mt-1">
              URL of your running LiteLLM proxy server
            </p>
          </div>
        )}
      </div>
      
      {state.provider === 'openai' && (
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Your API key will be stored locally and used to make requests to OpenAI's servers.
            Make sure you have sufficient credits in your OpenAI account.
          </AlertDescription>
        </Alert>
      )}
      
      {state.provider === 'litellm' && (
        <Alert>
          <AlertCircle className="h-4 w-4" />
          <AlertDescription>
            Make sure your LiteLLM proxy server is running and accessible.
            LiteLLM allows you to use multiple LLM providers through a unified interface.
          </AlertDescription>
        </Alert>
      )}
      
      <div className="flex justify-end">
        <Button
          onClick={handleNext}
          disabled={!canProceedToNext || isValidating}
          className="min-w-[100px]"
        >
          {isValidating ? (
            <>
              <Loader2 className="mr-2 h-4 w-4 animate-spin" />
              Validating...
            </>
          ) : (
            'Next'
          )}
        </Button>
      </div>
    </div>
  );
}
