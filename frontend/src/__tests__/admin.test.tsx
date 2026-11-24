import { screen, waitFor, within } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { http, HttpResponse } from "msw";
import { API, server, session } from "@/test/server";
import { renderRoute } from "@/test/render";

function loggedIn() {
  server.use(http.get(`${API}/auth/me`, () => HttpResponse.json(session)));
}

describe("área administrativa", () => {
  it("redireciona para o login sem sessão", async () => {
    const { router } = renderRoute("/admin");
    await waitFor(() => expect(router.state.location.pathname).toBe("/admin/login"));
    expect(await screen.findByRole("heading", { name: "Área dos anfitriões" })).toBeInTheDocument();
  });

  it("mostra erro genérico em credenciais inválidas e limpa a senha", async () => {
    const user = userEvent.setup();
    server.use(
      http.post(`${API}/auth/login`, () => HttpResponse.json({ error: { code: "unauthorized", message: "E-mail ou senha incorretos." } }, { status: 401 })),
    );
    renderRoute("/admin/login");
    await user.type(await screen.findByLabelText("E-mail"), "x@y.com");
    await user.type(screen.getByLabelText("Senha"), "errada");
    await user.click(screen.getByRole("button", { name: "Entrar" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("E-mail ou senha incorretos.");
    expect(screen.getByLabelText("Senha")).toHaveValue("");
  });

  it("entra e lista os convidados com suas escolhas", async () => {
    const user = userEvent.setup();
    server.use(http.post(`${API}/auth/login`, () => HttpResponse.json(session)));
    const { router } = renderRoute("/admin/login");
    await user.type(await screen.findByLabelText("E-mail"), "sofia@casa.com");
    await user.type(screen.getByLabelText("Senha"), "senha-forte");
    await user.click(screen.getByRole("button", { name: "Entrar" }));
    await waitFor(() => expect(router.state.location.pathname).toBe("/admin"));
    const table = await screen.findByRole("table");
    expect(within(table).getByText("Ana Lima")).toBeInTheDocument();
    expect(within(table).getAllByText("Presença")).toHaveLength(2);
    expect(within(table).getByText("Contribuição")).toBeInTheDocument();
    expect(screen.getByText("Olá, Sofia")).toBeInTheDocument();
  });

  it("filtra e busca pelo servidor", async () => {
    loggedIn();
    const user = userEvent.setup();
    const urls: string[] = [];
    server.use(
      http.get(`${API}/admin/registrations`, ({ request }) => {
        urls.push(new URL(request.url).search);
        return HttpResponse.json({ items: [], total: 0, page: 1, page_size: 15 });
      }),
    );
    renderRoute("/admin");
    await user.click(await screen.findByRole("radio", { name: /Contribuição/ }));
    await user.type(screen.getByLabelText(/Buscar/), "ana");
    await waitFor(() => expect(urls.at(-1)).toContain("q=ana"));
    expect(urls.at(-1)).toContain("type=contribute");
    expect(await screen.findByText("Ninguém encontrado com esses filtros.")).toBeInTheDocument();
  });

  it("ordena ao clicar no cabeçalho", async () => {
    loggedIn();
    const user = userEvent.setup();
    const { router } = renderRoute("/admin");
    const table = await screen.findByRole("table");
    await user.click(within(table).getByRole("button", { name: /Nome/ }));
    await waitFor(() => expect(router.state.location.search).toContain("sort=name"));
    expect(router.state.location.search).toContain("order=asc");
  });

  it("abre os detalhes, exibe mensagens como texto e exclui após confirmação", async () => {
    loggedIn();
    const user = userEvent.setup();
    let deleted: { id?: string; csrf?: string | null } = {};
    server.use(
      http.delete(`${API}/admin/registrations/:id`, ({ params, request }) => {
        deleted = { id: params.id as string, csrf: request.headers.get("X-CSRF-Token") };
        return new HttpResponse(null, { status: 204 });
      }),
    );
    renderRoute("/admin");
    const table = await screen.findByRole("table");
    await user.click(within(table).getByRole("row", { name: /Ana Lima/ }));

    const dialog = await screen.findByRole("dialog");
    expect(within(dialog).getByText("Felicidades <b>no lar</b>!")).toBeInTheDocument(); // não interpreta HTML
    await user.click(within(dialog).getByRole("button", { name: "Excluir inscrição" }));

    const confirm = await screen.findByRole("alertdialog");
    expect(confirm).toHaveTextContent("Não dá para desfazer");
    await user.click(within(confirm).getByRole("button", { name: "Excluir" }));
    await waitFor(() => expect(deleted.id).toBe("11111111-1111-4111-8111-111111111111"));
    expect(deleted.csrf).toBe("csrf-123");
  });

  it("encerra a sessão ao sair", async () => {
    loggedIn();
    const user = userEvent.setup();
    let loggedOut = false;
    server.use(
      http.post(`${API}/auth/logout`, () => {
        loggedOut = true;
        return new HttpResponse(null, { status: 204 });
      }),
    );
    const { router } = renderRoute("/admin");
    await user.click(await screen.findByRole("button", { name: "Sair" }));
    await waitFor(() => expect(router.state.location.pathname).toBe("/admin/login"));
    expect(loggedOut).toBe(true);
  });
});
