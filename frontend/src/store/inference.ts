import { atom, onMount, computed } from "nanostores";
import {
    DownloadModelRequest, InferenceServiceClient, GetLLMModelsRequest, Model, CancelDownloadRequest, DeleteModelRequest
} from "../../proto/inferenceservice"
import { createAuthenticatedClientOptions } from "../lib/auth";
import { getUIConfig } from "../lib/config";


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

export const downloadModel = async (modelName: string) => {

    try {
        const req = new DownloadModelRequest({
            model_name: modelName
        });
        const res = await getClient().DownloadModel(req, {});

        await ListLLMModels();

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

export const cancelDownload = async (modelName: string) => {
    try {
        const req = new CancelDownloadRequest({
            model_name: modelName
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

export const deleteModel = async (modelName: string) => {
    try {
        const req = new DeleteModelRequest({
            model_name: modelName
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