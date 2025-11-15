import React, { useEffect, useState } from "react";
import { listProducts, ProductList } from "./store/payment";
import { useStore } from "@nanostores/react";
import StripeBuy from "./StripeBuy";
import RazorpayBuy from "./RazorpayBuy";
import SubscriptionStripeBuy from "./SubscriptionStripeBuy";
import SubscriptionRazorpayBuy from "./SubscriptionRazorpayBuy";
import { Currency } from "../../proto/paymentservice";

const ListProducts: React.FC = () => {
    const products = useStore(ProductList);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState("");

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
                            
                            {/* Payment Type Badge */}
                            <div className="mb-3">
                                {product.is_recurring ? (
                                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-purple-100 text-purple-800">
                                        🔄 Subscription
                                        {product.interval_count > 1
                                          ? ` (Every ${product.interval_count} ${product.interval_period}s)`
                                          : ` (${
                                              ((product.interval_period || '').toString().toLowerCase()) === 'week' ? 'weekly' :
                                              ((product.interval_period || '').toString().toLowerCase()) === 'month' ? 'monthly' :
                                              ((product.interval_period || '').toString().toLowerCase()) === 'quarter' ? 'quarterly' :
                                              'yearly'
                                            })`
                                        }
                                    </span>
                                ) : (
                                    <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
                                        💳 One-time
                                    </span>
                                )}
                            </div>

                            <div className="flex justify-between items-center mb-3">
                                <span className="text-lg font-bold">
                                    {new Intl.NumberFormat(undefined, {
                                        style: "currency",
                                        currency:
                                            product.currency === Currency.USD ? "USD" :
                                            product.currency === Currency.INR ? "INR" : "USD",
                                    }).format(product.amount_in_smallest_unit / 100)}
                                    {product.is_recurring && (
                                        <span className="text-sm text-gray-500 font-normal">
                                            /{product.interval_count > 1 ? `${product.interval_count} ${product.interval_period}s` : product.interval_period}
                                        </span>
                                    )}
                                </span>
                                <span className="text-sm text-gray-500">ID: {product.id}</span>
                            </div>

                            {/* Buy/Subscribe Buttons or Access Status */}
                            <div className="space-y-2">
                                {product.has_access ? (
                                    // User has access - show access status
                                    <div className="w-full p-3 bg-green-50 border border-green-200 rounded-lg text-center">
                                        <span className="text-green-700 font-medium">
                                            ✅ You have access to this product
                                        </span>
                                    </div>
                                ) : (
                                    // User doesn't have access - show payment buttons
                                    <>
                                        {product.is_recurring ? (
                                            // Subscription buttons
                                            <>
                                                {product.stripe_product_id && (
                                                    <SubscriptionStripeBuy productId={product.id} className="w-full">
                                                        Subscribe with Stripe
                                                    </SubscriptionStripeBuy>
                                                )}
                                                {product.razorpay_product_id && (
                                                    <SubscriptionRazorpayBuy productId={product.id} className="w-full">
                                                        Subscribe with Razorpay
                                                    </SubscriptionRazorpayBuy>
                                                )}
                                            </>
                                        ) : (
                                            // One-time payment buttons
                                            <>
                                                {product.stripe_product_id && (
                                                    <StripeBuy productId={product.id} className="w-full">
                                                        Buy with Stripe
                                                    </StripeBuy>
                                                )}
                                                {product.razorpay_product_id && (
                                                    <RazorpayBuy productId={product.id} className="w-full">
                                                        Buy with Razorpay
                                                    </RazorpayBuy>
                                                )}
                                            </>
                                        )}
                                    </>
                                )}
                            </div>
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
};

export default ListProducts;
