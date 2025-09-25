import { atom, onMount } from "nanostores";
import {
    DownloadModelRequest, InferenceServiceClient, GetLLMModelsRequest, Model, CancelDownloadRequest, DeleteModelRequest
} from "../../proto/inferenceservice"
import { createAuthenticatedClientOptions } from "../lib/auth";
import { getUIConfig } from "../lib/config";

const client = new InferenceServiceClient(getUIConfig()?.API_URL || "http://localhost:8080", {}, createAuthenticatedClientOptions());

export const $llmModels = atom<Model[]>([]);
export const $isLoadingModels = atom<boolean>(false);

export const downloadModel = async (modelName: string) => { 
    
    try {
        const req = new DownloadModelRequest({
            model_name: modelName
        });
        const res = await client.DownloadModel(req, {});
        
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
        const res = client.GetLLMModels(req, {});
        
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
        const res = await client.CancelDownload(req, {});
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
        const res = await client.DeleteModel(req, {});
        console.log('DeleteModelResponse', res.message);
        await ListLLMModels();
        return res;
    } catch (error) {
        console.error('DeleteModel failed:', error);
        throw error;
    }
}