// Package middleware reúne os middlewares HTTP: request ID, logs, recuperação
// de pânico, tratamento central de erros, cabeçalhos de segurança, CORS,
// limite de corpo e rate limiting.
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"slices"
	"time"

	"github.com/chadecasanova/backend/internal/apperr"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const RequestIDKey = "request_id"

func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-ID")
		if _, err := uuid.Parse(id); err != nil {
			id = uuid.NewString()
		}
		c.Set(RequestIDKey, id)
		c.Header("X-Request-ID", id)
		c.Next()
	}
}

// Logger registra uma linha estruturada por requisição. Não registra query
// string, corpo nem cabeçalhos, para não vazar e-mails, senhas ou tokens.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		route := c.FullPath()
		if route == "" {
			route = "unmatched"
		}
		level := slog.LevelInfo
		if c.Writer.Status() >= 500 {
			level = slog.LevelError
		} else if c.Writer.Status() >= 400 {
			level = slog.LevelWarn
		}
		slog.Log(c.Request.Context(), level, "http_request",
			"request_id", c.GetString(RequestIDKey),
			"method", c.Request.Method,
			"route", route,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}

// Recovery captura pânicos: o stack trace vai apenas para o log do servidor.
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic", "request_id", c.GetString(RequestIDKey), "panic", r, "stack", string(debug.Stack()))
				e := apperr.Internal(nil)
				c.AbortWithStatusJSON(e.Status, gin.H{"error": e})
			}
		}()
		c.Next()
	}
}

// ErrorHandler converte o último erro registrado com c.Error numa resposta JSON.
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 || c.Writer.Written() {
			return
		}
		e := apperr.From(c.Errors.Last().Err)
		if e.Status >= 500 {
			slog.Error("erro interno", "request_id", c.GetString(RequestIDKey), "error", e.Error())
		}
		if e.Status == http.StatusTooManyRequests {
			c.Header("Retry-After", "60")
		}
		c.JSON(e.Status, gin.H{"error": e})
	}
}

func SecurityHeaders(production bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		h.Set("Cache-Control", "no-store")
		if production {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		c.Next()
	}
}

// CORS só libera origens da lista. Em produção o frontend é servido pelo mesmo
// domínio (Nginx), então a lista normalmente fica vazia.
func CORS(allowed []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if origin != "" && slices.Contains(allowed, origin) {
			h := c.Writer.Header()
			h.Set("Access-Control-Allow-Origin", origin)
			h.Set("Access-Control-Allow-Credentials", "true")
			h.Add("Vary", "Origin")
			if c.Request.Method == http.MethodOptions {
				h.Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE")
				h.Set("Access-Control-Allow-Headers", "Content-Type, X-CSRF-Token")
				h.Set("Access-Control-Max-Age", "600")
				c.AbortWithStatus(http.StatusNoContent)
				return
			}
		} else if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}

func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		c.Next()
	}
}
