import React, { useState } from "react";
import { createStripeCheckoutSession } from "./store/payment";

interface BuyProps {
    productId: string;
    className?: string;
    children?: React.ReactNode;
}

const Buy: React.FC<BuyProps> = ({ productId, className = "", children }) => {
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState("");

    const handleBuyNow = async () => {
        try {
            setIsLoading(true);
            setError("");
            const sessionUrl = await createStripeCheckoutSession(productId);
            // Redirect to Stripe checkout
            window.location.href = sessionUrl;
        } catch (err: any) {
            console.error("Failed to create checkout session:", err);
            setError(err?.message || "Failed to create checkout session");
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div>
            <button
                onClick={handleBuyNow}
                disabled={isLoading}
                className={`bg-blue-600 text-white p-2 rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed ${className}`}
            >
                {isLoading ? "Processing..." : (children || "Buy Now")}
            </button>
            {error && (
                <div className="text-red-600 text-sm mt-1">{error}</div>
            )}
        </div>
    );
};

export default Buy;
