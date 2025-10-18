import { atom } from "nanostores";
import {
    CreateProductRequest, PaymentServiceClient, ListProductsRequest, Product, CreateStripeCheckoutSessionRequest, CreateRazorpayCheckoutSessionRequest, Currency, PaymentType, Interval, CreateStripeSubscriptionCheckoutSessionRequest, CreateRazorpaySubscriptionCheckoutSessionRequest, CheckUserProductAccessRequest
} from "../../../proto/paymentservice"
import { createAuthenticatedClientOptions } from "../../lib/auth";
import { toast } from "sonner";

const client = new PaymentServiceClient(import.meta.env.VITE_API_URL, {}, createAuthenticatedClientOptions());


export const createProduct = async (
    name: string, 
    description: string, 
    price: string, 
    currency: Currency, 
    paymentType: PaymentType, 
    intervalCount?: number, 
    interval?: Interval
) => {
    try {
        const parsed = Number(price);
        if(!Number.isFinite(parsed) || parsed < 0) {
            toast.error("Price must be a non-negative number");
            throw new Error("invalid price");
        }
        const amountInMinorUnits = Math.round(parsed * 100);

        // Validate recurring payment parameters
        if (paymentType === PaymentType.RECURRING) {
            if (!intervalCount || intervalCount <= 0) {
                toast.error("Interval count must be greater than 0 for recurring payments");
                throw new Error("invalid interval count");
            }
            if (interval === undefined) {
                toast.error("Interval must be specified for recurring payments");
                throw new Error("invalid interval");
            }
        }

        const req = new CreateProductRequest({
            name: name,
            description: description,
            amount_in_smallest_unit: amountInMinorUnits,
            currency: currency,
            payment_type: paymentType,
            interval_count: intervalCount || 1,
            interval: interval || Interval.MONTH,
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
            amount: res.amount,
            currency: res.currency
        };
    } catch (err) {
        toast.error("Failed to create checkout session");
        throw err;
    }
}

// New subscription methods
export const createStripeSubscriptionCheckoutSession = async (productId: string) => {
    try {
        const req = new CreateStripeSubscriptionCheckoutSessionRequest({ 
            product_id: productId
        });
        const res = await client.CreateStripeSubscriptionCheckoutSession(req, {});
        toast.success("Subscription checkout session created successfully");
        return res.session_url;
    } catch (err) {
        toast.error("Failed to create subscription checkout session");
        throw err;
    }
}

export const createRazorpaySubscriptionCheckoutSession = async (productId: string) => {
    try {
        const req = new CreateRazorpaySubscriptionCheckoutSessionRequest({ 
            product_id: productId
        });
        const res = await client.CreateRazorpaySubscriptionCheckoutSession(req, {});
        return {
            subscriptionId: res.subscription_id,
            amount: res.amount,
            currency: res.currency
        };
    } catch (err) {
        toast.error("Failed to create subscription checkout session");
        throw err;
    }
}

export const checkUserProductAccess = async (productId: string) => {
    try {
        const req = new CheckUserProductAccessRequest({ 
            product_id: productId
        });
        const res = await client.CheckUserProductAccess(req, {});
        return {
            hasAccess: res.has_access,
            subscriptionId: res.subscription_id,
            status: res.status,
            currentPeriodEnd: res.current_period_end
        };
    } catch (err) {
        // Don't show toast for access check failures as they're expected
        throw err;
    }
}