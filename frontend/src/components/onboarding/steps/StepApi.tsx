import { useState } from 'react';
import { useStore } from '@nanostores/react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { $onboardingData, onboardingActions } from '@/store/setting';

type Provider = 'openai' | 'claude' | 'gemini';

export function StepApi() {
  const data = useStore($onboardingData);
  const [isValidating, setIsValidating] = useState(false);
  const [selectedProvider, setSelectedProvider] = useState<Provider>('openai');
  const [error, setError] = useState<string>('');

  const handleNext = async () => {
    setIsValidating(true);
    setError('');
    try {
      await onboardingActions.completeOnboarding();
      // If we reach here without redirect, something went wrong
    } catch (error) {
      console.error('Failed to complete onboarding:', error);
      setError(error instanceof Error ? error.message : 'Failed to save settings. Please try again.');
      setIsValidating(false);
    }
  };

  const handleBack = () => {
    onboardingActions.prevStep();
  };

  const getDefaultUrl = (provider: Provider): string => {
    switch (provider) {
      case 'openai':
        return 'https://api.openai.com/v1/chat/completions';
      case 'claude':
        return 'https://api.anthropic.com/v1/messages';
      case 'gemini':
        return 'https://generativelanguage.googleapis.com/v1beta/openai/chat/completions';
    }
  };

  const providerConfigs = {
    openai: {
      name: 'OpenAI',
      apiKey: data.OPENAI_API_KEY,
      apiUrl: data.OPENAI_API_URL,
      setApiKey: onboardingActions.setOpenaiApiKey,
      setApiUrl: onboardingActions.setOpenaiApiUrl,
    },
    claude: {
      name: 'Claude',
      apiKey: data.CLAUDE_API_KEY,
      apiUrl: data.CLAUDE_API_URL,
      setApiKey: onboardingActions.setClaudeApiKey,
      setApiUrl: onboardingActions.setClaudeApiUrl,
    },
    gemini: {
      name: 'Gemini',
      apiKey: data.GEMINI_API_KEY,
      apiUrl: data.GEMINI_API_URL,
      setApiKey: onboardingActions.setGeminiApiKey,
      setApiUrl: onboardingActions.setGeminiApiUrl,
    },
  };

  const currentProvider = providerConfigs[selectedProvider];

  return (
    <div className="space-y-6">
      {/* Provider Tabs */}
      <div className="flex gap-2 border-b">
        {(Object.keys(providerConfigs) as Provider[]).map((provider) => (
          <button
            key={provider}
            onClick={() => setSelectedProvider(provider)}
            className={`px-4 py-2 font-medium transition-colors ${
              selectedProvider === provider
                ? 'border-b-2 border-primary text-primary'
                : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            {providerConfigs[provider].name}
          </button>
        ))}
      </div>

      {/* Provider Settings Form */}
      <div className="space-y-4 py-4">
        <div>
          <Label htmlFor="api-key" className="text-base font-semibold">
            API Key (optional)
          </Label>
          <Input
            id="api-key"
            type="password"
            placeholder="sk-..."
            value={currentProvider.apiKey}
            onChange={(e) => currentProvider.setApiKey(e.target.value)}
            className="mt-2"
          />
          <p className="text-sm text-muted-foreground mt-1">
            Enter your {currentProvider.name} API key to enable this provider
          </p>
        </div>

        <div>
          <Label htmlFor="api-url" className="text-base font-semibold">
            API URL (optional)
          </Label>
          <Input
            id="api-url"
            type="url"
            placeholder={getDefaultUrl(selectedProvider)}
            value={currentProvider.apiUrl}
            onChange={(e) => currentProvider.setApiUrl(e.target.value)}
            className="mt-2"
          />
          <p className="text-sm text-muted-foreground mt-1">
            Custom API endpoint (leave empty for default)
          </p>
        </div>
      </div>

      <div className="bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-800 rounded-lg p-4">
        <p className="text-sm text-blue-800 dark:text-blue-200">
          <strong>Note:</strong> All API keys are stored locally and never shared. You can add or update these later in settings.
        </p>
      </div>

      {error && (
        <div className="bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg p-4">
          <p className="text-sm text-red-800 dark:text-red-200">
            <strong>Error:</strong> {error}
          </p>
        </div>
      )}

      <div className="flex justify-between pt-4">
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