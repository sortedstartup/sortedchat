import { useStore } from '@nanostores/react';
import { useEffect } from 'react';
import {
  $llmModels,
  ListLLMModels,
  $isLoadingModels
} from '../store/inference';
import {
  $providerSettings,
  $isLoadingProviderSettings,
  GetAllProviderSettings,
  SetProviderSetting
} from '../store/setting';
import { ProviderSettings } from '../../proto/chatservice';

const ModelProviders = () => {
  const models = useStore($llmModels);
  const isLoadingModels = useStore($isLoadingModels);
  const providerSettings = useStore($providerSettings);
  const isLoadingProviders = useStore($isLoadingProviderSettings);

  useEffect(() => {
    ListLLMModels();
    GetAllProviderSettings();
  }, []);

  // Get unique providers from models with counts
  const providers = Array.from(
    models.reduce((acc, model) => {
      if (model.provider) {
        const existing = acc.get(model.provider);
        acc.set(model.provider, (existing || 0) + 1);
      }
      return acc;
    }, new Map<string, number>())
  ).map(([name, count]) => ({ name, count }));

  const handleToggleProvider = async (providerName: string, currentEnabled: boolean) => {
    try {
      const currentSettings = providerSettings.get(providerName);

      const settings = new ProviderSettings({
        api_url: currentSettings?.api_url || '',
        api_key: currentSettings?.api_key || '',
        is_enabled: !currentEnabled,
      });

      await SetProviderSetting(providerName, settings);
    } catch (error) {
      console.error('Failed to toggle provider:', error);
    }
  };

  const isLoading = isLoadingModels || isLoadingProviders;

  if (isLoading && providers.length === 0) {
    return (
      <div className="h-full overflow-y-auto">
        <div className="container mx-auto px-4 py-8">
          <div className="flex items-center justify-center min-h-64">
            <div className="text-center">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
              <p className="text-muted-foreground">Loading providers...</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="h-full overflow-y-auto">
      <div className="container mx-auto px-4 py-8">
        <div className="flex items-center justify-between mb-8">
          <div>
            <h1 className="text-3xl font-bold text-foreground">Model Providers</h1>
            <p className="text-muted-foreground mt-2">
              Manage AI model providers and their settings
            </p>
          </div>

          <button
            onClick={() => {
              ListLLMModels();
              GetAllProviderSettings();
            }}
            disabled={isLoading}
            className="flex items-center space-x-2 px-4 py-2 bg-primary hover:bg-primary/90 disabled:bg-primary/50 text-primary-foreground rounded-md transition-colors"
          >
            <svg
              className={`w-4 h-4 ${isLoading ? 'animate-spin' : ''}`}
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
            <span>{isLoading ? 'Refreshing...' : 'Refresh'}</span>
          </button>
        </div>

        {providers.length === 0 ? (
          <div className="text-center py-12">
            <div className="text-muted-foreground/50 mb-4">
              <svg className="w-16 h-16 mx-auto" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={1.5} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
              </svg>
            </div>
            <h3 className="text-lg font-medium text-foreground mb-2">No providers found</h3>
            <p className="text-muted-foreground">No model providers are available.</p>
          </div>
        ) : (
          <div className="space-y-3">
            {providers.map((provider) => {
              const settings = providerSettings.get(provider.name);
              const isEnabled = settings?.is_enabled ?? false;

              return (
                <div
                  key={provider.name}
                  className="bg-card border border-border rounded-lg p-5 hover:shadow-md transition-shadow"
                >
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-4 flex-1">
                      <div className="flex-1">
                        <h3 className="text-lg font-semibold text-foreground capitalize">
                          {provider.name === 'local' ? 'Llama.cpp' : provider.name}
                        </h3>
                        <p className="text-sm text-muted-foreground mt-1">
                          {provider.count} {provider.count === 1 ? 'Model' : 'Models'}
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center space-x-3">
                      <button
                        onClick={() => handleToggleProvider(provider.name, isEnabled)}
                        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2 ${isEnabled ? 'bg-primary' : 'bg-muted'
                          }`}
                        role="switch"
                        aria-checked={isEnabled}
                      >
                        <span
                          className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${isEnabled ? 'translate-x-6' : 'translate-x-1'
                            }`}
                        />
                      </button>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
};

export default ModelProviders;

