package com.sortedstartup.exampleapp.plugins;

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
import okhttp3.*;
import java.io.IOException;
import org.json.JSONObject;
import org.json.JSONException;


import androidx.annotation.NonNull;
import java.util.List;
import java.util.ArrayList;
import java.util.HashMap;

@CapacitorPlugin(name = "PaymentPlugin")
public class PaymentPlugin extends Plugin {
    
    private List<ProductDetails> productDetailsList;
    private BillingClient billingClient;
    private HashMap<String, ProductDetails> productDetailsMap = new HashMap<>();
    
    @PluginMethod
    public void initialize(PluginCall call) {
        String packageName = getContext().getPackageName();
        Log.d("PaymentPlugin", "App Package Name: " + packageName);
        try {
            PurchasesUpdatedListener purchasesUpdatedListener = new PurchasesUpdatedListener() {
                @Override
                public void onPurchasesUpdated(BillingResult billingResult, List<Purchase> purchases) {
                    if (billingResult.getResponseCode() == BillingResponseCode.OK && purchases != null) {
                        for (Purchase purchase : purchases) {
                            Log.i("InAppPaymentPlugin", "Purchase: " + purchase.toString());
                            Log.i("InAppPaymentPlugin", "Making API call for purchase");
                            makeApiCall(purchase);
                        }
                    } else if (billingResult.getResponseCode() == BillingResponseCode.USER_CANCELED && purchases != null) {
                        // Handle an error caused by a user canceling the purchase flow.
                    } else if(billingResult.getResponseCode() == BillingResponseCode.ERROR && purchases != null) {
                        // Handle any other error codes.
                        Log.e("InAppPaymentPlugin", "Error: " + billingResult.getDebugMessage());
                        call.reject("Error: " + billingResult.getDebugMessage());
                    } else if(billingResult.getResponseCode() == BillingResponseCode.ITEM_ALREADY_OWNED && purchases != null) {
                        // Handle an error caused by a user already owning the item.
                        Log.e("InAppPaymentPlugin", "Item already owned: " + billingResult.getDebugMessage());
                        call.reject("Item already owned: " + billingResult.getDebugMessage());
                    } else if(billingResult.getResponseCode() == BillingResponseCode.SERVICE_TIMEOUT && purchases != null) {
                        // Handle an error caused by a user not owning the item.
                        
                    }
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
            e.printStackTrace();
            call.reject("Error initializing billing: " + e.getMessage());
        }
    }

    @PluginMethod
    public void getProducts(PluginCall call) {
        try {
            Log.i("InAppPaymentPlugin", "Getting products");
            //take product id from call
            // String productId = call.getString("productId", "exampleproduct1");

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
                            Log.i("InAppPaymentPlugin getProducts", "Products: " + queryProductDetailsResult.getProductDetailsList().toString());
                        for (ProductDetails productDetails : queryProductDetailsResult.getProductDetailsList()) {
                            // Process success retrieved product details here.
                            Log.i("InAppPaymentPlugin getProducts", "Product: " + productDetails.getProductId());
                            
                            // Store in HashMap
                            productDetailsMap.put(productDetails.getProductId(), productDetails);
                            
                            JSObject product = new JSObject();
                            product.put("productId", productDetails.getProductId());
                            product.put("type", productDetails.getProductType());
                            product.put("title", productDetails.getName());
                            product.put("description", productDetails.getDescription());
                            products.put(product);
                            Log.i("InAppPaymentPlugin getProducts", "Products: " + products.toString());
                        }
                        JSObject result = new JSObject();
                        result.put("products", products);
                        call.resolve(result);
                        } else {
                            Log.e("InAppPaymentPlugin getProducts", "Failed: " + billingResult.getDebugMessage());
                            call.reject("Failed: " + billingResult.getDebugMessage());
                        }
                    }
                });
        } catch (Exception e) {
            Log.e("InAppPaymentPlugin getProducts", "Failed: " + e.getMessage());
            e.printStackTrace();
            call.reject("Failed: " + e.getMessage());
        }
    }


    @PluginMethod
public void purchaseProduct(PluginCall call) {
    try {
        String productId = call.getString("productId", "exampleproduct1");
        String accountId = call.getString("accountId", "exampleaccount1");
        String customeProductId = call.getString("customeProductId", "examplecustomeproduct1");
        Activity activity = getActivity();
        ProductDetails productDetails = productDetailsMap.get(productId);

        if (productDetails == null) {
            call.reject("Product not found in productDetailsMap at purchaseProduct");
        }
        Log.i("InAppPaymentPlugin purchaseProduct", "Product details: " + productDetails.toString());
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
                            .setObfuscatedAccountId(accountId)
                            .setObfuscatedProfileId(customeProductId)
                            .build());
                            JSObject ret = new JSObject();
                    ret.put("code", 0);
                    call.resolve(ret);
    } catch (Exception e) {
        Log.e("InAppPaymentPlugin", "Error launching billing flow: " + e.getMessage());
        e.printStackTrace();
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


    private void makeApiCall(Purchase purchase) {
        Log.i("InAppPaymentPlugin:makeApiCall", "Making API call for purchase: " + purchase.toString());
        OkHttpClient client = new OkHttpClient();
        
        try {
            // Create JSON payload
            JSONObject json = new JSONObject();
            json.put("transactionId", purchase.getOrderId());
            json.put("productId", purchase.getProducts().get(0));
            json.put("purchaseToken", purchase.getPurchaseToken());
            json.put("purchaseTime", purchase.getPurchaseTime());
            json.put("purchaseState", purchase.getPurchaseState());
            json.put("packageName", purchase.getPackageName());
            json.put("accountId", purchase.getAccountIdentifiers().getObfuscatedAccountId());
            json.put("customProductId", purchase.getAccountIdentifiers().getObfuscatedProfileId());

            RequestBody body = RequestBody.create(
                json.toString(),
                MediaType.get("application/json; charset=utf-8")
            );
            
            Request request = new Request.Builder()
                .url("https://intercrural-brycen-anthropometrically.ngrok-free.dev/inapp-purchase-product")
                .post(body)
                .addHeader("Content-Type", "application/json")
                .build();
            
            client.newCall(request).enqueue(new Callback() {
                @Override
                public void onFailure(Call call, IOException e) {
                    Log.e("InAppPaymentPlugin", "API call failed: " + e.getMessage());
                }
                
                @Override
                public void onResponse(Call call, Response response) throws IOException {
                    if (response.isSuccessful()) {
                        Log.i("InAppPaymentPlugin", "API call successful: " + response.body().string());
                    } else {
                        Log.e("InAppPaymentPlugin", "API call failed with code: " + response.code());
                    }
                }
            });
            
        } catch (JSONException e) {
            e.printStackTrace();
            Log.e("InAppPaymentPlugin", "Error creating JSON: " + e.getMessage());
        }
    }
}

