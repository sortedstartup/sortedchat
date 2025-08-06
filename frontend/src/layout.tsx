import { Outlet } from "react-router-dom";
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { AppSidebar } from "@/components/app-sidebar";
import { Toaster } from "@/components/ui/sonner";

export function Layout() {
  return (
    <SidebarProvider className="h-screen">
      <AppSidebar />
      <main className="flex-1 flex flex-col h-full">
        <div className="flex-shrink-0">
          <SidebarTrigger />
        </div>
        <div className="flex-1 overflow-hidden">
          <Outlet />
        </div>
      </main>
      <Toaster />
    </SidebarProvider>
  );
}