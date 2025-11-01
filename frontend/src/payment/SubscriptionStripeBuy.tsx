import React, { useState } from "react";
import { createStripeSubscriptionCheckoutSession } from "./store/payment";

interface SubscriptionStripeBuyProps {
    productId: string;
    className?: string;
    children?: React.ReactNode;
}

const SubscriptionStripeBuy: React.FC<SubscriptionStripeBuyProps> = ({ productId, className = "", children }) => {
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState("");

    const handleSubscribeNow = async () => {
        try {
            setIsLoading(true);
            setError("");
            const sessionUrl = await createStripeSubscriptionCheckoutSession(productId);
            // Redirect to Stripe checkout
            window.location.href = sessionUrl;
        } catch (err: any) {
            console.error("Failed to create subscription checkout session:", err);
            setError(err?.message || "Failed to create subscription checkout session");
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div>
            <button
                onClick={handleSubscribeNow}
                disabled={isLoading}
                className={`bg-blue-600 text-white p-2 rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed ${className}`}
            >
                {isLoading ? "Processing..." : (children || "Subscribe with Stripe")}
            </button>
            {error && (
                <div className="text-red-600 text-sm mt-1">{error}</div>
            )}
        </div>
    );
};

export default SubscriptionStripeBuy;
