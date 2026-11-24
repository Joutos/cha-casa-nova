package routes_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/chadecasanova/backend/internal/config"
	"github.com/chadecasanova/backend/internal/handlers"
	"github.com/chadecasanova/backend/internal/testutil"
)

type R = testutil.Request

func decode(t *testing.T, body string) map[string]any {
	t.Helper()
	var m map[string]any
	if err := json.Unmarshal([]byte(body), &m); err != nil {
		t.Fatalf("JSON inválido: %s", body)
	}
	return m
}

func errCode(t *testing.T, body string) string {
	e, _ := decode(t, body)["error"].(map[string]any)
	s, _ := e["code"].(string)
	return s
}

type session struct {
	cookie *http.Cookie
	csrf   string
}

func login(t *testing.T, app *testutil.App) session {
	t.Helper()
	w := app.Do(R{Method: "POST", Path: "/api/v1/auth/login",
		Body: `{"email":"` + testutil.AdminEmail + `","password":"` + testutil.AdminPassword + `"}`})
	if w.Code != http.StatusOK {
		t.Fatalf("login: %d %s", w.Code, w.Body)
	}
	var cookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == handlers.SessionCookieName(false) {
			cookie = c
		}
	}
	if cookie == nil || !cookie.HttpOnly || cookie.SameSite != http.SameSiteStrictMode || cookie.Path != "/" {
		t.Fatalf("cookie de sessão inesperado: %+v", cookie)
	}
	if strings.Contains(w.Body.String(), "password") || strings.Contains(w.Body.String(), cookie.Value) {
		t.Fatal("resposta de login não deve conter senha nem o token da sessão")
	}
	return session{cookie: cookie, csrf: decode(t, w.Body.String())["csrf_token"].(string)}
}

func (s session) req(method, path, body string) R {
	return R{Method: method, Path: path, Body: body, Cookies: []*http.Cookie{s.cookie},
		Headers: map[string]string{"X-CSRF-Token": s.csrf}}
}

