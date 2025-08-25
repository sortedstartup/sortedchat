import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { Chat } from "./pages/chat";
import { Project } from "./pages/project";
import { Home } from "./pages/home";
import { Layout } from "./layout";
import SettingsPage from "./pages/setting";
import Models from "./pages/models";

const router = createBrowserRouter([
  {
    path: "/",
    element: <Layout />,
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
