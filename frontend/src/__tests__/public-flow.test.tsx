import { screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { API, server } from "@/test/server";
import { renderRoute } from "@/test/render";

describe("página inicial", () => {
  it("mostra o convite e exige ao menos uma escolha", async () => {
    const user = userEvent.setup();
    renderRoute("/");
    expect(screen.getByRole("heading", { level: 1, name: "CHÁ DE CASA NOVA" })).toBeInTheDocument();
    expect(screen.getByText("Data: 23 de Janeiro de 2027")).toBeInTheDocument();
    await user.click(screen.getByRole("button", { name: "Continuar" }));
    expect(screen.getByRole("alert")).toHaveTextContent("Escolha pelo menos uma opção");
  });

  it("leva à primeira etapa escolhida", async () => {
    const user = userEvent.setup();
    const { router } = renderRoute("/");
    await user.click(screen.getByLabelText(/Quero deixar uma mensagem!/));
    await user.click(screen.getByLabelText(/Quero contribuir com o nosso lar!/));
    await user.click(screen.getByRole("button", { name: "Continuar" }));
    await waitFor(() => expect(router.state.location.pathname).toBe("/contribuir"));
  });
});

describe("fluxo de presença", () => {
  it("valida os campos antes de enviar", async () => {
    const user = userEvent.setup();
    renderRoute("/presenca");
    await user.click(screen.getByRole("button", { name: "confirmar" }));
    expect(await screen.findByText("Informe seu nome.")).toBeInTheDocument();
    expect(screen.getByText("Informe seu e-mail.")).toBeInTheDocument();
    expect(screen.getByLabelText("Nome")).toHaveAttribute("aria-invalid", "true");
  });

  it("confirma, bloqueia duplo envio e segue para a próxima etapa escolhida", async () => {
    const user = userEvent.setup();
    let calls = 0;
    let body: Record<string, unknown> = {};
    server.use(
      http.post(`${API}/registrations`, async ({ request }) => {
        calls++;
        body = (await request.json()) as Record<string, unknown>;
        await new Promise((r) => setTimeout(r, 50));
        return HttpResponse.json({ message: "ok" }, { status: 201 });
      }),
    );
    const { router } = renderRoute("/");
    await user.click(screen.getByLabelText(/Quero estar presente!/));
    await user.click(screen.getByLabelText(/Quero contribuir com o nosso lar!/));
    await user.click(screen.getByRole("button", { name: "Continuar" }));

    await user.type(await screen.findByLabelText("Nome"), "  Ana   Lima ");
    await user.type(screen.getByLabelText("E-mail"), "ana@ex.com");
    const submit = screen.getByRole("button", { name: "confirmar" });
    await user.click(submit);
    await user.click(submit);

    await waitFor(() => expect(router.state.location.pathname).toBe("/presenca/confirmado"));
    expect(calls).toBe(1);
    expect(body).toEqual({ name: "Ana Lima", email: "ana@ex.com", website: "", contribute: true });

    // Ainda falta a contribuição: aparece "continuar" (Frame 4), não "Retornar".
    await user.click(await screen.findByRole("button", { name: "continuar" }));
    await waitFor(() => expect(router.state.location.pathname).toBe("/contribuir"));
    // Última etapa: volta para a página inicial (Frame 5).
    expect(await screen.findByRole("button", { name: "Retornar para página inicial" })).toBeInTheDocument();
  });

  it("mostra o aviso do servidor quando o e-mail já confirmou", async () => {
    const user = userEvent.setup();
    server.use(
      http.post(`${API}/registrations`, () =>
        HttpResponse.json({ error: { code: "already_confirmed", message: "Este e-mail já confirmou presença." } }, { status: 409 }),
      ),
    );
    renderRoute("/presenca");
    await user.type(screen.getByLabelText("Nome"), "Ana");
    await user.type(screen.getByLabelText("E-mail"), "ana@ex.com");
    await user.click(screen.getByRole("button", { name: "confirmar" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("Este e-mail já confirmou presença.");
  });

  it("associa erros de campo vindos do backend", async () => {
    const user = userEvent.setup();
    server.use(
      http.post(`${API}/registrations`, () =>
        HttpResponse.json({ error: { code: "validation_error", message: "x", fields: { email: "Domínio inválido." } } }, { status: 422 }),
      ),
    );
    renderRoute("/presenca");
    await user.type(screen.getByLabelText("Nome"), "Ana");
    await user.type(screen.getByLabelText("E-mail"), "ana@ex.com");
    await user.click(screen.getByRole("button", { name: "confirmar" }));
    expect(await screen.findByText("Domínio inválido.")).toBeInTheDocument();
  });
});

describe("mensagem", () => {
  it("envia a mensagem com nome e e-mail e agradece", async () => {
    const user = userEvent.setup();
    let body: Record<string, unknown> = {};
    server.use(
      http.post(`${API}/messages`, async ({ request }) => {
        body = (await request.json()) as Record<string, unknown>;
        return HttpResponse.json({ message: "ok" }, { status: 201 });
      }),
    );
    const { router } = renderRoute("/mensagem");
    await user.type(screen.getByLabelText("Nome"), "Tia Lúcia");
    await user.type(screen.getByLabelText("E-mail"), "lucia@ex.com");
    await user.type(screen.getByLabelText("Mensagem"), "Felicidades!");
    await user.click(screen.getByRole("button", { name: "Enviar" }));
    await waitFor(() => expect(router.state.location.pathname).toBe("/mensagem/enviada"));
    expect(body).toMatchObject({ name: "Tia Lúcia", email: "lucia@ex.com", message: "Felicidades!" });
    expect(await screen.findByRole("heading", { name: /Muito obrigada por esse carinho/ })).toBeInTheDocument();
  });

  it("não envia mensagem vazia", async () => {
    const user = userEvent.setup();
    renderRoute("/mensagem");
    await user.type(screen.getByLabelText("Nome"), "Tia Lúcia");
    await user.type(screen.getByLabelText("E-mail"), "lucia@ex.com");
    await user.click(screen.getByRole("button", { name: "Enviar" }));
    expect(await screen.findByText("Escreva sua mensagem antes de enviar.")).toBeInTheDocument();
  });
});

describe("navegação", () => {
  it("mostra a página 404 para rotas desconhecidas", () => {
    renderRoute("/nao-existe");
    expect(screen.getByRole("heading", { name: "Esta página não existe" })).toBeInTheDocument();
  });

  it("mostra o convite a partir do link pessoal", async () => {
    server.use(
      http.get(`${API}/invites/:token`, () =>
        HttpResponse.json({ guest_name: "Ana Lima", date_label: "23 de Janeiro de 2027", time: "11h00", address_line1: "Av Saranhema, 1727", address_line2: "Guarani", maps_query: "Av Saranhema 1727", signature: "Sofia e João" }),
      ),
    );
    renderRoute("/convite/abc");
    expect(await screen.findByRole("heading", { name: /Que bom que você vem, Ana/ })).toBeInTheDocument();
    expect(screen.getByTitle(/Mapa/)).toHaveAttribute("src", expect.stringContaining("output=embed"));
  });
});
