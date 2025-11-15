import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { Chat } from "./pages/chat";
import { Project } from "./pages/project";
import { Home } from "./pages/home";
import { Layout } from "./layout";
import SettingsPage from "./pages/setting";
import Models from "./pages/models";
import { useStore } from "@nanostores/react";
import { $auth } from "./auth/store/auth";
import { LoginPage } from "./auth/pages/login";
import { OnboardingWizard } from "./pages/onboard";
import { GetIsFirstBootStatus } from "./store/setting";
import React from "react";

// Protected route wrapper component with onboarding check
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const auth = useStore($auth);

  // Redirect to login if not authenticated
  if (!auth.isLoggedIn) {
    return <LoginPage />;
  }

  // Only check first boot after authentication
  return <AuthenticatedRoute>{children}</AuthenticatedRoute>;
}

// Component that handles first boot check after authentication
function AuthenticatedRoute({ children }: { children: React.ReactNode }) {
  const [isFirstBoot, setIsFirstBoot] = React.useState<boolean | null>(null);
  const [isLoading, setIsLoading] = React.useState(true);

  React.useEffect(() => {
    const checkFirstBoot = async () => {
      setIsLoading(true);
      try {
        const status = await GetIsFirstBootStatus();
        setIsFirstBoot(status);
      } catch (error) {
        console.error("Failed to check first boot status:", error);
        // On error, assume not first boot to avoid blocking user
        setIsFirstBoot(false);
      } finally {
        setIsLoading(false);
      }
    };

    checkFirstBoot();
  }, []);

  // Show loading state while checking
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        Loading...
      </div>
    );
  }

  // Show onboarding if it's the first boot
  if (isFirstBoot) {
    return <OnboardingWizard />;
  }

  return <>{children}</>;
}

const router = createBrowserRouter([
  {
    path: "/login",
    element: <LoginPage />,
  },
  {
    path: "/",
    element: (
      <ProtectedRoute>
        <Layout />
      </ProtectedRoute>
    ),
    children: [
      {
        index: true,
        element: <Home />,
      },
      {
        path: "chat/:chatId",
        element: <Chat />,
      },
      {
        path: "project/:projectId",
        element: <Project />,
      },
      {
        path: "project/:projectId/chat/:chatId",
        element: <Chat />,
      },
      {
        path: "setting",
        element: <SettingsPage />,
      },
      {
        path: "models",
        element: <Models />,
      },
      {
        path: "*",
        element: <Home />,
      },
    ],
  },
]);

function App() {
  return <RouterProvider router={router} />;
}

export default App;
