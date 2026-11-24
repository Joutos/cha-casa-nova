import type { ApiErrorBody, FieldErrors } from "@/types/api";

const BASE = (import.meta.env.VITE_API_URL ?? "").replace(/\/$/, "") + "/api/v1";

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly fields?: FieldErrors;

  constructor(status: number, body: Partial<ApiErrorBody>) {
    super(body.message ?? "Não foi possível concluir agora. Tente novamente.");
    this.status = status;
    this.code = body.code ?? "unknown_error";
    this.fields = body.fields;
  }
}

// O token CSRF fica só em memória (nunca em localStorage). Ele é reemitido pelo
// endpoint /auth/me sempre que a página é recarregada.
let csrfToken: string | null = null;
export const setCsrfToken = (t: string | null) => {
  csrfToken = t;
};

type Options = { method?: "GET" | "POST" | "PATCH" | "DELETE"; body?: unknown; signal?: AbortSignal };

export async function request<T>(path: string, { method = "GET", body, signal }: Options = {}): Promise<T> {
  const headers: Record<string, string> = { Accept: "application/json" };
  if (body !== undefined) headers["Content-Type"] = "application/json";
  if (method !== "GET" && csrfToken) headers["X-CSRF-Token"] = csrfToken;

  let res: Response;
  try {
    res = await fetch(BASE + path, {
      method,
      headers,
      body: body === undefined ? undefined : JSON.stringify(body),
      credentials: "same-origin",
      signal,
    });
  } catch (err) {
    if ((err as Error).name === "AbortError") throw err;
    throw new ApiError(0, { code: "network_error", message: "Sem conexão com o servidor. Verifique sua internet e tente de novo." });
  }

  if (res.status === 204) return undefined as T;
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new ApiError(res.status, (data as { error?: ApiErrorBody }).error ?? {});
  return data as T;
}

export const apiUrl = (path: string) => BASE + path;
