import { http, HttpResponse } from "msw";
import { setupServer } from "msw/node";
import type { GuestDetail, GuestSummary } from "@/types/api";

export const API = "http://localhost:3000/api/v1";

export const guests: GuestSummary[] = [
  { id: "11111111-1111-4111-8111-111111111111", name: "Ana Lima", email: "ana@ex.com", attending: true, wants_to_contribute: true, message_count: 1, created_at: "2026-11-01T14:00:00Z", updated_at: "2026-11-01T14:00:00Z" },
  { id: "22222222-2222-4222-8222-222222222222", name: "Bruno Costa", email: "bruno@ex.com", attending: true, wants_to_contribute: false, message_count: 0, created_at: "2026-11-02T14:00:00Z", updated_at: "2026-11-02T14:00:00Z" },
];

export const anaDetail: GuestDetail = {
  ...guests[0],
  attending_confirmed_at: "2026-11-01T14:00:00Z",
  messages: [{ id: "m1", body: "Felicidades <b>no lar</b>!", created_at: "2026-11-01T15:00:00Z" }],
  invite_url: "https://festa.example/convite/abc",
};

export const session = { admin: { id: "a1", name: "Sofia Andrade", email: "sofia@casa.com" }, csrf_token: "csrf-123" };

export const handlers = [
  http.get(`${API}/auth/me`, () => HttpResponse.json({ error: { code: "unauthorized", message: "Entre para acessar esta área." } }, { status: 401 })),
  http.get(`${API}/admin/stats`, () => HttpResponse.json({ total: 2, attending: 2, contribute: 1, with_message: 1 })),
  http.get(`${API}/admin/registrations`, () => HttpResponse.json({ items: guests, total: 2, page: 1, page_size: 15 })),
  http.get(`${API}/admin/registrations/:id`, () => HttpResponse.json(anaDetail)),
];

export const server = setupServer(...handlers);
