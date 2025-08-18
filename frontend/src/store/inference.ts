import {
    DownloadModelRequest, InferenceServiceClient
} from "../../proto/inferenceservice"


const client = new InferenceServiceClient(import.meta.env.VITE_API_URL);

export const downloadModel = async (modelName: string) => { 
    const req = new DownloadModelRequest({
        model_name: modelName
    });
    const res = await client.DownloadModel(req, {});
    console.log(res.message);
    return res;
}