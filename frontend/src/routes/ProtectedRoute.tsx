import type { ReactNode } from "react";
import { Navigate, useLocation } from "react-router-dom";
import { useAuth } from "@/contexts/AuthContext";
import { Spinner } from "@/components/Spinner";

// Proteção de interface. A proteção real está no backend: toda rota /admin
// exige sessão válida e token CSRF.
export function ProtectedRoute({ children }: { children: ReactNode }) {
  const { admin, isLoading } = useAuth();
  const location = useLocation();
  if (isLoading) {
    return (
      <div className="grid min-h-dvh place-items-center bg-brand text-2xl text-white" role="status">
        <span className="flex items-center gap-3">
          <Spinner /> Verificando acesso…
        </span>
      </div>
    );
  }
  if (!admin) return <Navigate to="/admin/login" replace state={{ from: location.pathname + location.search }} />;
  return <>{children}</>;
}
