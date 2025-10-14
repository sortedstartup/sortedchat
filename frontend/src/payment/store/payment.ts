import { atom } from "nanostores";
import {
    CreateProductRequest, PaymentServiceClient, ListProductsRequest, Product, CreateCheckoutSessionRequest
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

export const createCheckoutSession = async (productId: string) => {
    const req = new CreateCheckoutSessionRequest({
        product_id: productId,
    });
    const res = await client.CreateCheckoutSession(req, {});
    console.log("Checkout session created:", res.session_id);
    return res.session_id;
}