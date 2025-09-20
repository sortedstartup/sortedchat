import { createAuthenticatedClientOptions } from "../lib/auth";
import { IceCandidateRequest, OfferRequest, RealtimeServiceClient } from "../../proto/realtimeservice";
const client = new RealtimeServiceClient(import.meta.env.VITE_API_URL, {}, createAuthenticatedClientOptions());

export const offerRequest = async (offer: string, model: string) => {
    const req = new OfferRequest({
        offer: offer,
        model: model,
    });
    try {
        const res = await client.Offer(req, {});
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
        const res = await client.IceCandidate(req, {});
        console.log("iceCandidate", res);
        return res;
    } catch (error) {
        console.error("Failed to send ICE candidate:", error);
        throw error;
    }
}