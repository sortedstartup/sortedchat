import React, { useState } from "react";
import { createRazorpaySubscriptionCheckoutSession } from "./store/payment";

// Extend window interface for Razorpay
declare global {
    interface Window {
        Razorpay: any;
    }
}

interface SubscriptionRazorpayBuyProps {
    productId: string;
    className?: string;
    children?: React.ReactNode;
}

const SubscriptionRazorpayBuy: React.FC<SubscriptionRazorpayBuyProps> = ({ productId, className = "", children }) => {
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState("");

    const handleSubscribeNow = async () => {
        try {
            setIsLoading(true);
            setError("");
            const { subscriptionId, currency, amount } = await createRazorpaySubscriptionCheckoutSession(productId);

            // Initialize Razorpay checkout for subscription
            const options = {
                key: import.meta.env.VITE_RAZORPAY_KEY_ID, // Your Razorpay key ID mandatory
                subscription_id: subscriptionId, // For subscriptions, use subscription_id instead of order_id
                name: "SortedChat", // mandatory
                amount: amount,
                currency: currency, // mandatory
                handler: function () {
                    window.location.href = '/success'
                }
            };

            if (!window.Razorpay) {
                throw new Error('Razorpay SDK failed to load. Please check your internet connection and try again.');
            }

            const rzp = new window.Razorpay(options);
            rzp.open();
        } catch (err: any) {
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
                className={`bg-blue-500 text-white p-2 rounded hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed ${className}`}
            >
                {isLoading ? "Processing..." : (children || "Subscribe with Razorpay")}
            </button>
            {error && (
                <div className="text-red-600 text-sm mt-1">{error}</div>
            )}
        </div>
    );
};

export default SubscriptionRazorpayBuy;
