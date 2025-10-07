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
import { OnboardingPage } from "./routes/onboarding";
import { useFirstBoot } from "./hooks/useFirstBoot";
import { Loader2 } from "lucide-react";

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
  const { isFirstBoot, isLoading, error } = useFirstBoot();
  
  // Show loading spinner while checking first boot status
  if (isLoading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <Loader2 className="h-8 w-8 animate-spin mx-auto mb-4" />
          <p className="text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }
  
  // Show error if first boot check failed
  if (error) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <p className="text-red-600 mb-4">Failed to initialize application</p>
          <p className="text-gray-600">{error}</p>
        </div>
      </div>
    );
  }
  
  // Show onboarding if it's the first boot
  if (isFirstBoot) {
    return <OnboardingPage />;
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
  {
    path: "*",
    element: (
      <ProtectedRoute>
        <Layout />
      </ProtectedRoute>
    ),
  },
]);

function App() {
  return <RouterProvider router={router} />;
}

export default App;
