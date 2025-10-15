import { atom } from "nanostores";
import {
    CreateProductRequest, PaymentServiceClient, ListProductsRequest, Product, CreateCheckoutSessionRequest, Currency
} from "../../../proto/paymentservice"
import { createAuthenticatedClientOptions } from "../../lib/auth";
import { toast } from "sonner";

const client = new PaymentServiceClient(import.meta.env.VITE_API_URL, {}, createAuthenticatedClientOptions());


export const createProduct = async (name: string, description: string, priceInUSD: string, currency: Currency) => {
    try {
        // Convert USD to cents
        const amountInCents = Math.round(parseFloat(priceInUSD) * 100);

        const req = new CreateProductRequest({
            name: name,
            description: description,
            amount_in_cents: amountInCents,
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
        throw err;
    }
}

export const createCheckoutSession = async (productId: string) => {
    try {
        const req = new CreateCheckoutSessionRequest({ product_id: productId });
        const res = await client.CreateCheckoutSession(req, {});
        toast.success("Checkout session created successfully");
        return res.session_url;
    } catch (err) {
        toast.error("Failed to create checkout session");
        throw err;
    }
}