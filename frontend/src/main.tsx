import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { RouterProvider } from "react-router-dom";
import "@fontsource/hubballi/latin-400.css";
import "@fontsource/hubballi/latin-ext-400.css";
import "./index.css";
import { AppProviders } from "./App";
import { createRouter } from "./routes/router";

const router = createRouter();

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <AppProviders>
      <RouterProvider router={router} future={{ v7_startTransition: true }} />
    </AppProviders>
  </StrictMode>,
);
