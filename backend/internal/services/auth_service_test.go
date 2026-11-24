package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/chadecasanova/backend/internal/repositories"
	"github.com/chadecasanova/backend/internal/services"
	"github.com/chadecasanova/backend/internal/testutil"
)

func TestLoginAndSession(t *testing.T) {
	app := testutil.NewApp(t)
	ctx := context.Background()
	res, err := app.Auth.Login(ctx, " ADMIN@example.com ", testutil.AdminPassword)
	if err != nil {
		t.Fatal(err)
	}
	if !services.IsWellFormedToken(res.Token) || res.Admin.Email != testutil.AdminEmail {
		t.Fatalf("login inesperado: %+v", res)
	}
	admin, err := app.Auth.Authenticate(ctx, res.Token)
	if err != nil || admin.Email != testutil.AdminEmail {
		t.Fatalf("autenticação falhou: %v", err)
	}

	// O token não é guardado em texto puro.
	var stored string
	app.DB.Raw("SELECT token_hash FROM sessions LIMIT 1").Scan(&stored)
	if stored == res.Token || len(stored) != 64 {
		t.Fatal("token deveria estar armazenado como hash")
	}

	csrf := app.Auth.CSRFToken(res.Token)
	if !app.Auth.VerifyCSRF(res.Token, csrf) || app.Auth.VerifyCSRF(res.Token, csrf+"x") || app.Auth.VerifyCSRF(res.Token, "") {
		t.Fatal("verificação CSRF incorreta")
	}

	if err := app.Auth.Logout(ctx, res.Token); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Auth.Authenticate(ctx, res.Token); code(err) != "unauthorized" {
		t.Fatal("sessão deveria ter sido encerrada")
	}
}

func TestLoginDoesNotRevealWhichFieldIsWrong(t *testing.T) {
	app := testutil.NewApp(t)
	ctx := context.Background()
	_, errWrongPass := app.Auth.Login(ctx, testutil.AdminEmail, "senha-errada-123")
	_, errUnknown := app.Auth.Login(ctx, "ninguem@example.com", "senha-errada-123")
	if errWrongPass == nil || errUnknown == nil || errWrongPass.Error() != errUnknown.Error() {
		t.Fatalf("mensagens deveriam ser idênticas: %v / %v", errWrongPass, errUnknown)
	}
}

func TestExpiredSession(t *testing.T) {
	db := testutil.NewDB(t)
	repo := repositories.NewAuthRepository(db)
	auth, _ := services.NewAuthService(repo, []byte("0123456789abcdef0123456789abcdef"), time.Minute, testutil.FastArgon2)
	ctx := context.Background()
	if _, err := auth.CreateAdmin(ctx, "Admin", "a@example.com", "senha-muito-segura"); err != nil {
		t.Fatal(err)
	}
	res, err := auth.Login(ctx, "a@example.com", "senha-muito-segura")
	if err != nil {
		t.Fatal(err)
	}
	db.Exec("UPDATE sessions SET expires_at = ?", time.Now().UTC().Add(-time.Second))
	if _, err := auth.Authenticate(ctx, res.Token); code(err) != "unauthorized" {
		t.Fatal("sessão expirada deveria ser recusada")
	}
	if n, _ := auth.CleanupExpiredSessions(ctx); n != 1 {
		t.Fatalf("esperava remover 1 sessão, removeu %d", n)
	}
}

func TestCreateAdminRules(t *testing.T) {
	app := testutil.NewApp(t)
	ctx := context.Background()
	if _, err := app.Auth.CreateAdmin(ctx, "Outro", "novo@example.com", "curta"); code(err) != "validation_error" {
		t.Fatal("senha curta deveria falhar")
	}
	if _, err := app.Auth.CreateAdmin(ctx, "Outro", testutil.AdminEmail, "senha-muito-segura"); code(err) != "admin_exists" {
		t.Fatal("e-mail duplicado deveria falhar")
	}
	// Bootstrap não cria um segundo administrador.
	if err := app.Auth.EnsureBootstrapAdmin(ctx, "X", "x@example.com", "senha-muito-segura"); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Auth.Login(ctx, "x@example.com", "senha-muito-segura"); err == nil {
		t.Fatal("bootstrap não deveria rodar com administradores existentes")
	}
	// Troca de senha encerra sessões.
	res, _ := app.Auth.Login(ctx, testutil.AdminEmail, testutil.AdminPassword)
	if err := app.Auth.ResetPassword(ctx, testutil.AdminEmail, "outra-senha-segura-456"); err != nil {
		t.Fatal(err)
	}
	if _, err := app.Auth.Authenticate(ctx, res.Token); err == nil {
		t.Fatal("sessões antigas deveriam cair após troca de senha")
	}
}
