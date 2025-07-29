import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { Chat } from "./pages/chat";
import { Project } from "./pages/project";
import { Home } from "./pages/home";
import { Layout } from "./layout";

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
    ],
  },
]);

function App() {
  return <RouterProvider router={router} />;
}

export default App;