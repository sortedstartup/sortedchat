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
  
  React.useEffect(() => {
    GetIsFirstBootStatus().then(setIsFirstBoot);
  }, []);
  
  // Show loading state while checking
  if (isFirstBoot === null) {
    return <div>Loading...</div>; // Or your loading component
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
]);

function App() {
  return <RouterProvider router={router} />;
}

export default App;
