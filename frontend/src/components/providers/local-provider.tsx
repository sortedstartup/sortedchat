import { ProviderView } from './provider-view';

export const LocalProvider = ({ compact = false }: { compact?: boolean }) => {
    return (
        <ProviderView
            providerName="local"
            displayName="Local (llama.cpp)"
            compact={compact}
        />
    );
};
