import { createAuthenticatedClientOptions } from "../lib/auth";
import { IceCandidateRequest, OfferRequest, RealtimeServiceClient } from "../../proto/realtimeservice";
const client = new RealtimeServiceClient(import.meta.env.VITE_API_URL, {}, createAuthenticatedClientOptions());

export const offerRequest = async (offer: string, model: string) => {
    const req = new OfferRequest({
        offer: offer,
        model: model,
    });
    const res = await client.Offer(req, {});
    console.log("offerRequest", res);
    return res;
}

export const iceCandidate = async (candidate: string) => {
    const req = new IceCandidateRequest({
        candidate: candidate,
    });
    await client.IceCandidate(req, {});
}