package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chadecasanova/backend/internal/services"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct{ svc *services.RegistrationService }

func NewAdminHandler(svc *services.RegistrationService) *AdminHandler { return &AdminHandler{svc: svc} }

func atoi(s string) int { n, _ := strconv.Atoi(s); return n }

func (h *AdminHandler) params(c *gin.Context) (p services.ListQuery) {
	return services.ListQuery{
		Q: c.Query("q"), Type: c.Query("type"), Sort: c.Query("sort"), Order: c.Query("order"),
		Page: atoi(c.Query("page")), PageSize: atoi(c.Query("page_size")),
	}
}

func (h *AdminHandler) List(c *gin.Context) {
	p, err := h.params(c).Normalize()
	if err != nil {
		fail(c, err)
		return
	}
	res, err := h.svc.List(c.Request.Context(), p)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *AdminHandler) Stats(c *gin.Context) {
	st, err := h.svc.Stats(c.Request.Context())
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, st)
}

func (h *AdminHandler) Get(c *gin.Context) {
	d, err := h.svc.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *AdminHandler) Update(c *gin.Context) {
	var patch services.GuestPatch
	if err := decodeJSON(c, &patch); err != nil {
		fail(c, err)
		return
	}
	d, err := h.svc.Update(c.Request.Context(), c.Param("id"), patch)
	if err != nil {
		fail(c, err)
		return
	}
	c.JSON(http.StatusOK, d)
}

func (h *AdminHandler) Delete(c *gin.Context) {
	if err := h.svc.Delete(c.Request.Context(), c.Param("id")); err != nil {
		fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) Export(c *gin.Context) {
	p, err := h.params(c).Normalize()
	if err != nil {
		fail(c, err)
		return
	}
	filename := fmt.Sprintf("convidados-%s.csv", time.Now().Format("2006-01-02"))
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", `attachment; filename="`+filename+`"`)
	c.Status(http.StatusOK)
	if err := h.svc.ExportCSV(c.Request.Context(), p, c.Writer); err != nil {
		fail(c, err)
	}
}
