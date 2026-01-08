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
      <div className="mb-8 text-center max-w-4xl mx-auto">
        <h2 className="text-3xl font-extrabold mb-3 bg-clip-text text-transparent bg-gradient-to-r from-purple-600 to-pink-500">
          Connect Remote Providers
        </h2>
        <p className="text-gray-600 dark:text-gray-400 text-lg leading-relaxed mb-8">
          You can use any of these remote models by getting their API keys from their respective providers.
        </p>

        {/* API Key Links */}
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-2">
          <a 
            href="https://aistudio.google.com/api-keys" 
            target="_blank" 
            rel="noopener noreferrer"
            className="group p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-card hover:border-blue-500 transition-all text-left shadow-sm hover:shadow-md"
          >
            <div className="flex items-center justify-between mb-2">
              <span className="font-bold text-sm text-foreground">Google Gemini</span>
              <svg className="w-4 h-4 text-gray-400 group-hover:text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
              </svg>
            </div>
            <p className="text-xs text-muted-foreground">Get your Gemini API key from Google AI Studio.</p>
          </a>

          <a 
            href="https://platform.openai.com/api-keys" 
            target="_blank" 
            rel="noopener noreferrer"
            className="group p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-card hover:border-emerald-500 transition-all text-left shadow-sm hover:shadow-md"
          >
            <div className="flex items-center justify-between mb-2">
              <span className="font-bold text-sm text-foreground">OpenAI</span>
              <svg className="w-4 h-4 text-gray-400 group-hover:text-emerald-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
              </svg>
            </div>
            <p className="text-xs text-muted-foreground">Access GPT-4o, o1, and more via OpenAI platform.</p>
          </a>

          <a 
            href="https://platform.claude.com/dashboard" 
            target="_blank" 
            rel="noopener noreferrer"
            className="group p-4 rounded-xl border border-gray-200 dark:border-gray-800 bg-card hover:border-orange-500 transition-all text-left shadow-sm hover:shadow-md"
          >
            <div className="flex items-center justify-between mb-2">
              <span className="font-bold text-sm text-foreground">Anthropic Claude</span>
              <svg className="w-4 h-4 text-gray-400 group-hover:text-orange-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M10 6H6a2 2 0 00-2 2v10a2 2 0 002 2h10a2 2 0 002-2v-4M14 4h6m0 0v6m0-6L10 14" />
              </svg>
            </div>
            <p className="text-xs text-muted-foreground">Use Claude 3.5 Sonnet and Opus models.</p>
          </a>
        </div>
      </div>

      <div className="flex-1 flex overflow-hidden border rounded-2xl bg-background/50 backdrop-blur-sm shadow-inner mb-6">
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