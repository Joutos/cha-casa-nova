package services

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"html/template"
	"mime"
	"net"
	"net/smtp"
	"net/url"
	"strconv"
	"strings"
	textTemplate "text/template"
	"time"

	"github.com/chadecasanova/backend/internal/config"
)

type InviteEmail struct {
	To        string
	Invite    Invite
	InviteURL string
}

// Mailer envia o convite (Frame 6 do Figma). Sem SMTP configurado, usa NoopMailer.
type Mailer interface {
	Enabled() bool
	SendInvite(ctx context.Context, e InviteEmail) error
}

type NoopMailer struct{}

func (NoopMailer) Enabled() bool                                 { return false }
func (NoopMailer) SendInvite(context.Context, InviteEmail) error { return nil }

type SMTPMailer struct{ cfg config.SMTP }

func NewMailer(cfg config.SMTP) Mailer {
	if !cfg.Enabled() {
		return NoopMailer{}
	}
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) Enabled() bool { return true }

func (m *SMTPMailer) SendInvite(ctx context.Context, e InviteEmail) error {
	msg, err := BuildInviteMessage(m.cfg.From, e)
	if err != nil {
		return err
	}
	addr := net.JoinHostPort(m.cfg.Host, strconv.Itoa(m.cfg.Port))
	dialer := &net.Dialer{Timeout: 15 * time.Second}

	var conn net.Conn
	if m.cfg.Port == 465 { // TLS implícito
		conn, err = tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{ServerName: m.cfg.Host, MinVersion: tls.VersionTLS12})
	} else {
		conn, err = dialer.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return err
	}
	if dl, ok := ctx.Deadline(); ok {
		_ = conn.SetDeadline(dl)
	}
	c, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		conn.Close()
		return err
	}
	defer c.Close()

	if ok, _ := c.Extension("STARTTLS"); ok && m.cfg.Port != 465 {
		if err := c.StartTLS(&tls.Config{ServerName: m.cfg.Host, MinVersion: tls.VersionTLS12}); err != nil {
			return err
		}
	}
	if m.cfg.Username != "" {
		if err := c.Auth(smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)); err != nil {
			return err
		}
	}
	from := m.cfg.From
	if a, err := parseAddr(from); err == nil {
		from = a
	}
	if err := c.Mail(from); err != nil {
		return err
	}
	if err := c.Rcpt(e.To); err != nil {
		return err
	}
	wc, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := wc.Write(msg); err != nil {
		return err
	}
	if err := wc.Close(); err != nil {
		return err
	}
	return c.Quit()
}

func parseAddr(s string) (string, error) {
	if i := strings.LastIndex(s, "<"); i >= 0 && strings.HasSuffix(s, ">") {
		return s[i+1 : len(s)-1], nil
	}
	return s, nil
}

// MapsURL monta o link do Google Maps para o endereço configurado.
func MapsURL(query string) string {
	if query == "" {
		return ""
	}
	return "https://www.google.com/maps/search/?api=1&query=" + url.QueryEscape(query)
}

type inviteView struct {
	Invite
	InviteURL string
	MapsURL   string
}

// BuildInviteMessage gera a mensagem MIME multipart (texto + HTML).
func BuildInviteMessage(from string, e InviteEmail) ([]byte, error) {
	view := inviteView{Invite: e.Invite, InviteURL: e.InviteURL, MapsURL: MapsURL(e.Invite.MapsQuery)}
	var html, text bytes.Buffer
	if err := inviteHTML.Execute(&html, view); err != nil {
		return nil, err
	}
	if err := inviteText.Execute(&text, view); err != nil {
		return nil, err
	}
	b := make([]byte, 12)
	_, _ = rand.Read(b)
	boundary := "ccn-" + hex.EncodeToString(b)

	var buf bytes.Buffer
	h := func(k, v string) { fmt.Fprintf(&buf, "%s: %s\r\n", k, v) }
	h("From", from)
	h("To", e.To)
	h("Subject", mime.QEncoding.Encode("utf-8", "Seu convite para o nosso Chá de Casa Nova ♡"))
	h("Date", time.Now().Format(time.RFC1123Z))
	h("MIME-Version", "1.0")
	h("Content-Type", `multipart/alternative; boundary="`+boundary+`"`)
	buf.WriteString("\r\n")
	for _, part := range []struct{ ct, body string }{
		{"text/plain; charset=utf-8", text.String()},
		{"text/html; charset=utf-8", html.String()},
	} {
		fmt.Fprintf(&buf, "--%s\r\nContent-Type: %s\r\nContent-Transfer-Encoding: 8bit\r\n\r\n", boundary, part.ct)
		buf.WriteString(strings.ReplaceAll(part.body, "\n", "\r\n"))
		buf.WriteString("\r\n")
	}
	fmt.Fprintf(&buf, "--%s--\r\n", boundary)
	return buf.Bytes(), nil
}

