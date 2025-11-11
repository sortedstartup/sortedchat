package com.test.plugins;

import com.getcapacitor.Plugin;
import com.getcapacitor.PluginCall;
import com.getcapacitor.PluginMethod;
import com.getcapacitor.annotation.CapacitorPlugin;
import com.getcapacitor.JSObject;
import com.getcapacitor.JSArray;

import com.android.billingclient.api.BillingClient;
import com.android.billingclient.api.BillingClient.BillingResponseCode;
import com.android.billingclient.api.BillingClient.ProductType;
import com.android.billingclient.api.BillingClientStateListener;
import com.android.billingclient.api.BillingResult;
import com.android.billingclient.api.Purchase;
import com.android.billingclient.api.PurchasesUpdatedListener;
import com.android.billingclient.api.PendingPurchasesParams;
import com.android.billingclient.api.ProductDetails;
import com.android.billingclient.api.QueryProductDetailsParams;
import com.android.billingclient.api.QueryProductDetailsParams.Product;
import com.android.billingclient.api.QueryProductDetailsResult;
import com.android.billingclient.api.BillingFlowParams;
import com.android.billingclient.api.QueryPurchasesParams;
import com.android.billingclient.api.AcknowledgePurchaseParams;
import com.android.billingclient.api.ProductDetailsResponseListener;
import com.android.billingclient.api.PurchasesResponseListener;
import com.android.billingclient.api.BillingFlowParams;
import com.android.billingclient.api.BillingFlowParams.ProductDetailsParams;
import com.android.billingclient.api.QueryPurchasesParams;


import androidx.annotation.NonNull;
import java.util.List;
import java.util.ArrayList;

@CapacitorPlugin(name = "PaymentPlugin")
public class PaymentPlugin extends Plugin {
    
    private BillingClient billingClient;
    
    @PluginMethod
    public void initialize(PluginCall call) {
        try {
            PurchasesUpdatedListener purchasesUpdatedListener = new PurchasesUpdatedListener() {
                @Override
                public void onPurchasesUpdated(BillingResult billingResult, List<Purchase> purchases) {
                    // To be implemented in a later section.
                }
            };

            PendingPurchasesParams pendingPurchasesParams = 
                PendingPurchasesParams.newBuilder()
                    .enableOneTimeProducts()
                    .build();

            billingClient = BillingClient.newBuilder(getContext())
                .setListener(purchasesUpdatedListener)
                .enablePendingPurchases(pendingPurchasesParams)
                .build();

            billingClient.startConnection(new BillingClientStateListener() {
                @Override
                public void onBillingSetupFinished(BillingResult billingResult) {
                    if (billingResult.getResponseCode() == BillingResponseCode.OK) {
                        JSObject jsObject = new JSObject();
                        jsObject.put("success", true);
                        call.resolve(jsObject);
                    } else {
                        call.reject("Billing setup failed: " + billingResult.getDebugMessage());
                    }
                }
                
                @Override
                public void onBillingServiceDisconnected() {
                    call.reject("Billing service disconnected");
                }
            });
            
        } catch (Exception e) {
            call.reject("Error initializing billing: " + e.getMessage());
        }
    }

    @PluginMethod
    public void getProducts(PluginCall call) {
        try {

            List<Product> productList = new ArrayList<>();
            productList.add(
                Product.newBuilder()
                    .setProductId("androd.test.purchased") 
                    .setProductType(ProductType.INAPP)
                    .build()
            );

            QueryProductDetailsParams queryProductDetailsParams =
                QueryProductDetailsParams.newBuilder()
                    .setProductList(productList)
                    .build();

            billingClient.queryProductDetailsAsync(
                queryProductDetailsParams,
                new ProductDetailsResponseListener() {
                    @Override
                    public void onProductDetailsResponse(
                        BillingResult billingResult,
                        QueryProductDetailsResult queryProductDetailsResult
                    ) {
                        if (billingResult.getResponseCode() == BillingResponseCode.OK) {
                            
                            JSArray products = new JSArray();
                            
                            for (ProductDetails productDetails : queryProductDetailsResult.getProductDetailsList()) {
                                JSObject product = new JSObject();
                                product.put("productId", productDetails.getProductId());
                                product.put("type", productDetails.getProductType());
                                product.put("title", productDetails.getName());
                                product.put("description", productDetails.getDescription());
                                //will add price amount and currency
                                
                                products.put(product);
                            }
                            
                            JSObject result = new JSObject();
                            result.put("products", products);
                            call.resolve(result);
                            
                        } else {
                            call.reject("Failed: " + billingResult.getDebugMessage());
                        }
                    }
                }
            );
        } catch (Exception e) {
            call.reject("Error: " + e.getMessage());
        }
    }

