package middleware

import (
	"context"
	"net/http"
	"net/url"

	"github.com/chadecasanova/backend/internal/apperr"
	"github.com/chadecasanova/backend/internal/models"
	"github.com/gin-gonic/gin"
)

const (
	adminKey        = "admin"
	sessionTokenKey = "session_token"
	CSRFHeader      = "X-CSRF-Token"
)

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (*models.Admin, error)
	VerifyCSRF(sessionToken, provided string) bool
}

// RequireAdmin autentica pelo cookie de sessão e autoriza apenas administradores
// com sessão válida. Todas as rotas /admin e o logout passam por aqui.
func RequireAdmin(auth Authenticator, cookieName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie(cookieName)
		if err != nil || token == "" {
			c.Error(apperr.Unauthorized("Entre para acessar esta área."))
			c.Abort()
			return
		}
		admin, err := auth.Authenticate(c.Request.Context(), token)
		if err != nil {
			c.Error(err)
			c.Abort()
			return
		}
		c.Set(adminKey, admin)
		c.Set(sessionTokenKey, token)
		c.Next()
	}
}

// RequireCSRF exige, em métodos que alteram estado, o cabeçalho X-CSRF-Token
// derivado da sessão e, se houver, um Origin igual ao host da requisição.
func RequireCSRF(auth Authenticator, allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
			c.Next()
			return
		}
		if origin := c.GetHeader("Origin"); origin != "" && !originAllowed(origin, c.Request.Host, allowedOrigins) {
			c.Error(apperr.Forbidden("Origem da requisição não permitida."))
			c.Abort()
			return
		}
		token := c.GetString(sessionTokenKey)
		if !auth.VerifyCSRF(token, c.GetHeader(CSRFHeader)) {
			c.Error(apperr.Forbidden("Token de segurança inválido. Recarregue a página e tente de novo."))
			c.Abort()
			return
		}
		c.Next()
	}
}

func originAllowed(origin, host string, allowed []string) bool {
	for _, a := range allowed {
		if a == origin {
			return true
		}
	}
	u, err := url.Parse(origin)
	return err == nil && u.Host == host
}

func CurrentAdmin(c *gin.Context) *models.Admin {
	if v, ok := c.Get(adminKey); ok {
		return v.(*models.Admin)
	}
	return nil
}

func SessionToken(c *gin.Context) string { return c.GetString(sessionTokenKey) }

// SameOrigin rejeita requisições que alteram estado vindas de outra origem
// (quando o navegador envia Origin). Complementa SameSite=Strict e o token CSRF.
func SameOrigin(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions:
		default:
			if origin := c.GetHeader("Origin"); origin != "" && !originAllowed(origin, c.Request.Host, allowedOrigins) {
				c.Error(apperr.Forbidden("Origem da requisição não permitida."))
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
