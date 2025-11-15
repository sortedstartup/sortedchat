import React, { useState } from "react";
import { createProduct } from "./store/payment";
import { Currency, PaymentType, Interval } from "../../proto/paymentservice";

const CreateProduct: React.FC = () => {
    const [name, setName] = useState("");
    const [description, setDescription] = useState("");
    const [cost, setCost] = useState("");
    const [currency, setCurrency] = useState<Currency>(Currency.USD);
    const [paymentType, setPaymentType] = useState<PaymentType>(PaymentType.ONE_TIME);
    const [intervalCount, setIntervalCount] = useState<number>(1);
    const [interval, setInterval] = useState<Interval>(Interval.MONTH);
    const [loading, setLoading] = useState(false);
    const [message, setMessage] = useState("");

    const onSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setMessage("");
        try {
            const id = await createProduct(
                name, 
                description, 
                cost, 
                currency, 
                paymentType, 
                paymentType === PaymentType.RECURRING ? intervalCount : undefined,
                paymentType === PaymentType.RECURRING ? interval : undefined
            );
            setMessage(`Created product with id: ${id}`);
            setName("");
            setDescription("");
            setCost("");
            setCurrency(Currency.USD);
            setPaymentType(PaymentType.ONE_TIME);
            setIntervalCount(1);
            setInterval(Interval.MONTH);
        } catch (err: any) {
            setMessage(err?.message || "Failed to create product");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="p-4 max-w-md mx-auto">
            <h1 className="text-xl font-semibold mb-4">Create Product</h1>
            <form onSubmit={onSubmit} className="space-y-4">
                <input
                    className="w-full border p-2 rounded"
                    placeholder="Name"
                    value={name}
                    onChange={(e) => setName(e.target.value)}
                    required
                />
                <textarea
                    className="w-full border p-2 rounded"
                    placeholder="Description"
                    value={description}
                    onChange={(e) => setDescription(e.target.value)}
                    required
                />
                <input
                    className="w-full border p-2 rounded"
                    placeholder="Cost"
                    value={cost}
                    onChange={(e) => setCost(e.target.value)}
                    required
                />
                <select
                    className="w-full border p-2 rounded"
                    value={currency}
                    onChange={(e) => setCurrency(parseInt(e.target.value) as Currency)}
                    required
                >
                    <option value={Currency.USD}>USD</option>
                    <option value={Currency.INR}>INR</option>
                </select>

                {/* Payment Type Radio Buttons */}
                <div className="space-y-2">
                    <label className="block text-sm font-medium">Payment Type</label>
                    <div className="flex space-x-4">
                        <label className="flex items-center">
                            <input
                                type="radio"
                                name="paymentType"
                                value={PaymentType.ONE_TIME}
                                checked={paymentType === PaymentType.ONE_TIME}
                                onChange={(e) => setPaymentType(parseInt(e.target.value) as PaymentType)}
                                className="mr-2"
                            />
                            One-time Payment
                        </label>
                        <label className="flex items-center">
                            <input
                                type="radio"
                                name="paymentType"
                                value={PaymentType.RECURRING}
                                checked={paymentType === PaymentType.RECURRING}
                                onChange={(e) => setPaymentType(parseInt(e.target.value) as PaymentType)}
                                className="mr-2"
                            />
                            Recurring Payment
                        </label>
                    </div>
                </div>

                {/* Recurring Payment Options */}
                {paymentType === PaymentType.RECURRING && (
                    <div className="space-y-3 p-3 border rounded bg-gray-50">
                        <h3 className="text-sm font-medium">Recurring Payment Settings</h3>
                        
                        <div>
                            <label className="block text-sm font-medium mb-1">Interval Count</label>
                            <input
                                type="number"
                                min="1"
                                className="w-full border p-2 rounded"
                                placeholder="e.g., 1 for every month, 3 for every 3 months"
                                value={intervalCount}
                                onChange={(e) => setIntervalCount(parseInt(e.target.value) || 1)}
                                required
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium mb-1">Billing Interval</label>
                            <select
                                className="w-full border p-2 rounded"
                                value={interval}
                                onChange={(e) => setInterval(parseInt(e.target.value) as Interval)}
                                required
                            >
                                <option value={Interval.WEEK}>Weekly</option>
                                <option value={Interval.MONTH}>Monthly</option>
                                <option value={Interval.QUARTER}>Quarterly</option>
                                <option value={Interval.YEAR}>Yearly</option>
                            </select>
                        </div>

                        <div className="text-xs text-gray-600">
                            {intervalCount > 1 ? (
                                <>Billing every {intervalCount} {
                                    interval === Interval.WEEK ? 'weeks' :
                                    interval === Interval.MONTH ? 'months' :
                                    interval === Interval.QUARTER ? 'quarters' :
                                    'years'
                                }</>
                            ) : (
                                <>Billing {
                                    interval === Interval.WEEK ? 'weekly' :
                                    interval === Interval.MONTH ? 'monthly' :
                                    interval === Interval.QUARTER ? 'quarterly' :
                                    'yearly'
                                }</>
                            )}
                        </div>
                    </div>
                )}

                <button
                    type="submit"
                    disabled={loading}
                    className="w-full bg-blue-600 text-white p-2 rounded disabled:opacity-50"
                >
                    {loading ? "Creating..." : "Create"}
                </button>
            </form>
            {message && (
                <div className="mt-3 text-sm">{message}</div>
            )}
        </div>
    );
};

export default CreateProduct;


