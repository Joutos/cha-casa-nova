import { apiUrl, request } from "./api";
import type { GuestDetail, GuestListQuery, GuestSummary, Paginated, Stats } from "@/types/api";

export function toSearchParams(q: Partial<GuestListQuery>) {
  const params = new URLSearchParams();
  Object.entries(q).forEach(([k, v]) => {
    if (v !== undefined && v !== "" && v !== null) params.set(k, String(v));
  });
  return params;
}

export const listGuests = (q: GuestListQuery, signal?: AbortSignal) =>
  request<Paginated<GuestSummary>>(`/admin/registrations?${toSearchParams(q)}`, { signal });

export const getGuest = (id: string, signal?: AbortSignal) =>
  request<GuestDetail>(`/admin/registrations/${encodeURIComponent(id)}`, { signal });

export const getStats = (signal?: AbortSignal) => request<Stats>("/admin/stats", { signal });

export interface GuestPatch {
  name?: string;
  attending?: boolean;
  wants_to_contribute?: boolean;
}

export const updateGuest = (id: string, patch: GuestPatch) =>
  request<GuestDetail>(`/admin/registrations/${encodeURIComponent(id)}`, { method: "PATCH", body: patch });

export const deleteGuest = (id: string) =>
  request<void>(`/admin/registrations/${encodeURIComponent(id)}`, { method: "DELETE" });

// O download é um GET autenticado pelo cookie; o navegador cuida do arquivo.
export const exportUrl = (q: Partial<GuestListQuery>) => {
  const { page: _p, page_size: _s, ...filters } = q;
  void _p;
  void _s;
  return apiUrl(`/admin/registrations/export?${toSearchParams(filters)}`);
};
