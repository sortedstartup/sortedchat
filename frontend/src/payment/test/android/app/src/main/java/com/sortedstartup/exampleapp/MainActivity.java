package com.sortedstartup.exampleapp;

import com.getcapacitor.BridgeActivity;
import android.os.Bundle;
import com.sortedstartup.exampleapp.plugins.PaymentPlugin;

public class MainActivity extends BridgeActivity {
    @Override
    protected void onCreate(Bundle savedInstanceState) {
        registerPlugin(PaymentPlugin.class);
        super.onCreate(savedInstanceState);
    }
}
