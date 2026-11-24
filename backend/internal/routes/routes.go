// Package routes monta o roteador Gin com todos os middlewares e endpoints.
package routes

import (
	"context"
	"net/http"

	"github.com/chadecasanova/backend/internal/apperr"
	"github.com/chadecasanova/backend/internal/config"
	"github.com/chadecasanova/backend/internal/handlers"
	"github.com/chadecasanova/backend/internal/middleware"
	"github.com/chadecasanova/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type Deps struct {
	Config        *config.Config
	Registrations *services.RegistrationService
	Auth          *services.AuthService
	Ping          func(ctx context.Context) error
	LoginLimiter  *middleware.RateLimiter
	PublicLimiter *middleware.RateLimiter
}

func New(d Deps) (*gin.Engine, error) {
	cfg := d.Config
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.HandleMethodNotAllowed = true
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		return nil, err
	}

	r.Use(
		middleware.RequestID(),
		middleware.Logger(),
		middleware.Recovery(),
		middleware.SecurityHeaders(cfg.IsProduction()),
		middleware.CORS(cfg.AllowedOrigins),
		middleware.BodyLimit(16<<10),
		middleware.ErrorHandler(),
		middleware.SameOrigin(cfg.AllowedOrigins),
	)

	cookie := handlers.CookieSettings{Name: handlers.SessionCookieName(cfg.CookieSecure), Secure: cfg.CookieSecure}
	public := handlers.NewPublicHandler(d.Registrations, d.Ping)
	auth := handlers.NewAuthHandler(d.Auth, cookie)
	admin := handlers.NewAdminHandler(d.Registrations)
	requireAdmin := middleware.RequireAdmin(d.Auth, cookie.Name)
	requireCSRF := middleware.RequireCSRF(d.Auth, cfg.AllowedOrigins)

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", public.Health)
		v1.POST("/registrations", d.PublicLimiter.Middleware(), public.CreateRegistration)
		v1.POST("/messages", d.PublicLimiter.Middleware(), public.CreateMessage)
		v1.GET("/invites/:token", d.PublicLimiter.Middleware(), public.GetInvite)

		a := v1.Group("/auth")
		a.POST("/login", d.LoginLimiter.Middleware(), auth.Login)
		a.POST("/logout", requireAdmin, requireCSRF, auth.Logout)
		a.GET("/me", requireAdmin, auth.Me)

		adm := v1.Group("/admin", requireAdmin, requireCSRF)
		adm.GET("/stats", admin.Stats)
		adm.GET("/registrations", admin.List)
		adm.GET("/registrations/export", admin.Export)
		adm.GET("/registrations/:id", admin.Get)
		adm.PATCH("/registrations/:id", admin.Update)
		adm.DELETE("/registrations/:id", admin.Delete)
	}

	notFound := func(c *gin.Context) {
		e := apperr.NotFound("Endpoint não encontrado.")
		c.JSON(e.Status, gin.H{"error": e})
	}
	r.NoRoute(notFound)
	r.NoMethod(func(c *gin.Context) {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": gin.H{"code": "method_not_allowed", "message": "Método não permitido."}})
	})
	return r, nil
}
