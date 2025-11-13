import { registerPlugin } from "@capacitor/core";


export enum ProductType {
    CONSUMABLE = "consumable",           // Can be purchased multiple times
    NON_CONSUMABLE = "non_consumable",   // One-time permanent purchase
    SUBSCRIPTION = "subscription",        // Auto-renewable subscription
}

export enum PurchaseState {
    PENDING = "pending",       // Payment not completed yet
    PURCHASED = "purchased",   // Successfully purchased
    FAILED = "failed",        // Purchase failed
}

export enum ResponseCode {
    SUCCESS = 0,
    USER_CANCELLED = 1,
    PAYMENT_INVALID = 2,
    NOT_AVAILABLE = 3,
    ERROR = 4,
}

export interface Product {
    productId: string;
    type: ProductType;
    title: string;
    description: string;
    priceAmount: number;        
    currency: string;           
}

export interface Purchase {
    transactionId: string;      // Unique transaction identifier
    productId: string;
    purchaseDate: number;       // Unix timestamp
    state: PurchaseState;
}

export interface PurchaseResult {
    code: ResponseCode;
    message?: string;
    purchase?: Purchase;
}

export interface IAPPlugin {
    
    //andriod: BillingClient.startConnection()
    //ios:  (no explicit connection required)
    initialize(): Promise<{ success: boolean }>;

    //andriod: queryProductDetailsAsync()
    //ios: Product.products(for:)
    getProducts(options: {
        productIds: string[];
    }): Promise<{ products: Product[] }>;

    //andriod: launchBillingFlow()
    //ios: product.purchase()
    purchaseProduct(options: {
        productId: string;
        accountId?: string;
        customProductId?: string;
    }): Promise<PurchaseResult>;

    //andriod: queryPurchasesAsync()
    //ios: Transaction.currentEntitlements()
    getPurchases(options?: {
        type?: ProductType;
    }): Promise<{ purchases: Purchase[] }>;
}



const IAP = registerPlugin<IAPPlugin>("PaymentPlugin", {
    web: () => import("./PaymentPluginWeb").then(m => new m.PaymentPluginWeb()),
});

export default IAP;
