import { ProviderView } from './provider-view';

export const OpenaiProvider = () => {
    return (
        <ProviderView
            providerName="openai"
            displayName="OpenAI"
            defaultApiUrl="https://api.openai.com/v1/chat/completions"
        />
    );
};
