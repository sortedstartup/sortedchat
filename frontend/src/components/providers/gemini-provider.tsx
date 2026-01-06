import { ProviderView } from './provider-view';

export const GeminiProvider = () => {
    return (
        <ProviderView
            providerName="gemini"
            displayName="Gemini"
            defaultApiUrl="https://generativelanguage.googleapis.com/v1beta/openai/chat/completions"
        />
    );
};
