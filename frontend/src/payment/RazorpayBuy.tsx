import React, { useState } from "react";
import { createRazorpayCheckoutSession } from "./store/payment";

declare global {
    interface Window {
        Razorpay: any;
    }
}

interface RazorpayBuyProps {
    productId: string;
    className?: string;
    children?: React.ReactNode;
}

const RazorpayBuy: React.FC<RazorpayBuyProps> = ({ productId, className = "", children }) => {
    const [isLoading, setIsLoading] = useState(false);
    const [error, setError] = useState("");

    const handleBuyNow = async () => {
        try {
            setIsLoading(true);
            setError("");
            const { orderId, amount, currency } = await createRazorpayCheckoutSession(productId);

            // Initialize Razorpay checkout
            const options = {
                key: import.meta.env.VITE_RAZORPAY_KEY_ID, // Your Razorpay key ID mandatory
                amount: amount, // mandatory - amount in smallest currency unit
                order_id: orderId, // mandatory
                name: "SortedChat", // mandatory
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
                className={`bg-blue-500 text-white p-2 rounded hover:bg-blue-600 disabled:opacity-50 disabled:cursor-not-allowed ${className}`}
            >
                {isLoading ? "Processing..." : (children || "Buy with Razorpay")}
            </button>
            {error && (
                <div className="text-red-600 text-sm mt-1">{error}</div>
            )}
        </div>
    );
};

export default RazorpayBuy;
