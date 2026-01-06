
import { useStore } from '@nanostores/react';
import { useState, useEffect } from 'react';
import {
  $llmModels,
  ListLLMModels,
  $isLoadingModels,
  $modelsByProvider
} from '../store/inference';
import {
  $providerSettings,
  $isLoadingProviderSettings,
  GetAllProviderSettings,
  SetProviderSetting
} from '../store/setting';
import { ModelCard } from '@/components/model-card';
import { ProviderSettings } from '../../proto/chatservice';


const Models = () => {
  const models = useStore($llmModels);
  const modelsByProvider = useStore($modelsByProvider);
  const isLoading = useStore($isLoadingModels);
  const providerSettings = useStore($providerSettings);
  const isLoadingProviders = useStore($isLoadingProviderSettings);

  const [selectedProvider, setSelectedProvider] = useState<string | null>(null);
  const [apiUrl, setApiUrl] = useState('');
  const [apiKey, setApiKey] = useState('');
  const [isSaving, setIsSaving] = useState(false);

  // Load data on mount
  useEffect(() => {
    ListLLMModels();
    GetAllProviderSettings();
  }, []);

  // Get unique providers from models
  const providers = Object.keys(modelsByProvider);

  // Auto-select first provider
  useEffect(() => {
    if (providers.length > 0 && !selectedProvider) {
      setSelectedProvider(providers[0]);
    }
  }, [providers, selectedProvider]);

  const handleRefresh = () => {
    ListLLMModels();
    GetAllProviderSettings();
  };

  // Get current provider settings
  const currentProviderSettings = selectedProvider
    ? providerSettings.get(selectedProvider)
    : null;

  // Filter models by selected provider
  const filteredModels = selectedProvider
    ? modelsByProvider[selectedProvider] || []
    : [];

  // Default URLs for providers
  const getDefaultUrl = (provider: string): string => {
    switch (provider) {
      case 'openai':
        return 'https://api.openai.com/v1/chat/completions';
      case 'claude':
        return 'https://api.anthropic.com/v1/messages';
      case 'gemini':
        return 'https://generativelanguage.googleapis.com/v1beta/openai/chat/completions';
      default:
        return '';
    }
  };

  // Update form when provider changes
  useEffect(() => {
    if (selectedProvider) {
      const settings = providerSettings.get(selectedProvider);
      setApiUrl(settings?.api_url || getDefaultUrl(selectedProvider));
      setApiKey(settings?.api_key || '');
    }
  }, [selectedProvider, providerSettings]);

  const handleSaveProvider = async () => {
    if (!selectedProvider) return;

    setIsSaving(true);
    try {
      const settings = new ProviderSettings({
        api_url: apiUrl,
        api_key: apiKey,
        is_enabled: currentProviderSettings?.is_enabled ?? true,
      });

      await SetProviderSetting(selectedProvider, settings);
    } catch (error) {
      console.error('Failed to save provider settings:', error);
    } finally {
      setIsSaving(false);
    }
  };

  if ((isLoading || isLoadingProviders) && models.length === 0) {
    return (
      <div className="h-full overflow-y-auto">
        <div className="container mx-auto px-4 py-8">
          <div className="flex items-center justify-center min-h-64">
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
              <p className="text-muted-foreground">Loading models...</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full flex overflow-hidden">
      {/* Sidebar */}
      <div className="w-64 border-r border-border bg-card overflow-y-auto">
        <div className="p-4 border-b border-border">
          <h2 className="text-lg font-semibold text-foreground">Providers</h2>
        </div>
        <div className="p-2">
          {providers.map((provider) => (
            <button
              key={provider}
              onClick={() => setSelectedProvider(provider)}
              className={`w-full text-left px-4 py-3 rounded-md mb-1 transition-colors ${selectedProvider === provider
                ? 'bg-primary text-primary-foreground'
                : 'hover:bg-accent text-foreground'
                }`}
            >
              <div className="font-medium capitalize">{provider === 'local' ? 'Local (llama.cpp)' : provider}</div>
              <div className="text-xs opacity-70 mt-1">
                {models.filter(m => m.provider === provider).length} models
              </div>
            </button>
          ))}
        </div>
      </div>

      {/* Main Content */}
      <div className="flex-1 overflow-y-auto">
        <div className="container mx-auto px-6 py-8">
          {/* Header */}
          <div className="flex items-center justify-between mb-6">
            <div>
              <h1 className="text-3xl font-bold text-foreground capitalize">
                {selectedProvider === 'local' ? 'Local (llama.cpp)' : selectedProvider} Models
              </h1>
              <p className="text-muted-foreground mt-2">
                {filteredModels.length} models available
              </p>
            </div>

            <button
              onClick={handleRefresh}
              disabled={isLoading || isLoadingProviders}
              className="flex items-center space-x-2 px-4 py-2 bg-primary hover:bg-primary/90 disabled:bg-primary/50 text-primary-foreground rounded-md transition-colors"
            >
              <svg
                className={`w-4 h-4 ${(isLoading || isLoadingProviders) ? 'animate-spin' : ''}`}
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"
                />
              </svg>
              <span>{(isLoading || isLoadingProviders) ? 'Refreshing...' : 'Refresh'}</span>
            </button>
          </div>

          {/* Provider Settings */}
          {selectedProvider && selectedProvider !== 'local' && (
            <div className="bg-card border border-border rounded-lg p-6 mb-6">
              <div className="mb-4">
                <h3 className="text-lg font-semibold text-foreground">Provider Settings</h3>
              </div>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    API URL
                  </label>
                  <input
                    type="text"
                    value={apiUrl}
                    onChange={(e) => setApiUrl(e.target.value)}
                    className="w-full px-3 py-2 bg-background border border-input rounded-md text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder={getDefaultUrl(selectedProvider)}
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-foreground mb-2">
                    API Key
                  </label>
                  <input
                    type="password"
                    value={apiKey}
                    onChange={(e) => setApiKey(e.target.value)}
                    className="w-full px-3 py-2 bg-background border border-input rounded-md text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="Enter API Key"
                  />
                </div>

                <div>
                  <button
                    onClick={handleSaveProvider}
                    disabled={isSaving}
                    className="px-4 py-2 bg-primary hover:bg-primary/90 disabled:bg-primary/50 text-primary-foreground rounded-md transition-colors"
                  >
                    {isSaving ? 'Saving...' : 'Save'}
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Models Grid */}
          {filteredModels.length === 0 ? (
            <div className="text-center py-12">
              <div className="text-muted-foreground/50 mb-4">
                <svg className="w-16 h-16 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
                </svg>
              </div>
              <h3 className="text-lg font-medium text-foreground mb-2">No models found</h3>
              <p className="text-muted-foreground">No models available for this provider.</p>
            </div>
          ) : (
            <div className="grid grid-cols-1 lg:grid-cols-2 xl:grid-cols-3 gap-6">
              {filteredModels.map((model) => (
                <ModelCard key={model.id} model={model} isLocal={selectedProvider === 'local'} />
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default Models;