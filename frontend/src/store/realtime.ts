import { createAuthenticatedClientOptions } from "../lib/auth";
import { IceCandidateRequest, OfferRequest, RealtimeServiceClient } from "../../proto/realtimeservice";
import { toast } from "sonner";
import { atom } from "nanostores";

const client = new RealtimeServiceClient(import.meta.env.VITE_API_URL, {}, createAuthenticatedClientOptions());

export const $isConnected = atom<boolean>(false);

export const offerRequest = async (offer: string, model: string) => {
    const req = new OfferRequest({
        offer: offer,
        model: model,
    });
    try {
        const res = await client.Offer(req, {});
        console.log("offerRequest response", res);
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
        const res = await client.IceCandidate(req, {});
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