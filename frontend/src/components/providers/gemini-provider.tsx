import { ProviderView } from './provider-view';

interface GeminiProviderProps {
    onApiKeyChange?: (key: string) => void;
    onApiUrlChange?: (url: string) => void;
}

export const GeminiProvider = ({ onApiKeyChange, onApiUrlChange }: GeminiProviderProps) => {
    return (
        <ProviderView
            providerName="gemini"
            displayName="Gemini"
            defaultApiUrl="https://generativelanguage.googleapis.com/v1beta/openai/chat/completions"
            onApiKeyChange={onApiKeyChange}
            onApiUrlChange={onApiUrlChange}
        />
    );
};
