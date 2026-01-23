import { ProviderView } from './provider-view';

interface ClaudeProviderProps {
    onApiKeyChange?: (key: string) => void;
    onApiUrlChange?: (url: string) => void;
    onSaveSuccess?: () => void;
}

export const ClaudeProvider = ({ onApiKeyChange, onApiUrlChange, onSaveSuccess }: ClaudeProviderProps) => {
    return (
        <ProviderView
            providerName="claude"
            displayName="Claude"
            defaultApiUrl="https://api.anthropic.com/v1/messages"
            onApiKeyChange={onApiKeyChange}
            onApiUrlChange={onApiUrlChange}
            onSaveSuccess={onSaveSuccess}
        />
    );
};
