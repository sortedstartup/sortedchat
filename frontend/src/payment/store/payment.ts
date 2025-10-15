import { atom } from "nanostores";
import {
    CreateProductRequest, PaymentServiceClient, ListProductsRequest, Product, CreateStripeCheckoutSessionRequest, CreateRazorpayCheckoutSessionRequest
} from "../../../proto/paymentservice"
import { createAuthenticatedClientOptions } from "../../lib/auth";

const client = new PaymentServiceClient(import.meta.env.VITE_API_URL, {}, createAuthenticatedClientOptions());


export const createProduct = async (name: string, description: string, cost: string, currency: string) => {
    const req = new CreateProductRequest({
        name: name,
        description: description,
        price: cost,
        currency: currency,
    });
    const res = await client.CreateProduct(req, {});

    console.log(res);
    console.log(res.message);
    console.log(res.id);
    return res.id;
}

export const ProductList = atom<Product[]>([]);

export const listProducts = async () => {
    const req = new ListProductsRequest({});
    const res = await client.ListProducts(req, {});
    ProductList.set(res.products);
    res.products.map((product) => {
        console.log(product.description, product.name, product.price, product.currency, product.id);
    });
    return res.products;
}

export const createStripeCheckoutSession = async (productId: string) => {
    try {
        const req = new CreateStripeCheckoutSessionRequest({ product_id: productId });
        const res = await client.CreateStripeCheckoutSession(req, {});
        return res.session_url;
    } catch (err) {
        console.error("Failed to create checkout session:", err);
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