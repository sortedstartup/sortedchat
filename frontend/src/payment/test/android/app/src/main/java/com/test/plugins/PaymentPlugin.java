package com.test.plugins;

import com.getcapacitor.Plugin;
import com.getcapacitor.PluginCall;
import com.getcapacitor.PluginMethod;
import com.getcapacitor.annotation.CapacitorPlugin;
import com.getcapacitor.JSObject;
import com.getcapacitor.JSArray;
import android.util.Log;


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
import android.app.Activity;


import androidx.annotation.NonNull;
import java.util.List;
import java.util.ArrayList;

@CapacitorPlugin(name = "PaymentPlugin")
public class PaymentPlugin extends Plugin {
    
    private List<ProductDetails> productDetailsList;
    private BillingClient billingClient;
    
    @PluginMethod
    public void initialize(PluginCall call) {
        String packageName = getContext().getPackageName();
        Log.d("PaymentPlugin", "App Package Name: " + packageName);
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

            Log.i("InAppPaymentPlugin", "Starting Billing connection");

            billingClient.startConnection(new BillingClientStateListener() {
                @Override
                public void onBillingSetupFinished(BillingResult billingResult) {
                    if (billingResult.getResponseCode() == BillingResponseCode.OK) {
                        JSObject jsObject = new JSObject();
                        jsObject.put("success", true);
                        Log.i("InAppPaymentPlugin", "Billing connection successful");
                        getProducts(call); // get products from the billing client
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

    // @PluginMethod
    public void getProducts(PluginCall call) {
        try {
            Log.i("InAppPaymentPlugin", "Getting products");
            List<Product> productList = new ArrayList<>();
            productList.add(
                Product.newBuilder()
                    .setProductId("exampleproduct1") 
                    .setProductType(ProductType.INAPP)
                    .build()
            );

            QueryProductDetailsParams queryProductDetailsParams =
                QueryProductDetailsParams.newBuilder()
                    .setProductList(productList)
                    .build();

            JSArray products = new JSArray();
            
            billingClient.queryProductDetailsAsync(
                queryProductDetailsParams,
                new ProductDetailsResponseListener() {
                    public void onProductDetailsResponse(BillingResult billingResult,
                            QueryProductDetailsResult queryProductDetailsResult) {
                        if (billingResult.getResponseCode() == BillingResponseCode.OK) {
                            Log.i("InAppPaymentPlugin getProducts sanskar", "Products: " + queryProductDetailsResult.getProductDetailsList().toString());
                        for (ProductDetails productDetails : queryProductDetailsResult.getProductDetailsList()) {
                            // Process success retrieved product details here.
                            Log.i("InAppPaymentPlugin getProducts sanskar", "Product: " + productDetails.getProductId());
                            JSObject product = new JSObject();
                            product.put("productId", productDetails.getProductId());
                            product.put("type", productDetails.getProductType());
                            product.put("title", productDetails.getName());
                            product.put("description", productDetails.getDescription());
                            products.put(product);
                            Log.i("InAppPaymentPlugin getProducts sanskar", "Products: " + products.toString());
                        }
                        JSObject result = new JSObject();
                        result.put("products", products);
                        call.resolve(result);
                        } else {
                            Log.e("InAppPaymentPlugin getProducts sanskar", "Failed: " + billingResult.getDebugMessage());
                            call.reject("Failed: " + billingResult.getDebugMessage());
                        }
                    }
                });
        } catch (Exception e) {
            Log.e("InAppPaymentPlugin getProducts sanskar", "Failed: " + e.getMessage());
            call.reject("Failed: " + e.getMessage());
        }
    }


    @PluginMethod
public void purchaseProduct(PluginCall call) {
    try {
        String productId = call.getString("productId", "exampleproduct1");
        Activity activity = getActivity();
        ProductDetails productDetails = this.productDetailsList.get(0);
        Log.i("InAppPaymentPlugin", "Launching billing flow for product: " + productDetails.getProductId());
         billingClient.launchBillingFlow(activity,
                    BillingFlowParams
                            .newBuilder()
                            .setProductDetailsParamsList(
                                    List.of(
                                            BillingFlowParams.ProductDetailsParams.newBuilder()
                                                    // retrieve a value for "productDetails" by calling queryProductDetailsAsync()
                                                    .setProductDetails(productDetails)

                                                    // For one-time products, "setOfferToken" method shouldn't be called.
                                                    // For subscriptions, to get an offer token, call
                                                    // ProductDetails.subscriptionOfferDetails() for a list of offers
                                                    // that are available to the user.
                                                    //.setOfferToken(selectedOfferToken)
                                                    .build()
                                    ))
                            .build());
                            JSObject ret = new JSObject();
                    ret.put("code", 0);
                    call.resolve(ret);
    } catch (Exception e) {
        Log.e("InAppPaymentPlugin", "Error launching billing flow: " + e.getMessage());
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

