import { WebPlugin } from "@capacitor/core";
import { IAPPlugin, Product, Purchase, PurchaseResult, ResponseCode, ProductType, PurchaseState } from "./PaymentPlugin";

export class PaymentPluginWeb extends WebPlugin implements IAPPlugin {
    private products: Product[] = [
        { productId: "prod_1", type: ProductType.CONSUMABLE, title: "100 Coins", description: "Get 100 game coins", priceAmount: 1.99, currency: "USD" },
        { productId: "prod_2", type: ProductType.NON_CONSUMABLE, title: "Premium Upgrade", description: "Unlock premium features", priceAmount: 4.99, currency: "USD" },
    ];

    private purchases: Purchase[] = [];

    async initialize(): Promise<{ success: boolean }> {
        console.log("[FakeIAP] Initialized");
        return { success: true };
    }

    async getProducts(options: { productIds: string[] }): Promise<{ products: Product[] }> {
        const filtered = this.products.filter(p => options.productIds.includes(p.productId));
        console.log("[FakeIAP] Returning products", filtered);
        return { products: filtered };
    }

    async purchaseProduct(options: { productId: string }): Promise<PurchaseResult> {
        const purchase: Purchase = {
            transactionId: "txn_" + Date.now(),
            productId: options.productId,
            purchaseDate: Date.now(),
            state: PurchaseState.PURCHASED,
        };
        this.purchases.push(purchase);
        console.log("[FakeIAP] Purchased", purchase);
        return { code: ResponseCode.SUCCESS, purchase };
    }

    async getPurchases(): Promise<{ purchases: Purchase[] }> {
        console.log("[FakeIAP] Returning purchases", this.purchases);
        return { purchases: this.purchases };
    }
}
