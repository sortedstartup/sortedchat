import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { Chat } from "./pages/chat";
import { Project } from "./pages/project";
import { Home } from "./pages/home";

const router = createBrowserRouter([
  { path: "/", element: <Home /> },
  { path: "/chat/:chatId", element: <Chat /> },
  // { path: "/login", element: <Login /> },
  { path: "/project/:projectId", element: <Project /> },
  // { path: "/setting", element: <Setting /> },
]);

function App() {
  return <RouterProvider router={router} />;
}

export default App;