func TestHealthAndSecurityHeaders(t *testing.T) {
	app := testutil.NewApp(t)
	w := app.Do(R{Method: "GET", Path: "/api/v1/health"})
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"ok"`) {
		t.Fatalf("health: %d %s", w.Code, w.Body)
	}
	for _, h := range []string{"X-Content-Type-Options", "X-Frame-Options", "Content-Security-Policy", "Referrer-Policy", "X-Request-ID"} {
		if w.Header().Get(h) == "" {
			t.Errorf("cabeçalho ausente: %s", h)
		}
	}
	if w := app.Do(R{Method: "GET", Path: "/api/v1/nada"}); w.Code != 404 || errCode(t, w.Body.String()) != "not_found" {
		t.Fatalf("404 esperado: %d", w.Code)
	}
}

func TestCreateRegistration(t *testing.T) {
	app := testutil.NewApp(t)
	body := `{"name":"Ana Lima","email":"ana@example.com","contribute":true,"website":""}`
	if w := app.Do(R{Method: "POST", Path: "/api/v1/registrations", Body: body}); w.Code != http.StatusCreated {
		t.Fatalf("criação: %d %s", w.Code, w.Body)
	}
	w := app.Do(R{Method: "POST", Path: "/api/v1/registrations", Body: body})
	if w.Code != http.StatusConflict || errCode(t, w.Body.String()) != "already_confirmed" {
		t.Fatalf("duplicado: %d %s", w.Code, w.Body)
	}
}

func TestCreateRegistrationValidation(t *testing.T) {
	app := testutil.NewApp(t)
	w := app.Do(R{Method: "POST", Path: "/api/v1/registrations", Body: `{"name":"A","email":"nao-e-email"}`})
	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("esperava 422: %d", w.Code)
	}
	fields := decode(t, w.Body.String())["error"].(map[string]any)["fields"].(map[string]any)
	if fields["name"] == nil || fields["email"] == nil {
		t.Fatalf("campos de erro ausentes: %v", fields)
	}
	for _, bad := range []string{`{"name":"Ana","email":"a@b.com","admin":true}`, `not json`, `{"name":"Ana"}{"x":1}`} {
		if w := app.Do(R{Method: "POST", Path: "/api/v1/registrations", Body: bad}); w.Code != http.StatusBadRequest {
			t.Errorf("corpo %q deveria dar 400, deu %d", bad, w.Code)
		}
	}
	big := `{"name":"Ana","email":"a@b.com","website":"` + strings.Repeat("x", 20000) + `"}`
	if w := app.Do(R{Method: "POST", Path: "/api/v1/registrations", Body: big}); w.Code != http.StatusRequestEntityTooLarge {
		t.Errorf("corpo grande deveria dar 413, deu %d", w.Code)
	}
}

func TestHoneypotLooksSuccessfulButStoresNothing(t *testing.T) {
	app := testutil.NewApp(t)
	w := app.Do(R{Method: "POST", Path: "/api/v1/registrations", Body: `{"name":"Bot","email":"bot@spam.com","website":"http://x"}`})
	if w.Code != http.StatusCreated {
		t.Fatalf("honeypot deveria parecer sucesso: %d", w.Code)
	}
	var n int64
	app.DB.Raw("SELECT COUNT(*) FROM guests").Scan(&n)
	if n != 0 {
		t.Fatal("honeypot não deveria gravar")
	}
}

func TestCreateMessage(t *testing.T) {
	app := testutil.NewApp(t)
	w := app.Do(R{Method: "POST", Path: "/api/v1/messages", Body: `{"name":"Bia","email":"bia@example.com","message":"<b>Parabéns!</b>"}`})
	if w.Code != http.StatusCreated {
		t.Fatalf("mensagem: %d %s", w.Code, w.Body)
	}
	if w := app.Do(R{Method: "POST", Path: "/api/v1/messages", Body: `{"name":"Bia","email":"bia@example.com","message":"   "}`}); w.Code != 422 {
		t.Fatalf("mensagem vazia deveria dar 422: %d", w.Code)
	}
}

func TestPublicRateLimit(t *testing.T) {
	app := testutil.NewApp(t, func(c *config.Config) { c.PublicRatePerMinute = 2 })
	var last int
	for i := 0; i < 3; i++ {
		w := app.Do(R{Method: "POST", Path: "/api/v1/registrations", Body: `{"name":"Ana","email":"x` + string(rune('a'+i)) + `@example.com"}`})
		last = w.Code
		if i == 2 && w.Header().Get("Retry-After") == "" {
			t.Error("Retry-After ausente")
		}
	}
	if last != http.StatusTooManyRequests {
		t.Fatalf("terceira requisição deveria ser limitada: %d", last)
	}
}

func TestLoginRateLimitAndGenericError(t *testing.T) {
	app := testutil.NewApp(t, func(c *config.Config) { c.LoginRatePerMinute = 3 })
	codes := []int{}
	var bodies []string
	for _, email := range []string{testutil.AdminEmail, "ninguem@example.com", testutil.AdminEmail, testutil.AdminEmail} {
		w := app.Do(R{Method: "POST", Path: "/api/v1/auth/login", Body: `{"email":"` + email + `","password":"errada-errada"}`})
		codes = append(codes, w.Code)
		bodies = append(bodies, w.Body.String())
	}
	if codes[0] != 401 || codes[1] != 401 || codes[3] != 429 {
		t.Fatalf("códigos inesperados: %v", codes)
	}
	if bodies[0] != bodies[1] {
		t.Fatal("erro de login deve ser idêntico para e-mail existente e inexistente")
	}
}

func TestAdminRequiresAuthentication(t *testing.T) {
	app := testutil.NewApp(t)
	for _, r := range []R{
		{Method: "GET", Path: "/api/v1/admin/registrations"},
		{Method: "GET", Path: "/api/v1/admin/registrations/export"},
		{Method: "GET", Path: "/api/v1/auth/me"},
		{Method: "DELETE", Path: "/api/v1/admin/registrations/00000000-0000-0000-0000-000000000000"},
		{Method: "GET", Path: "/api/v1/admin/registrations", Cookies: []*http.Cookie{{Name: "ccn_session", Value: strings.Repeat("a", 43)}}},
	} {
		if w := app.Do(r); w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sem sessão: %d", r.Method, r.Path, w.Code)
		}
	}
}

func TestAdminFlow(t *testing.T) {
	app := testutil.NewApp(t)
	app.Do(R{Method: "POST", Path: "/api/v1/registrations", Body: `{"name":"Carla","email":"carla@example.com"}`})
	app.Do(R{Method: "POST", Path: "/api/v1/messages", Body: `{"name":"=cmd|calc","email":"evil@example.com","message":"Oi"}`})
	s := login(t, app)

	w := app.Do(s.req("GET", "/api/v1/auth/me", ""))
	if w.Code != 200 || decode(t, w.Body.String())["csrf_token"] != s.csrf {
		t.Fatalf("me: %d %s", w.Code, w.Body)
	}

	w = app.Do(s.req("GET", "/api/v1/admin/registrations?q=carla&sort=name&order=asc&page=1&page_size=10", ""))
	if w.Code != 200 {
		t.Fatalf("lista: %d %s", w.Code, w.Body)
	}
	list := decode(t, w.Body.String())
	items := list["items"].([]any)
	if list["total"].(float64) != 1 || len(items) != 1 {
		t.Fatalf("busca: %v", list)
	}
	id := items[0].(map[string]any)["id"].(string)
	if strings.Contains(w.Body.String(), "invite_token") {
		t.Fatal("listagem não deve expor o token do convite")
	}

	if w := app.Do(s.req("GET", "/api/v1/admin/registrations?sort=password_hash", "")); w.Code != 422 {
		t.Fatalf("ordenação inválida deveria dar 422: %d", w.Code)
	}

	w = app.Do(s.req("GET", "/api/v1/admin/registrations/"+id, ""))
	if w.Code != 200 || !strings.Contains(w.Body.String(), `"invite_url":"https://festa.example/convite/`) {
		t.Fatalf("detalhe: %d %s", w.Code, w.Body)
	}

	// PATCH sem CSRF é recusado; com CSRF passa.
	noCSRF := s.req("PATCH", "/api/v1/admin/registrations/"+id, `{"wants_to_contribute":true}`)
	delete(noCSRF.Headers, "X-CSRF-Token")
	if w := app.Do(noCSRF); w.Code != http.StatusForbidden {
		t.Fatalf("PATCH sem CSRF: %d", w.Code)
	}
	crossOrigin := s.req("PATCH", "/api/v1/admin/registrations/"+id, `{"wants_to_contribute":true}`)
	crossOrigin.Headers["Origin"] = "https://malicioso.example"
	if w := app.Do(crossOrigin); w.Code != http.StatusForbidden {
		t.Fatalf("PATCH de outra origem: %d", w.Code)
	}
	w = app.Do(s.req("PATCH", "/api/v1/admin/registrations/"+id, `{"wants_to_contribute":true}`))
	if w.Code != 200 || !decode(t, w.Body.String())["wants_to_contribute"].(bool) {
		t.Fatalf("PATCH: %d %s", w.Code, w.Body)
	}

	w = app.Do(s.req("GET", "/api/v1/admin/stats", ""))
	if st := decode(t, w.Body.String()); st["total"].(float64) != 2 || st["contribute"].(float64) != 1 {
		t.Fatalf("stats: %v", st)
	}

	// Export: CSV com BOM, separador ';' e proteção contra fórmulas.
	w = app.Do(s.req("GET", "/api/v1/admin/registrations/export?sort=name&order=asc", ""))
	csv := w.Body.String()
	if w.Code != 200 || !strings.HasPrefix(csv, "\ufeffNome;E-mail;") || !strings.Contains(w.Header().Get("Content-Disposition"), "attachment") {
		t.Fatalf("export: %d %q", w.Code, csv)
	}
	if !strings.Contains(csv, "'=cmd|calc") || strings.Contains(csv, ";=cmd") {
		t.Fatalf("célula perigosa não neutralizada: %q", csv)
	}

	if w := app.Do(s.req("DELETE", "/api/v1/admin/registrations/"+id, "")); w.Code != http.StatusNoContent {
		t.Fatalf("DELETE: %d", w.Code)
	}
	if w := app.Do(s.req("GET", "/api/v1/admin/registrations/"+id, "")); w.Code != 404 {
		t.Fatalf("após DELETE: %d", w.Code)
	}

	// Logout invalida a sessão no servidor.
	if w := app.Do(s.req("POST", "/api/v1/auth/logout", "")); w.Code != http.StatusNoContent {
		t.Fatalf("logout: %d", w.Code)
	}
	if w := app.Do(s.req("GET", "/api/v1/auth/me", "")); w.Code != 401 {
		t.Fatalf("sessão deveria estar encerrada: %d", w.Code)
	}
}

