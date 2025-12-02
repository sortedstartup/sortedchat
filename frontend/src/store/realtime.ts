import { createAuthenticatedClientOptions } from "../lib/auth";
import { IceCandidateRequest, OfferRequest, RealtimeServiceClient } from "../../proto/realtimeservice";
import { getUIConfig } from "../lib/config";
import { ListModelsRequest, ModelListInfo } from "../../proto/chatservice";
import { getChatClient } from "./chat";
import { atom } from "nanostores";

import { atom } from "nanostores";
import { toast } from "sonner";


let _realtimeClient: RealtimeServiceClient | undefined;

export const $isConnected = atom<boolean>(false);


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
        const offer = res.offer;

        if (offer) {
            console.log(offer);
            $isConnected.set(true);
            toast.success("Offer request sent successfully");
            return offer; // still returns SDP string for existing callers
        }

        // No offer in a successful response: treat as an error
        throw new Error("Failed to send offer request");
    } catch (error) {
        $isConnected.set(false);
        toast.error(
            error instanceof Error && error.message
                ? error.message
                : "Failed to send offer request"
        );
        throw error;
    }
}

export const iceCandidate = async (candidate: string) => {
    const req = new IceCandidateRequest({
        candidate: candidate,
    });
    try {
        const res = await getClient().IceCandidate(req, {});
        if (res.message) {
            // toast.success("ICE candidate sent successfully");
            return res;
        } else {
            // toast.error("Failed to send ICE candidate");
            throw new Error("Failed to send ICE candidate");
        }
    } catch (error) {
        // toast.error("Failed to send ICE candidate");
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