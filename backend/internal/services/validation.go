package services

import (
	"net/mail"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	NameMin    = 2
	NameMax    = 100
	EmailMax   = 254
	MessageMax = 1000
)

// NormalizeName remove espaços extras e rejeita caracteres de controle ou marcação.
func NormalizeName(raw string) (string, string) {
	name := strings.Join(strings.Fields(raw), " ")
	switch n := utf8.RuneCountInString(name); {
	case n == 0:
		return "", "Informe seu nome."
	case n < NameMin:
		return "", "O nome precisa ter pelo menos 2 letras."
	case n > NameMax:
		return "", "O nome pode ter no máximo 100 caracteres."
	}
	for _, r := range name {
		if unicode.IsControl(r) || r == '<' || r == '>' {
			return "", "O nome contém caracteres não permitidos."
		}
	}
	return name, ""
}

// NormalizeEmail aceita apenas um endereço simples (sem nome de exibição) e o devolve em minúsculas.
func NormalizeEmail(raw string) (string, string) {
	email := strings.ToLower(strings.TrimSpace(raw))
	if email == "" {
		return "", "Informe seu e-mail."
	}
	if len(email) > EmailMax {
		return "", "O e-mail é longo demais."
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email || addr.Name != "" {
		return "", "Confira o e-mail: ele parece incompleto."
	}
	at := strings.LastIndexByte(email, '@')
	domain := email[at+1:]
	if !strings.Contains(domain, ".") || strings.HasPrefix(domain, ".") || strings.HasSuffix(domain, ".") {
		return "", "Confira o e-mail: ele parece incompleto."
	}
	return email, ""
}

// NormalizeMessage padroniza quebras de linha e remove caracteres de controle (exceto \n e \t).
func NormalizeMessage(raw string) (string, string) {
	msg := strings.ReplaceAll(raw, "\r\n", "\n")
	msg = strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || !unicode.IsControl(r) {
			return r
		}
		return -1
	}, msg)
	msg = strings.TrimSpace(msg)
	switch n := utf8.RuneCountInString(msg); {
	case n == 0:
		return "", "Escreva sua mensagem antes de enviar."
	case n > MessageMax:
		return "", "A mensagem pode ter no máximo 1000 caracteres."
	}
	return msg, ""
}
