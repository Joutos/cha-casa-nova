import { lazy, Suspense, type ReactNode } from "react";
import { createBrowserRouter, Outlet, ScrollRestoration } from "react-router-dom";
import HomePage from "@/pages/HomePage";
import PresencePage from "@/pages/PresencePage";
import PresenceConfirmedPage from "@/pages/PresenceConfirmedPage";
import ContributePage from "@/pages/ContributePage";
import MessagePage from "@/pages/MessagePage";
import MessageSentPage from "@/pages/MessageSentPage";
import InvitePage from "@/pages/InvitePage";
import PrivacyPage from "@/pages/PrivacyPage";
import NotFoundPage from "@/pages/NotFoundPage";
import { ProtectedRoute } from "./ProtectedRoute";
import { AuthProvider } from "@/contexts/AuthContext";
import { Spinner } from "@/components/Spinner";

// O painel é carregado sob demanda: convidados não baixam o código administrativo.
const LoginPage = lazy(() => import("@/pages/admin/LoginPage"));
const DashboardPage = lazy(() => import("@/pages/admin/DashboardPage"));

const Loading = () => (
  <div className="grid min-h-dvh place-items-center bg-brand text-white" role="status">
    <Spinner className="size-8" />
    <span className="sr-only">Carregando</span>
  </div>
);
const lazyPage = (node: ReactNode) => <Suspense fallback={<Loading />}>{node}</Suspense>;

function Root() {
  return (
    <>
      <ScrollRestoration />
      <Outlet />
    </>
  );
}

export const routes = [
  {
    element: <Root />,
    children: [
      { path: "/", element: <HomePage /> },
      { path: "/presenca", element: <PresencePage /> },
      { path: "/presenca/confirmado", element: <PresenceConfirmedPage /> },
      { path: "/contribuir", element: <ContributePage /> },
      { path: "/mensagem", element: <MessagePage /> },
      { path: "/mensagem/enviada", element: <MessageSentPage /> },
      { path: "/convite/:token", element: <InvitePage /> },
      { path: "/privacidade", element: <PrivacyPage /> },
      {
        // A sessão só é consultada dentro da área administrativa.
        element: (
          <AuthProvider>
            <Outlet />
          </AuthProvider>
        ),
        children: [
          { path: "/admin/login", element: lazyPage(<LoginPage />) },
          { path: "/admin", element: <ProtectedRoute>{lazyPage(<DashboardPage />)}</ProtectedRoute> },
        ],
      },
      { path: "*", element: <NotFoundPage /> },
    ],
  },
];

export const routerFuture = {
  v7_relativeSplatPath: true,
  v7_fetcherPersist: true,
  v7_normalizeFormMethod: true,
  v7_partialHydration: true,
  v7_skipActionErrorRevalidation: true,
} as const;

export const createRouter = () => createBrowserRouter(routes, { future: routerFuture });