var inviteText = textTemplate.Must(textTemplate.New("text").Parse(`Que bom que você vem, {{.GuestName}}! ♡

Estamos muito felizes em poder compartilhar esse momento com você!
Nossa casinha está ganhando forma e, aos poucos, estamos construindo um cantinho para chamar de lar.
Ter você pertinho da gente para celebrar esse novo começo torna tudo ainda mais especial.
Preparamos tudo com muito carinho e mal podemos esperar para receber você!

{{.DateLabel}} às {{.Time}}
{{.AddressLine1}}
{{.AddressLine2}}
{{if .MapsURL}}Mapa: {{.MapsURL}}{{end}}
{{if .InviteURL}}Convite completo: {{.InviteURL}}{{end}}

Com carinho,
{{.Signature}} e Pepé 🐾
`))

// Layout do Frame 6: fundo azul, cartão creme arredondado, textos em #4365BF.
var inviteHTML = template.Must(template.New("html").Parse(`<!doctype html>
<html lang="pt-BR"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width">
<title>Convite</title></head>
<body style="margin:0;padding:32px 12px;background:#4365BF;font-family:Hubballi,'Trebuchet MS',Arial,sans-serif;color:#4365BF;">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0"><tr><td align="center">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:520px;background:#E8E2D7;border-radius:32px;">
<tr><td style="padding:40px 36px 32px;text-align:center;">
  <div style="font-size:30px;line-height:1.3;">Que bom que você vem, {{.GuestName}}! ♡</div>
  <p style="font-size:17px;line-height:1.5;text-align:left;margin:24px 0 0;">Estamos muito felizes em poder compartilhar esse momento com você!</p>
  <p style="font-size:17px;line-height:1.5;text-align:left;margin:16px 0 0;">Nossa casinha está ganhando forma e, aos poucos, estamos construindo um cantinho para chamar de lar.</p>
  <p style="font-size:17px;line-height:1.5;text-align:left;margin:16px 0 0;">Ter você pertinho da gente para celebrar esse novo começo torna tudo ainda mais especial.</p>
  <p style="font-size:17px;line-height:1.5;text-align:left;margin:16px 0 0;">Preparamos tudo com muito carinho e mal podemos esperar para receber você!</p>
  <div style="font-size:26px;margin-top:28px;">{{.DateLabel}}, às {{.Time}}</div>
  <div style="font-size:20px;margin-top:8px;">{{.AddressLine1}}<br>{{.AddressLine2}}</div>
  {{if .MapsURL}}<p style="margin:24px 0 0;"><a href="{{.MapsURL}}" style="display:inline-block;background:#4365BF;color:#ffffff;text-decoration:none;font-size:18px;padding:14px 28px;border-radius:25px;">Ver no mapa</a></p>{{end}}
  {{if .InviteURL}}<p style="margin:16px 0 0;font-size:15px;"><a href="{{.InviteURL}}" style="color:#4365BF;">Abrir o convite no navegador</a></p>{{end}}
  <div style="text-align:right;margin-top:32px;font-size:16px;">Com carinho,<br><span style="font-size:22px;">{{.Signature}}</span><br><span style="font-size:15px;">e Pepé</span></div>
</td></tr></table>
</td></tr></table>
</body></html>`))
