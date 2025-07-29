import { createBrowserRouter, RouterProvider } from "react-router-dom";
import { Chat } from "./pages/chat";
import { Project } from "./pages/project";

const router = createBrowserRouter([
  { path: "/", element: <Chat /> },
  // { path: "/login", element: <Login /> },
  { path: "/project", element: <Project /> },
  // { path: "/setting", element: <Setting /> },
]);

function App() {
  return <RouterProvider router={router} />;
}

export default App;