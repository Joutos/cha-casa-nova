package handlers

import (
	"net/http"
	"time"

	"github.com/chadecasanova/backend/internal/middleware"
	"github.com/chadecasanova/backend/internal/models"
	"github.com/chadecasanova/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type CookieSettings struct {
	Name   string
	Secure bool
}

// SessionCookieName usa o prefixo __Host- quando o cookie é Secure: o navegador
// passa a exigir HTTPS, Path=/ e ausência de Domain.
func SessionCookieName(secure bool) string {
	if secure {
		return "__Host-ccn_session"
	}
	return "ccn_session"
}

type AuthHandler struct {
	svc    *services.AuthService
	cookie CookieSettings
}

func NewAuthHandler(svc *services.AuthService, cookie CookieSettings) *AuthHandler {
	return &AuthHandler{svc: svc, cookie: cookie}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type adminView struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func view(a *models.Admin) adminView { return adminView{ID: a.ID, Name: a.Name, Email: a.Email} }

func (h *AuthHandler) setCookie(c *gin.Context, value string, maxAge int) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name: h.cookie.Name, Value: value, Path: "/", MaxAge: maxAge,
		HttpOnly: true, Secure: h.cookie.Secure, SameSite: http.SameSiteStrictMode,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := decodeJSON(c, &req); err != nil {
		fail(c, err)
		return
	}
	res, err := h.svc.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		fail(c, err)
		return
	}
	h.setCookie(c, res.Token, int(time.Until(res.ExpiresAt).Seconds()))
	c.JSON(http.StatusOK, gin.H{
		"admin":      view(res.Admin),
		"csrf_token": h.svc.CSRFToken(res.Token),
		"expires_at": res.ExpiresAt,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	if err := h.svc.Logout(c.Request.Context(), middleware.SessionToken(c)); err != nil {
		fail(c, err)
		return
	}
	h.setCookie(c, "", -1)
	c.Status(http.StatusNoContent)
}

func (h *AuthHandler) Me(c *gin.Context) {
	admin := middleware.CurrentAdmin(c)
	c.JSON(http.StatusOK, gin.H{
		"admin":      view(admin),
		"csrf_token": h.svc.CSRFToken(middleware.SessionToken(c)),
	})
}
