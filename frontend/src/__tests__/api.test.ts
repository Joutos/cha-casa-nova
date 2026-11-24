import { http, HttpResponse } from "msw";
import { API, server } from "@/test/server";
import { ApiError, request, setCsrfToken } from "@/services/api";

describe("cliente da API", () => {
  afterEach(() => setCsrfToken(null));

  it("converte erros da API em ApiError com campos", async () => {
    server.use(
      http.post(`${API}/registrations`, () =>
        HttpResponse.json({ error: { code: "validation_error", message: "Alguns campos", fields: { email: "Confira o e-mail" } } }, { status: 422 }),
      ),
    );
    const err = (await request("/registrations", { method: "POST", body: {} }).catch((e: ApiError) => e)) as ApiError;
    expect(err).toBeInstanceOf(ApiError);
    expect(err.status).toBe(422);
    expect(err.fields).toEqual({ email: "Confira o e-mail" });
  });

  it("envia o token CSRF apenas em métodos que alteram estado", async () => {
    const seen: Record<string, string | null> = {};
    server.use(
      http.get(`${API}/admin/stats`, ({ request: r }) => {
        seen.get = r.headers.get("X-CSRF-Token");
        return HttpResponse.json({});
      }),
      http.delete(`${API}/admin/registrations/:id`, ({ request: r }) => {
        seen.delete = r.headers.get("X-CSRF-Token");
        return new HttpResponse(null, { status: 204 });
      }),
    );
    setCsrfToken("tok");
    await request("/admin/stats");
    await request("/admin/registrations/x", { method: "DELETE" });
    expect(seen).toEqual({ get: null, delete: "tok" });
  });

  it("explica falhas de rede", async () => {
    server.use(http.get(`${API}/health`, () => HttpResponse.error()));
    const err = (await request("/health").catch((e: ApiError) => e)) as ApiError;
    expect(err.code).toBe("network_error");
  });
});
