// Package apperr define o erro de domínio usado em toda a API.
// Handlers registram o erro com c.Error(err) e o middleware ErrorHandler
// o converte numa resposta JSON padronizada, sem expor detalhes internos.
package apperr

import (
	"errors"
	"net/http"
)

type Error struct {
	Status  int               `json:"-"`
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
	Err     error             `json:"-"` // causa interna, apenas para logs
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Code + ": " + e.Err.Error()
	}
	return e.Code + ": " + e.Message
}

func (e *Error) Unwrap() error { return e.Err }

func Validation(fields map[string]string) *Error {
	return &Error{Status: http.StatusUnprocessableEntity, Code: "validation_error",
		Message: "Alguns campos precisam de atenção.", Fields: fields}
}

func BadRequest(code, msg string) *Error {
	return &Error{Status: http.StatusBadRequest, Code: code, Message: msg}
}

func NotFound(msg string) *Error {
	return &Error{Status: http.StatusNotFound, Code: "not_found", Message: msg}
}

func Conflict(code, msg string) *Error {
	return &Error{Status: http.StatusConflict, Code: code, Message: msg}
}

func Unauthorized(msg string) *Error {
	return &Error{Status: http.StatusUnauthorized, Code: "unauthorized", Message: msg}
}

func Forbidden(msg string) *Error {
	return &Error{Status: http.StatusForbidden, Code: "forbidden", Message: msg}
}

func TooManyRequests() *Error {
	return &Error{Status: http.StatusTooManyRequests, Code: "rate_limited",
		Message: "Muitas tentativas em pouco tempo. Aguarde um instante e tente de novo."}
}

func Internal(err error) *Error {
	return &Error{Status: http.StatusInternalServerError, Code: "internal_error",
		Message: "Não foi possível concluir agora. Tente novamente em alguns minutos.", Err: err}
}

// From converte qualquer erro num *Error; erros desconhecidos viram 500.
func From(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return Internal(err)
}
