import { expect, test } from "@playwright/test";

const ADMIN_EMAIL = process.env.E2E_ADMIN_EMAIL ?? "admin@example.com";
const ADMIN_PASSWORD = process.env.E2E_ADMIN_PASSWORD ?? "troque-esta-senha-123";
const unique = () => `${Date.now()}-${Math.random().toString(36).slice(2, 7)}`;

test("convidado escolhe as três opções e percorre o fluxo completo", async ({ page }, info) => {
  const email = `e2e-${info.project.name}-${unique()}@example.com`;
  await page.goto("/");
  await expect(page.getByRole("heading", { level: 1, name: "CHÁ DE CASA NOVA" })).toBeVisible();

  await page.getByLabel("Quero estar presente!").check();
  await page.getByLabel("Quero contribuir com o nosso lar!").check();
  await page.getByLabel("Quero deixar uma mensagem!").check();
  await page.getByRole("button", { name: "Continuar" }).click();

  await expect(page).toHaveURL(/\/presenca$/);
  await page.getByLabel("Nome").fill("Convidada E2E");
  await page.getByLabel("E-mail").fill(email);
  await page.getByRole("button", { name: "confirmar" }).click();

  await expect(page).toHaveURL(/\/presenca\/confirmado$/);
  await page.getByRole("button", { name: "continuar" }).click();
  await expect(page).toHaveURL(/\/contribuir$/);
  await page.getByRole("button", { name: "continuar" }).click();

  await expect(page).toHaveURL(/\/mensagem$/);
  await expect(page.getByText(/Enviando como Convidada E2E/)).toBeVisible();
  await page.getByLabel("Mensagem").fill("Felicidades no novo lar!");
  await page.getByRole("button", { name: "Enviar" }).click();
  await expect(page.getByRole("heading", { name: /Muito obrigada/ })).toBeVisible();

  // Reenvio com o mesmo e-mail é recusado.
  await page.goto("/presenca");
  await page.getByLabel("Nome").fill("Convidada E2E");
  await page.getByLabel("E-mail").fill(email);
  await page.getByRole("button", { name: "confirmar" }).click();
  await expect(page.getByRole("alert")).toContainText("já confirmou presença");
});

test("anfitriã entra, encontra o convidado, lê a mensagem e sai", async ({ page }) => {
  const email = `e2e-admin-${unique()}@example.com`;
  const res = await page.request.post("/api/v1/messages", {
    data: { name: "Pessoa E2E", email, message: "Mensagem para o painel", contribute: true, website: "" },
  });
  expect(res.status()).toBe(201);

  await page.goto("/admin");
  await expect(page).toHaveURL(/\/admin\/login$/);
  await page.getByLabel("E-mail").fill(ADMIN_EMAIL);
  await page.getByLabel("Senha").fill(ADMIN_PASSWORD);
  await page.getByRole("button", { name: "Entrar" }).click();
  await expect(page).toHaveURL(/\/admin$/);

  await page.getByLabel(/Buscar/).fill(email);
  const row = page.getByText("Pessoa E2E").locator("visible=true").first();
  await row.click();
  const dialog = page.getByRole("dialog");
  await expect(dialog.getByText("Mensagem para o painel")).toBeVisible();
  await dialog.getByRole("button", { name: "Excluir inscrição" }).click();
  await page.getByRole("alertdialog").getByRole("button", { name: "Excluir" }).click();
  await expect(page.getByText("Inscrição excluída.")).toBeVisible();

  await page.getByRole("button", { name: "Sair" }).click();
  await expect(page).toHaveURL(/\/admin\/login$/);
  await page.goto("/admin");
  await expect(page).toHaveURL(/\/admin\/login$/);
});

test("API recusa acesso administrativo sem sessão", async ({ request }) => {
  expect((await request.get("/api/v1/admin/registrations")).status()).toBe(401);
  expect((await request.get("/api/v1/health")).status()).toBe(200);
});