func TestInviteEndpoint(t *testing.T) {
	app := testutil.NewApp(t)
	app.Do(R{Method: "POST", Path: "/api/v1/registrations", Body: `{"name":"Eva","email":"eva@example.com"}`})
	var token string
	app.DB.Raw("SELECT invite_token FROM guests LIMIT 1").Scan(&token)
	w := app.Do(R{Method: "GET", Path: "/api/v1/invites/" + token})
	if w.Code != 200 || decode(t, w.Body.String())["guest_name"] != "Eva" {
		t.Fatalf("convite: %d %s", w.Code, w.Body)
	}
	if w := app.Do(R{Method: "GET", Path: "/api/v1/invites/" + strings.Repeat("b", 43)}); w.Code != 404 {
		t.Fatalf("token desconhecido: %d", w.Code)
	}
}

func TestCORS(t *testing.T) {
	app := testutil.NewApp(t, func(c *config.Config) { c.AllowedOrigins = []string{"http://localhost:5173"} })
	w := app.Do(R{Method: "OPTIONS", Path: "/api/v1/registrations", Headers: map[string]string{"Origin": "http://localhost:5173"}})
	if w.Code != 204 || w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("preflight permitido: %d", w.Code)
	}
	w = app.Do(R{Method: "OPTIONS", Path: "/api/v1/registrations", Headers: map[string]string{"Origin": "https://outro.example"}})
	if w.Code != 403 || w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("preflight negado: %d", w.Code)
	}
}