    @PluginMethod
public void purchaseProduct(PluginCall call) {
    try {
        String productId = call.getString("productId", "andriod.test.purchased");
        
        // Build simple product list
        List<Product> productList = new ArrayList<>();
        productList.add(
            Product.newBuilder()
                .setProductId(productId)
                .setProductType(ProductType.INAPP)
                .build()
        );

        QueryProductDetailsParams queryParams =
            QueryProductDetailsParams.newBuilder()
                .setProductList(productList)
                .build();

        billingClient.queryProductDetailsAsync(queryParams, 
            new ProductDetailsResponseListener() {
                @Override
                public void onProductDetailsResponse(
                    BillingResult billingResult,
                    QueryProductDetailsResult queryProductDetailsResult
                ) {
                    // Get product details
                    ProductDetails productDetails = 
                        queryProductDetailsResult.getProductDetailsList().get(0);
                    
                    // Build params
                    List<ProductDetailsParams> paramsList = new ArrayList<>();
                    paramsList.add(
                        ProductDetailsParams.newBuilder()
                            .setProductDetails(productDetails)
                            .build()
                    );

                    // Launch payment UI
                    billingClient.launchBillingFlow(
                        getActivity(), 
                        BillingFlowParams.newBuilder()
                            .setProductDetailsParamsList(paramsList)
                            .build()
                    );
                    
                    // Just return success
                    JSObject ret = new JSObject();
                    ret.put("code", 0);
                    call.resolve(ret);
                }
            }
        );
        
    } catch (Exception e) {
        JSObject ret = new JSObject();
        ret.put("code", 4);
        ret.put("message", e.getMessage());
        call.resolve(ret);
    }
}




    @PluginMethod
    public void getPurchases(PluginCall call) {
        try {
            QueryPurchasesParams params = QueryPurchasesParams.newBuilder()
                .setProductType(ProductType.INAPP) 
                .build();
                
            billingClient.queryPurchasesAsync(params, new PurchasesResponseListener() {
                @Override
                public void onQueryPurchasesResponse(
                    BillingResult billingResult, 
                    List<Purchase> purchases
                ) {
                    if (billingResult.getResponseCode() == BillingResponseCode.OK) {
                        JSArray purchaseArray = new JSArray();
                        
                        for (Purchase purchase : purchases) {
                            JSObject purchaseObj = new JSObject();
                            purchaseObj.put("transactionId", purchase.getOrderId());
                            purchaseObj.put("productId", purchase.getProducts().get(0));
                            purchaseObj.put("purchaseDate", purchase.getPurchaseTime());
                            purchaseObj.put("purchaseToken", purchase.getPurchaseToken()); 
                            purchaseObj.put("state", purchase.getPurchaseState() == Purchase.PurchaseState.PURCHASED ? "purchased" : "pending"); 
                            purchaseObj.put("isAcknowledged", purchase.isAcknowledged());
                            purchaseArray.put(purchaseObj);
                        }
                        
                        JSObject ret = new JSObject();
                        ret.put("purchases", purchaseArray);
                        call.resolve(ret);
                    } else {
                        call.reject("Failed to get purchases: " + billingResult.getDebugMessage());
                    }
                }
            });
            
        } catch (Exception e) {
            call.reject("Error getting purchases: " + e.getMessage());
        }
    }

    

}

