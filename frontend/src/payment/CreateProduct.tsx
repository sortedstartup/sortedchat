import React, { useState } from "react";
import { createProduct } from "../store/payment";

const CreateProduct: React.FC = () => {
    const [name, setName] = useState("");
    const [description, setDescription] = useState("");
    const [cost, setCost] = useState("");
    const [currency, setCurrency] = useState("");
    const [loading, setLoading] = useState(false);
    const [message, setMessage] = useState("");

    const onSubmit = async (e: React.FormEvent) => {
        e.preventDefault();
        setLoading(true);
        setMessage("");
        try {
            const id = await createProduct(name, description, cost, currency);
            setMessage(`Created product with id: ${id}`);
            setName("");
            setDescription("");
            setCost("");
            setCurrency("");
        } catch (err: any) {
            setMessage(err?.message || "Failed to create product");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="p-4 max-w-md mx-auto">
            <h1 className="text-xl font-semibold mb-4">Create Product</h1>
            <form onSubmit={onSubmit} className="space-y-3">
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
                <input
                    className="w-full border p-2 rounded"
                    placeholder="Currency (e.g. USD)"
                    value={currency}
                    onChange={(e) => setCurrency(e.target.value)}
                    required
                />
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


