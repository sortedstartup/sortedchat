import { useStore } from '@nanostores/react';
import { useState, useEffect } from 'react';
import {
    $modelsByProvider,
    $isLoadingModels,
    ListLLMModels
} from '../../store/inference';
import {
    $providerSettings,
    SetProviderSetting,
    GetAllProviderSettings
} from '../../store/setting';
import { ModelCard } from '@/components/model-card';
import { ProviderSettings } from '../../../proto/chatservice';
import { AddModelDialog } from '@/components/add-model-dialog';

interface ProviderViewProps {
    providerName: string;
    displayName?: string;
    defaultApiUrl?: string;
    onApiKeyChange?: (key: string) => void;
    onApiUrlChange?: (url: string) => void;
    compact?: boolean;
}

export const ProviderView = ({
    providerName,
    displayName,
    defaultApiUrl = '',
    onApiKeyChange,
    onApiUrlChange,
    compact = false
}: ProviderViewProps) => {
    const modelsByProvider = useStore($modelsByProvider);
    const providerSettings = useStore($providerSettings);
    const isLoading = useStore($isLoadingModels);

    const [apiUrl, setApiUrl] = useState('');
    const [apiKey, setApiKey] = useState('');
    const [isSaving, setIsSaving] = useState(false);
    const [isAddModalOpen, setIsAddModalOpen] = useState(false);

    // Get current provider settings
    const currentProviderSettings = providerSettings.get(providerName);

    // Filter models by provider
    const filteredModels = modelsByProvider[providerName] || [];

    // Update form when provider settings change
    useEffect(() => {
        const settings = providerSettings.get(providerName);
        setApiUrl(settings?.api_url || defaultApiUrl);
        setApiKey(settings?.api_key || '');
    }, [providerName, providerSettings, defaultApiUrl]);

    const handleApiUrlChange = (value: string) => {
        setApiUrl(value);
        onApiUrlChange?.(value);
    };

    const handleApiKeyChange = (value: string) => {
        setApiKey(value);
        onApiKeyChange?.(value);
    };

    const handleSaveProvider = async () => {
        setIsSaving(true);
        try {
            const settings = new ProviderSettings({
                api_url: apiUrl,
                api_key: apiKey,
                is_enabled: currentProviderSettings?.is_enabled ?? true,
            });

            await SetProviderSetting(providerName, settings);
        } catch (error) {
            console.error('Failed to save provider settings:', error);
        } finally {
            setIsSaving(false);
        }
    };

    const handleRefresh = () => {
        ListLLMModels();
        GetAllProviderSettings();
    };

    const displayTitle = displayName || providerName.charAt(0).toUpperCase() + providerName.slice(1);

    return (
        <div className={`container mx-auto ${compact ? 'px-3 py-3' : 'px-6 py-8'}`}>
            {/* Header */}
            <div className={`flex items-center justify-between ${compact ? 'mb-3' : 'mb-6'}`}>
                <div>
                    <h1 className="text-lg sm:text-xl lg:text-2xl font-bold text-foreground capitalize">
                        {displayTitle} Models
                    </h1>
                    <p className="text-sm text-muted-foreground mt-1">
                        {filteredModels.length} models available
                    </p>
                </div>

                <div className="flex gap-2">
                    <button
                        onClick={() => setIsAddModalOpen(true)}
                        className="flex items-center space-x-2 px-4 py-2 bg-secondary hover:bg-secondary/80 text-secondary-foreground rounded-md transition-colors"
                    >
                        <svg
                            className="w-4 h-4"
                            fill="none"
                            stroke="currentColor"
                            viewBox="0 0 24 24"
                        >
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                        </svg>
                        <span>Add Model</span>
                    </button>
                    <button
                        onClick={handleRefresh}
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
            </div>

            <AddModelDialog
                providerName={providerName}
                isOpen={isAddModalOpen}
                onOpenChange={setIsAddModalOpen}
            />

            {/* Provider Settings - Hide for local provider if needed, but logic in original file showed it only if not local */}
            {providerName !== 'local' && (
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
                                onChange={(e) => handleApiUrlChange(e.target.value)}
                                className="w-full px-3 py-2 bg-background border border-input rounded-md text-foreground focus:outline-none focus:ring-2 focus:ring-primary"
                                placeholder={defaultApiUrl}
                            />
                            <p className="text-[10px] text-muted-foreground mt-1">
                                Note: Use OpenAI chat completions API compatible endpoint only
                            </p>
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-foreground mb-2">
                                API Key
                            </label>
                            <input
                                type="password"
                                value={apiKey}
                                onChange={(e) => handleApiKeyChange(e.target.value)}
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
                        <ModelCard key={model.id} model={model} isLocal={providerName === 'local'} />
                    ))}
                </div>
            )}
        </div>
    );
};
