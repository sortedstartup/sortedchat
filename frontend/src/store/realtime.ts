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
        if (res.offer || res.offer !== "") {
            console.log(res.offer);
            $isConnected.set(true);
            toast.success("Offer request sent successfully");
            return res;
        } else {
            $isConnected.set(false);
            toast.error("Failed to send offer request");
            throw new Error("Failed to send offer request");
        }
    } catch (error) {
        $isConnected.set(false);
        toast.error("Failed to send offer request");
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