import { ProviderView } from './provider-view';

export const ClaudeProvider = () => {
    return (
        <ProviderView
            providerName="claude"
            displayName="Claude"
            defaultApiUrl="https://api.anthropic.com/v1/messages"
        />
    );
};
