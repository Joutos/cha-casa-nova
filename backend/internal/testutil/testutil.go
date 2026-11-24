// Package testutil monta o app completo sobre um SQLite em memória.
package testutil

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chadecasanova/backend/internal/config"
	"github.com/chadecasanova/backend/internal/database"
	"github.com/chadecasanova/backend/internal/middleware"
	"github.com/chadecasanova/backend/internal/repositories"
	"github.com/chadecasanova/backend/internal/routes"
	"github.com/chadecasanova/backend/internal/services"
	"github.com/chadecasanova/backend/migrations"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// FastArgon2 deixa os testes rápidos; produção usa services.DefaultArgon2.
var FastArgon2 = services.Argon2Params{Memory: 1024, Time: 1, Threads: 1, SaltLen: 16, KeyLen: 32}

const (
	AdminEmail    = "admin@example.com"
	AdminPassword = "senha-muito-segura-123"
)

type FakeMailer struct {
	mu   sync.Mutex
	Sent []services.InviteEmail
}

func (f *FakeMailer) Enabled() bool { return true }
func (f *FakeMailer) SendInvite(_ context.Context, e services.InviteEmail) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.Sent = append(f.Sent, e)
	return nil
}
func (f *FakeMailer) Count() int { f.mu.Lock(); defer f.mu.Unlock(); return len(f.Sent) }

type App struct {
	DB            *gorm.DB
	Config        *config.Config
	Router        *gin.Engine
	Auth          *services.AuthService
	Registrations *services.RegistrationService
	Mailer        *FakeMailer
}

type Option func(*config.Config)

func NewDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(context.Background(), db, migrations.FS); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if s, err := db.DB(); err == nil {
			s.Close()
		}
	})
	return db
}

func NewApp(t *testing.T, opts ...Option) *App {
	t.Helper()
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{
		Env: "test", SessionTTL: time.Hour, SessionSecret: []byte(strings.Repeat("s", 32)),
		TrustedProxies: []string{}, LoginRatePerMinute: 100, PublicRatePerMinute: 100,
		MaxMessagesPerGuest: 3, PublicBaseURL: "https://festa.example",
		Event: config.Event{DateLabel: "23 de Janeiro de 2027", Time: "11h00", AddressLine1: "Rua Teste, 1",
			AddressLine2: "Bairro, Cidade", MapsQuery: "Rua Teste 1", Signature: "Sofia e João"},
	}
	for _, o := range opts {
		o(cfg)
	}
	db := NewDB(t)
	auth, err := services.NewAuthService(repositories.NewAuthRepository(db), cfg.SessionSecret, cfg.SessionTTL, FastArgon2)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := auth.CreateAdmin(context.Background(), "Sofia", AdminEmail, AdminPassword); err != nil {
		t.Fatal(err)
	}
	mailer := &FakeMailer{}
	regs := services.NewRegistrationService(repositories.NewGuestRepository(db), mailer, cfg.Event, cfg.PublicBaseURL, cfg.MaxMessagesPerGuest)
	sqlDB, _ := db.DB()
	r, err := routes.New(routes.Deps{
		Config: cfg, Registrations: regs, Auth: auth, Ping: sqlDB.PingContext,
		LoginLimiter: middleware.NewRateLimiter(cfg.LoginRatePerMinute), PublicLimiter: middleware.NewRateLimiter(cfg.PublicRatePerMinute),
	})
	if err != nil {
		t.Fatal(err)
	}
	return &App{DB: db, Config: cfg, Router: r, Auth: auth, Registrations: regs, Mailer: mailer}
}

type Request struct {
	Method  string
	Path    string
	Body    string
	Headers map[string]string
	Cookies []*http.Cookie
}

func (a *App) Do(req Request) *httptest.ResponseRecorder {
	var body *strings.Reader
	if req.Body != "" {
		body = strings.NewReader(req.Body)
	} else {
		body = strings.NewReader("")
	}
	r := httptest.NewRequest(req.Method, req.Path, body)
	r.RemoteAddr = "203.0.113.10:1234"
	if req.Body != "" {
		r.Header.Set("Content-Type", "application/json")
	}
	for k, v := range req.Headers {
		r.Header.Set(k, v)
	}
	for _, c := range req.Cookies {
		r.AddCookie(c)
	}
	w := httptest.NewRecorder()
	a.Router.ServeHTTP(w, r)
	return w
}
