import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { onboardingActions } from '@/store/setting';
import { OpenaiProvider } from '@/components/providers/openai-provider';
import { ClaudeProvider } from '@/components/providers/claude-provider';
import { GeminiProvider } from '@/components/providers/gemini-provider';

type Provider = 'openai' | 'claude' | 'gemini';

export function StepRemote() {
  const [isValidating, setIsValidating] = useState(false);
  const [selectedProvider, setSelectedProvider] = useState<Provider>('openai');
  const [error, setError] = useState<string>('');

  const handleNext = async () => {
    setIsValidating(true);
    setError('');
    try {
      await onboardingActions.completeOnboarding();
    } catch (error) {
      console.error('Failed to complete onboarding:', error);
      setError(error instanceof Error ? error.message : 'Failed to save settings. Please try again.');
      setIsValidating(false);
    }
  };

  const handleBack = () => {
    onboardingActions.prevStep();
  };

  const renderProviderContent = () => {
    switch (selectedProvider) {
      case 'openai':
        return <OpenaiProvider
          onApiKeyChange={onboardingActions.setOpenaiApiKey}
          onApiUrlChange={onboardingActions.setOpenaiApiUrl}
        />;
      case 'claude':
        return <ClaudeProvider
          onApiKeyChange={onboardingActions.setClaudeApiKey}
          onApiUrlChange={onboardingActions.setClaudeApiUrl}
        />;
      case 'gemini':
        return <GeminiProvider
          onApiKeyChange={onboardingActions.setGeminiApiKey}
          onApiUrlChange={onboardingActions.setGeminiApiUrl}
        />;
      default:
        return null;
    }
  };

  const providers: Provider[] = ['openai', 'claude', 'gemini'];

  return (
    <div className="flex flex-col h-full">
      <div className="flex-1 flex overflow-hidden border rounded-lg bg-background mb-4">
        {/* Sidebar */}
        <div className="w-48 border-r border-border bg-card overflow-y-auto flex-shrink-0">
          <div className="p-3 border-b border-border">
            <h2 className="text-sm font-semibold text-foreground">Providers</h2>
          </div>
          <div className="p-2">
            {providers.map((provider) => (
              <button
                key={provider}
                onClick={() => setSelectedProvider(provider)}
                className={`w-full text-left px-3 py-2 rounded-md mb-1 transition-colors text-sm ${selectedProvider === provider
                  ? 'bg-primary text-primary-foreground'
                  : 'hover:bg-accent text-foreground'
                  }`}
              >
                <div className="font-medium capitalize">{provider}</div>
              </button>
            ))}
          </div>
        </div>

        {/* Main Content */}
        <div className="flex-1 overflow-y-auto bg-background">
          {renderProviderContent()}
        </div>
      </div>

      <div className="bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-800 rounded-lg p-4 mb-4">
        <p className="text-sm text-blue-800 dark:text-blue-200">
          <strong>Note:</strong> Please save your settings in each provider tab before completing setup.
        </p>
      </div>

      {error && (
        <div className="bg-red-50 dark:bg-red-950 border border-red-200 dark:border-red-800 rounded-lg p-4 mb-4">
          <p className="text-sm text-red-800 dark:text-red-200">
            <strong>Error:</strong> {error}
          </p>
        </div>
      )}

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
          {isValidating ? 'Finishing...' : 'Complete Setup'}
        </Button>
      </div>
    </div>
  );
}