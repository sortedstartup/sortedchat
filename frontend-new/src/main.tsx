import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar"
import { AppSidebar } from "@/components/app-sidebar"

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <SidebarProvider>
      <AppSidebar />
      <SidebarTrigger />
    <App />
    </SidebarProvider>
  </React.StrictMode>
);