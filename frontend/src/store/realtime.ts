import { createAuthenticatedClientOptions } from "../lib/auth";
import { IceCandidateRequest, OfferRequest, RealtimeServiceClient } from "../../proto/realtimeservice";
import { getUIConfig } from "../lib/config";


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


export const offerRequest = async (offer: string, model: string) => {
    const req = new OfferRequest({
        offer: offer,
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