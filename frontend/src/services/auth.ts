import { ApiError, request, setCsrfToken } from "./api";
import type { SessionInfo } from "@/types/api";

export async function fetchSession(): Promise<SessionInfo | null> {
  try {
    const s = await request<SessionInfo>("/auth/me");
    setCsrfToken(s.csrf_token);
    return s;
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      setCsrfToken(null);
      return null;
    }
    throw err;
  }
}

export async function login(email: string, password: string): Promise<SessionInfo> {
  const s = await request<SessionInfo>("/auth/login", { method: "POST", body: { email, password } });
  setCsrfToken(s.csrf_token);
  return s;
}

export async function logout(): Promise<void> {
  try {
    await request<void>("/auth/logout", { method: "POST" });
  } finally {
    setCsrfToken(null);
  }
}
