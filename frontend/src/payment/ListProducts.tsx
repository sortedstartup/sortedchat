import React, { useEffect, useState } from "react";
import { listProducts, ProductList, createCheckoutSession } from "../store/payment";
import { useStore } from "@nanostores/react";

const ListProducts: React.FC = () => {
    const products = useStore(ProductList);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");
    const [buyingProductId, setBuyingProductId] = useState<string | null>(null);

    useEffect(() => {
        const fetchProducts = async () => {
            try {
                setLoading(true);
                await listProducts();
            } catch (err: any) {
                setError(err?.message || "Failed to fetch products");
            } finally {
                setLoading(false);
            }
        };

        fetchProducts();
    }, []);

    const handleBuyNow = async (productId: string) => {
        try {
            setBuyingProductId(productId);
            const sessionId = await createCheckoutSession(productId);
            // You can redirect to Stripe checkout or handle the session ID as needed
            console.log("Redirecting to checkout with session:", sessionId); //change variable name sessionId to session URL
           
            window.location.href = sessionId;
        } catch (err: any) {
            console.error("Failed to create checkout session:", err);
            setError(err?.message || "Failed to create checkout session");
        }
    };

    if (loading) {
        return (
            <div className="p-4 max-w-4xl mx-auto">
                <h1 className="text-xl font-semibold mb-4">Products</h1>
                <div className="text-center">Loading...</div>
            </div>
        );
    }

    if (error) {
        return (
            <div className="p-4 max-w-4xl mx-auto">
                <h1 className="text-xl font-semibold mb-4">Products</h1>
                <div className="text-red-600">Error: {error}</div>
            </div>
        );
    }

    return (
        <div className="p-4 max-w-4xl mx-auto">
            <h1 className="text-xl font-semibold mb-4">Products</h1>
            {products.length === 0 ? (
                <div className="text-gray-500">No products found</div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
                    {products.map((product) => (
                        <div key={product.id} className="border rounded-lg p-4 shadow-sm">
                            <h3 className="font-semibold text-lg mb-2">{product.name}</h3>
                            <p className="text-gray-600 mb-2">{product.description}</p>
                            <div className="flex justify-between items-center mb-3">
                                <span className="text-lg font-bold">
                                    {product.price} {product.currency}
                                </span>
                                <span className="text-sm text-gray-500">ID: {product.id}</span>
                            </div>
                            <button
                                onClick={() => handleBuyNow(product.id)}
                                disabled={buyingProductId === product.id}
                                className="w-full bg-blue-600 text-white p-2 rounded hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                {buyingProductId === product.id ? "Processing..." : "Buy Now"}
                            </button>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};

export default ListProducts;
