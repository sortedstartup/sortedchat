import { atom } from "nanostores";
import {
    DownloadModelRequest, InferenceServiceClient, ListLLMModelsRequest, Model
} from "../../proto/inferenceservice"

const client = new InferenceServiceClient(import.meta.env.VITE_API_URL);

export const $llmModels = atom<Model[]>([]);
export const $isLoadingModels = atom<boolean>(false);
export const $downloadingModels = atom<Set<string>>(new Set());

export const downloadModel = async (modelName: string) => { 
    const downloadingSet = new Set($downloadingModels.get());
    downloadingSet.add(modelName);
    $downloadingModels.set(downloadingSet);
    
    try {
        const req = new DownloadModelRequest({
            model_name: modelName
        });
        const res = await client.DownloadModel(req, {});
        console.log(res.message);
        
        await ListLLMModels();
        
        return res;
    } catch (error) {
        console.error('Download failed:', error);
        throw error;
    } finally {
        const updatedDownloading = new Set($downloadingModels.get());
        updatedDownloading.delete(modelName);
        $downloadingModels.set(updatedDownloading);
    }
}

export const ListLLMModels = async () => {
    $isLoadingModels.set(true);
    
    try {
        const req = new ListLLMModelsRequest({});
        const res = client.ListLLMModels(req, {});
        
        res.on('data', (data) => {
            console.log(data.models);
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

ListLLMModels();