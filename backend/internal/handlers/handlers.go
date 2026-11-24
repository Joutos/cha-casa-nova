// Package handlers traduz HTTP <-> serviços. Erros são registrados com c.Error
// e renderizados pelo middleware ErrorHandler.
package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/chadecasanova/backend/internal/apperr"
	"github.com/gin-gonic/gin"
)

// decodeJSON aceita um único objeto JSON e recusa campos desconhecidos.
func decodeJSON(c *gin.Context, dst any) error {
	dec := json.NewDecoder(c.Request.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return &apperr.Error{Status: http.StatusRequestEntityTooLarge, Code: "payload_too_large",
				Message: "O conteúdo enviado é grande demais."}
		}
		return apperr.BadRequest("invalid_json", "Não conseguimos ler os dados enviados.")
	}
	if dec.Decode(&struct{}{}) != io.EOF {
		return apperr.BadRequest("invalid_json", "Envie apenas um objeto JSON.")
	}
	return nil
}

func fail(c *gin.Context, err error) {
	c.Error(err)
	c.Abort()
}
