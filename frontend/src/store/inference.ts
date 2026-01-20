import { atom, onMount, computed } from "nanostores";
import {
    DownloadModelRequest, InferenceServiceClient, GetLLMModelsRequest, Model, CancelDownloadRequest, DeleteModelRequest
} from "../../proto/inferenceservice"
import { createAuthenticatedClientOptions } from "../lib/auth";
import { getUIConfig } from "../lib/config";
import { persistentAtom } from '@nanostores/persistent';


let _inferenceClient: InferenceServiceClient | undefined;

function getClient(): InferenceServiceClient {
    if (!_inferenceClient) {
        const config = getUIConfig();
        if (!config) {
            throw new Error("UI config not loaded, cannot initialize chat client.");
        }
        _inferenceClient = new InferenceServiceClient(
            config.API_URL,
            {},
            createAuthenticatedClientOptions()
        );
    }
    return _inferenceClient;
}

export const $llmModels = atom<Model[]>([]);
export const $isLoadingModels = atom<boolean>(false);

export const $modelsByProvider = computed($llmModels, (models) => {
    const groups: Record<string, Model[]> = {};
    models.forEach((model) => {
        const provider = model.provider || "other";
        if (!groups[provider]) {
            groups[provider] = [];
        }
        groups[provider].push(model);
    });
    return groups;
});

export const downloadModel = async (modelId: string) => {
    try {
        const req = new DownloadModelRequest({
            model_id: modelId
        });
        const res = await getClient().DownloadModel(req, {});

        // Wait for model list to refresh completely
        await new Promise<void>((resolve, reject) => {
            const req = new GetLLMModelsRequest({});
            const stream = getClient().GetLLMModels(req, {});

            stream.on('data', (data) => {
                $llmModels.set(data.models);
            });

            stream.on('end', () => {
                console.log('Models list loaded after download');
                resolve();
            });

            stream.on('error', (err) => {
                console.error('Error loading models after download:', err);
                reject(err);
            });
        });

        return res;
    } catch (error) {
        console.error('Download failed:', error);
        throw error;
    }
}

export const ListLLMModels = async () => {
    $isLoadingModels.set(true);

    try {
        const req = new GetLLMModelsRequest({});
        const res = getClient().GetLLMModels(req, {});

        res.on('data', (data) => {
            $llmModels.set(data.models);
        });

        res.on('end', () => {
            console.log('Models list loaded');
            $isLoadingModels.set(false);
        });

        res.on('error', (err) => {
            console.error('Error loading models:', err);
            $isLoadingModels.set(false);
        });

        return res;
    } catch (error) {
        $isLoadingModels.set(false);
        throw error;
    }
}

onMount($llmModels, () => {
    ListLLMModels();

    return () => {
        // Disabled mode
    };
});

export const cancelDownload = async (modelId: string) => {
    try {
        const req = new CancelDownloadRequest({
            model_id: modelId
        });
        const res = await getClient().CancelDownload(req, {});
        console.log('CancelDownloadResponse', res.message);
        await ListLLMModels();
        return res;
    } catch (error) {
        console.error('CancelDownload failed:', error);
        throw error;
    }
}

export const deleteModel = async (modelId: string) => {
    try {
        const req = new DeleteModelRequest({
            model_id: modelId
        });
        const res = await getClient().DeleteModel(req, {});
        console.log('DeleteModelResponse', res.message);
        await ListLLMModels();
        return res;
    } catch (error) {
        console.error('DeleteModel failed:', error);
        throw error;
    }
}

export const $pinnedModels = persistentAtom<string[]>('pinned_models', [], {
    encode: JSON.stringify,
    decode: JSON.parse,
});

export const togglePinnedModel = (modelId: string) => {
    const currentPinned = $pinnedModels.get();
    if (currentPinned.includes(modelId)) {
        $pinnedModels.set(currentPinned.filter(id => id !== modelId));
    } else {
        $pinnedModels.set([...currentPinned, modelId]);
    }
};