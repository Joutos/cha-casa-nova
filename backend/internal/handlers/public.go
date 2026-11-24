package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/chadecasanova/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type PublicHandler struct {
	svc  *services.RegistrationService
	ping func(ctx context.Context) error
}

func NewPublicHandler(svc *services.RegistrationService, ping func(ctx context.Context) error) *PublicHandler {
	return &PublicHandler{svc: svc, ping: ping}
}

func (h *PublicHandler) Health(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
	defer cancel()
	if err := h.ping(ctx); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unavailable"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

type registrationRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Contribute bool   `json:"contribute"`
	Website    string `json:"website"` // honeypot: fica oculto no formulário
}

const confirmedMessage = "Presença confirmada! Fique de olho na sua caixa de entrada."

// CreateRegistration confirma a presença de um convidado.
func (h *PublicHandler) CreateRegistration(c *gin.Context) {
	var req registrationRequest
	if err := decodeJSON(c, &req); err != nil {
		fail(c, err)
		return
	}
	_, err := h.svc.ConfirmAttendance(c.Request.Context(), services.AttendanceInput{
		Name: req.Name, Email: req.Email, Contribute: req.Contribute, Honeypot: req.Website,
	})
	if err != nil && !errors.Is(err, services.ErrSpam) {
		fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": confirmedMessage})
}

type messageRequest struct {
	Name       string `json:"name"`
	Email      string `json:"email"`
	Message    string `json:"message"`
	Contribute bool   `json:"contribute"`
	Website    string `json:"website"`
}

func (h *PublicHandler) CreateMessage(c *gin.Context) {
	var req messageRequest
	if err := decodeJSON(c, &req); err != nil {
		fail(c, err)
		return
	}
	_, err := h.svc.LeaveMessage(c.Request.Context(), services.MessageInput{
		Name: req.Name, Email: req.Email, Body: req.Message, Contribute: req.Contribute, Honeypot: req.Website,
	})
	if err != nil && !errors.Is(err, services.ErrSpam) {
		fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Mensagem recebida. Obrigado pelo carinho!"})
}

func (h *PublicHandler) GetInvite(c *gin.Context) {
	inv, err := h.svc.GetInvite(c.Request.Context(), c.Param("token"))
	if err != nil {
		fail(c, err)
		return
	}
	c.Header("X-Robots-Tag", "noindex")
	c.JSON(http.StatusOK, inv)
}
