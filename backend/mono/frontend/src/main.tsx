import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import "./index.css";
import "./console-bridge";

ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>
);