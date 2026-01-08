import { ProviderView } from './provider-view';

interface OpenaiProviderProps {
    onApiKeyChange?: (key: string) => void;
    onApiUrlChange?: (url: string) => void;
}

export const OpenaiProvider = ({ onApiKeyChange, onApiUrlChange }: OpenaiProviderProps) => {
    return (
        <ProviderView
            providerName="openai"
            displayName="OpenAI"
            defaultApiUrl="https://api.openai.com/v1/chat/completions"
            onApiKeyChange={onApiKeyChange}
            onApiUrlChange={onApiUrlChange}
        />
    );
};
