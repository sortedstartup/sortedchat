import { createAuthenticatedClientOptions } from "../lib/auth";
import { IceCandidateRequest, OfferRequest, RealtimeServiceClient } from "../../proto/realtimeservice";
import { getUIConfig } from "../lib/config";
import { ListModelsRequest, ModelListInfo } from "../../proto/chatservice";
import { getChatClient } from "./chat";
import { atom } from "nanostores";


let _realtimeClient: RealtimeServiceClient | undefined;

function getClient(): RealtimeServiceClient {
  if (!_realtimeClient) {
    const config = getUIConfig();
    if (!config) {
      throw new Error("UI config not loaded, cannot initialize chat client.");
    }
    _realtimeClient = new RealtimeServiceClient(
      config.API_URL,
      {},
      createAuthenticatedClientOptions()
    );
  }
  return _realtimeClient;
}

export const $realtimeModelList = atom<ModelListInfo[]>([]);

export const offerRequest = async (offer: string, provider: string, model: string) => {
  console.log("offerRequest", offer, provider, model);
    const req = new OfferRequest({
        offer: offer,
        provider: provider,
        model: model,
    });
    try {
        const res = await getClient().Offer(req, {});
        console.log("offerRequest", res);
        return res;
    } catch (error) {
        console.error("Failed to send offer request:", error);
        throw error;
    }
}

export const iceCandidate = async (candidate: string) => {
    const req = new IceCandidateRequest({
        candidate: candidate,
    });
    try {
        const res = await getClient().IceCandidate(req, {});
        console.log("iceCandidate", res);
        return res;
    } catch (error) {
        console.error("Failed to send ICE candidate:", error);
        throw error;
    }
}

export const listModels = async () => {
    const req = new ListModelsRequest({});
    try {
        const res = await getChatClient().ListModel(req, {});
        console.log("listModels", res);
        
        // Filter models where realtime is true and set realtimeModelList
        if (res.models) {
            const realtimeModels = res.models.filter(
                (model) => model.capabilities?.realtime === true
            );
            $realtimeModelList.set(realtimeModels);
        }
        
        return res;
    } catch (error) {
        console.error("Failed to list models:", error);
        throw error;
    }
};