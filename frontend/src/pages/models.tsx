
import { useStore } from '@nanostores/react';
import { useState, useEffect } from 'react';
import {
  $llmModels,
  ListLLMModels,
  $isLoadingModels,
  $modelsByProvider
} from '../store/inference';
import {
  $isLoadingProviderSettings,
  GetAllProviderSettings,
  $providerSettings,
} from '../store/setting';
import { OpenaiProvider } from '@/components/providers/openai-provider';
import { ClaudeProvider } from '@/components/providers/claude-provider';
import { GeminiProvider } from '@/components/providers/gemini-provider';
import { LocalProvider } from '@/components/providers/local-provider';
import { ProviderView } from '@/components/providers/provider-view';
import { AddProviderDialog } from '@/components/add-provider-dialog';
import { PlusIcon } from 'lucide-react';


const Models = () => {
  const models = useStore($llmModels);
  const modelsByProvider = useStore($modelsByProvider);
  const isLoading = useStore($isLoadingModels);
  const isLoadingProviders = useStore($isLoadingProviderSettings);

  const [selectedProvider, setSelectedProvider] = useState<string | null>(null);
  const [isAddProviderModalOpen, setIsAddProviderModalOpen] = useState(false);
  const providerSettingsMap = useStore($providerSettings);

  // Load data on mount
  useEffect(() => {
    ListLLMModels();
    GetAllProviderSettings();
  }, []);

  // Get unique providers from models AND settings
  const providers = Array.from(new Set([
    ...Object.keys(modelsByProvider),
    ...Array.from(providerSettingsMap.keys())
  ]));

  // Auto-select first provider
  useEffect(() => {
    if (providers.length > 0 && !selectedProvider) {
      setSelectedProvider(providers[0]);
    }
  }, [providers, selectedProvider]);

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

  const renderProviderContent = () => {
    if (!selectedProvider) return null;

    switch (selectedProvider) {
      case 'openai':
        return <OpenaiProvider />;
      case 'claude':
        return <ClaudeProvider />;
      case 'gemini':
        return <GeminiProvider />;
      case 'local':
        return <LocalProvider />;
      default:
        return <ProviderView providerName={selectedProvider} />;
    }
  };

  return (
    <div className="h-full flex overflow-hidden">
      {/* Sidebar */}
      <div className="w-64 border-r border-border bg-card overflow-y-auto">
        <div className="p-4 border-b border-border flex items-center justify-between">
          <h2 className="text-lg font-semibold text-foreground">Providers</h2>
          <button
            onClick={() => setIsAddProviderModalOpen(true)}
            className="p-1.5 rounded-md hover:bg-accent text-muted-foreground hover:text-foreground transition-colors"
            title="Add Provider"
          >
            <PlusIcon size={18} />
          </button>
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
        {renderProviderContent()}
      </div>

      <AddProviderDialog
        isOpen={isAddProviderModalOpen}
        onOpenChange={setIsAddProviderModalOpen}
      />
    </div>
  );
};

export default Models;