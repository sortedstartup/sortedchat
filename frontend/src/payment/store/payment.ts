import { atom } from "nanostores";
import {
    CreateProductRequest, PaymentServiceClient, ListProductsRequest, Product, CreateStripeCheckoutSessionRequest, CreateRazorpayCheckoutSessionRequest, Currency
} from "../../../proto/paymentservice"
import { createAuthenticatedClientOptions } from "../../lib/auth";
import { toast } from "sonner";

const client = new PaymentServiceClient(import.meta.env.VITE_API_URL, {}, createAuthenticatedClientOptions());


export const createProduct = async (name: string, description: string, price: string, currency: Currency) => {
    try {
        const parsed = Number(price);
        if(!Number.isFinite(parsed) || parsed < 0) {
            toast.error("Price must be a non-negative number");
            throw new Error("invalid price");
        }
        const amountInMinorUnits = Math.round(parsed * 100);

        const req = new CreateProductRequest({
            name: name,
            description: description,
            amount_in_cents: amountInMinorUnits,
            currency: currency,
        });
        const res = await client.CreateProduct(req, {});
        toast.success("Product created successfully");
        return res.id;
    } catch (err) {
        toast.error("Failed to create product");
        throw err;
    }
}

export const ProductList = atom<Product[]>([]);

export const listProducts = async () => {
    try {
        const req = new ListProductsRequest({});
        const res = await client.ListProducts(req, {});
        ProductList.set(res.products);
        return res.products;
    } catch (err) {
        toast.error("Failed to list products");
        throw err;
    }
}

export const createStripeCheckoutSession = async (productId: string) => {
    try {
        const req = new CreateStripeCheckoutSessionRequest({ product_id: productId });
        const res = await client.CreateStripeCheckoutSession(req, {});
        toast.success("Checkout session created successfully");
        return res.session_url;
    } catch (err) {
        toast.error("Failed to create checkout session");
        throw err;
    }
}

export const createRazorpayCheckoutSession = async (productId: string) => {
    try {
    const req = new CreateRazorpayCheckoutSessionRequest({ product_id: productId });
        const res = await client.CreateRazorpayCheckoutSession(req, {});
        return {
            orderId: res.order_id,
            amount: parseInt(res.amount),
            currency: res.currency
        };
    } catch (err) {
        toast.error("Failed to create checkout session");
        throw err;
    }
}

export const createRazorpayCheckoutSession = async (productId: string) => {
    try {
    const req = new CreateRazorpayCheckoutSessionRequest({ product_id: productId });
        const res = await client.CreateRazorpayCheckoutSession(req, {});
        return {
            orderId: res.order_id,
            amount: parseInt(res.amount),
            currency: res.currency
        };
    } catch (err) {
        console.error("Failed to create checkout session:", err);
        throw err;
    }
}