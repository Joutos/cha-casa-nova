// Comando principal.
//
//	server                 inicia a API (padrão)
//	server migrate         apenas aplica as migrations
//	server create-admin    cria um administrador (ADMIN_NAME, ADMIN_EMAIL, ADMIN_PASSWORD)
//	server reset-password  troca a senha de um administrador (ADMIN_EMAIL, ADMIN_PASSWORD)
//	server healthcheck     verifica /api/v1/health (usado pelo Docker)
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chadecasanova/backend/internal/apperr"
	"github.com/chadecasanova/backend/internal/config"
	"github.com/chadecasanova/backend/internal/database"
	"github.com/chadecasanova/backend/internal/middleware"
	"github.com/chadecasanova/backend/internal/repositories"
	"github.com/chadecasanova/backend/internal/routes"
	"github.com/chadecasanova/backend/internal/services"
	"github.com/chadecasanova/backend/migrations"
	"gorm.io/gorm"
)

func main() {
	// Usado pelo HEALTHCHECK do Docker (a imagem final não tem curl).
	if len(os.Args) > 1 && os.Args[1] == "healthcheck" {
		os.Exit(healthcheck())
	}
	if err := run(); err != nil {
		slog.Error("encerrando com erro", "error", describe(err))
		os.Exit(1)
	}
}

func healthcheck() int {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	client := &http.Client{Timeout: 3 * time.Second}
	res, err := client.Get("http://127.0.0.1:" + port + "/api/v1/health")
	if err != nil {
		return 1
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}

func describe(err error) string {
	var ae *apperr.Error
	if errors.As(err, &ae) && len(ae.Fields) > 0 {
		return fmt.Sprintf("%s %v", ae.Message, ae.Fields)
	}
	return err.Error()
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel})))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Open(cfg.DatabasePath)
	if err != nil {
		return err
	}
	if err := database.Migrate(ctx, db, migrations.FS); err != nil {
		return err
	}

	authSvc, err := services.NewAuthService(repositories.NewAuthRepository(db), cfg.SessionSecret, cfg.SessionTTL, services.DefaultArgon2)
	if err != nil {
		return err
	}

	cmd := "serve"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	switch cmd {
	case "migrate":
		slog.Info("migrations em dia")
		return nil
	case "create-admin":
		a, err := authSvc.CreateAdmin(ctx, cfg.BootstrapAdminName, cfg.BootstrapAdminEmail, cfg.BootstrapAdminPassword)
		if err != nil {
			return err
		}
		slog.Info("administrador criado", "admin_id", a.ID)
		return nil
	case "reset-password":
		if err := authSvc.ResetPassword(ctx, cfg.BootstrapAdminEmail, cfg.BootstrapAdminPassword); err != nil {
			return err
		}
		slog.Info("senha atualizada e sessões encerradas")
		return nil
	case "serve":
		return serve(ctx, cfg, db, authSvc)
	default:
		return fmt.Errorf("comando desconhecido: %s", cmd)
	}
}

func serve(ctx context.Context, cfg *config.Config, db *gorm.DB, authSvc *services.AuthService) error {
	if err := authSvc.EnsureBootstrapAdmin(ctx, cfg.BootstrapAdminName, cfg.BootstrapAdminEmail, cfg.BootstrapAdminPassword); err != nil {
		return fmt.Errorf("bootstrap do administrador: %s", describe(err))
	}

	mailer := services.NewMailer(cfg.SMTP)
	if !mailer.Enabled() {
		slog.Warn("SMTP não configurado: convites não serão enviados por e-mail (o link aparece no painel)")
	}
	regSvc := services.NewRegistrationService(repositories.NewGuestRepository(db), mailer, cfg.Event, cfg.PublicBaseURL, cfg.MaxMessagesPerGuest)

	loginLimiter := middleware.NewRateLimiter(cfg.LoginRatePerMinute)
	publicLimiter := middleware.NewRateLimiter(cfg.PublicRatePerMinute)

	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	router, err := routes.New(routes.Deps{
		Config: cfg, Registrations: regSvc, Auth: authSvc, Ping: sqlDB.PingContext,
		LoginLimiter: loginLimiter, PublicLimiter: publicLimiter,
	})
	if err != nil {
		return err
	}

	go func() {
		t := time.NewTicker(10 * time.Minute)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				loginLimiter.Cleanup()
				publicLimiter.Cleanup()
				if n, err := authSvc.CleanupExpiredSessions(ctx); err != nil {
					slog.Warn("limpeza de sessões falhou", "error", err)
				} else if n > 0 {
					slog.Info("sessões expiradas removidas", "count", n)
				}
			}
		}
	}()

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    16 << 10,
	}
	errCh := make(chan error, 1)
	go func() {
		slog.Info("API ouvindo", "port", cfg.Port, "env", cfg.Env)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errCh:
		return err
	}
	slog.Info("desligando")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	return sqlDB.Close()
}
