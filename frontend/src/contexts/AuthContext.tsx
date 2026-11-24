import { createContext, useContext, useMemo, type ReactNode } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { fetchSession, login as apiLogin, logout as apiLogout } from "@/services/auth";
import type { Admin } from "@/types/api";

interface AuthValue {
  admin: Admin | null;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
}

const AuthContext = createContext<AuthValue | null>(null);
export const SESSION_KEY = ["session"] as const;

// A sessão vive num cookie HttpOnly; aqui só guardamos quem está logado.
export function AuthProvider({ children }: { children: ReactNode }) {
  const qc = useQueryClient();
  const session = useQuery({ queryKey: SESSION_KEY, queryFn: fetchSession, staleTime: 5 * 60_000, retry: false });

  const loginMutation = useMutation({
    mutationFn: ({ email, password }: { email: string; password: string }) => apiLogin(email, password),
    onSuccess: (s) => qc.setQueryData(SESSION_KEY, s),
  });

  const logoutMutation = useMutation({
    mutationFn: apiLogout,
    onSettled: () => {
      qc.removeQueries({ queryKey: ["admin"] });
      qc.setQueryData(SESSION_KEY, null);
    },
  });

  const value = useMemo<AuthValue>(
    () => ({
      admin: session.data?.admin ?? null,
      isLoading: session.isLoading,
      login: async (email, password) => {
        await loginMutation.mutateAsync({ email, password });
      },
      logout: async () => {
        await logoutMutation.mutateAsync();
      },
    }),
    [session.data, session.isLoading, loginMutation, logoutMutation],
  );
  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth precisa estar dentro de AuthProvider");
  return ctx;
}
