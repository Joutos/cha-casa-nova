package services

import (
	"strings"
	"testing"
)

func TestNormalizeName(t *testing.T) {
	cases := []struct {
		in, want string
		ok       bool
	}{
		{"  Maria   da  Silva ", "Maria da Silva", true},
		{"Jo", "Jo", true},
		{"J", "", false},
		{"", "", false},
		{"<script>", "", false},
		{"Ana\x00", "", false},
		{strings.Repeat("a", 101), "", false},
		{"João Ñandú", "João Ñandú", true},
	}
	for _, c := range cases {
		got, msg := NormalizeName(c.in)
		if (msg == "") != c.ok || got != c.want {
			t.Errorf("NormalizeName(%q) = %q, %q", c.in, got, msg)
		}
	}
}

func TestNormalizeEmail(t *testing.T) {
	cases := []struct {
		in, want string
		ok       bool
	}{
		{" Maria@Example.COM ", "maria@example.com", true},
		{"maria@example", "", false},
		{"Maria <maria@example.com>", "", false},
		{"maria@@example.com", "", false},
		{"maria@example.com\r\nBcc: x@y.com", "", false},
		{"", "", false},
		{strings.Repeat("a", 250) + "@x.com", "", false},
	}
	for _, c := range cases {
		got, msg := NormalizeEmail(c.in)
		if (msg == "") != c.ok || got != c.want {
			t.Errorf("NormalizeEmail(%q) = %q, %q", c.in, got, msg)
		}
	}
}

func TestNormalizeMessage(t *testing.T) {
	got, msg := NormalizeMessage("  Oi!\r\nFelicidades\x07  ")
	if msg != "" || got != "Oi!\nFelicidades" {
		t.Fatalf("got %q %q", got, msg)
	}
	if _, msg := NormalizeMessage("   "); msg == "" {
		t.Fatal("mensagem vazia deveria falhar")
	}
	if _, msg := NormalizeMessage(strings.Repeat("é", 1001)); msg == "" {
		t.Fatal("mensagem longa deveria falhar")
	}
	if _, msg := NormalizeMessage(strings.Repeat("é", 1000)); msg != "" {
		t.Fatal("1000 caracteres multibyte deveriam passar")
	}
}

func TestSafeCSVCell(t *testing.T) {
	for in, want := range map[string]string{
		"=HYPERLINK(\"x\")": "'=HYPERLINK(\"x\")", "+1": "'+1", "-2": "'-2", "@SUM": "'@SUM", "Ana": "Ana", "": "",
	} {
		if got := SafeCSVCell(in); got != want {
			t.Errorf("SafeCSVCell(%q) = %q", in, got)
		}
	}
}

func TestPasswordHashing(t *testing.T) {
	p := Argon2Params{Memory: 1024, Time: 1, Threads: 1, SaltLen: 16, KeyLen: 32}
	h1, err := HashPassword("correta-senha-123", p)
	if err != nil {
		t.Fatal(err)
	}
	h2, _ := HashPassword("correta-senha-123", p)
	if h1 == h2 {
		t.Fatal("salt deveria tornar os hashes diferentes")
	}
	if !strings.HasPrefix(h1, "$argon2id$v=19$m=1024,t=1,p=1$") {
		t.Fatalf("formato inesperado: %s", h1)
	}
	if ok, err := VerifyPassword("correta-senha-123", h1); !ok || err != nil {
		t.Fatal("senha correta deveria validar")
	}
	if ok, _ := VerifyPassword("errada", h1); ok {
		t.Fatal("senha errada não deveria validar")
	}
	if _, err := VerifyPassword("x", "$bcrypt$abc"); err == nil {
		t.Fatal("hash inválido deveria dar erro")
	}
}

func TestTokens(t *testing.T) {
	tok, err := randomToken(32)
	if err != nil || !IsWellFormedToken(tok) {
		t.Fatalf("token mal formado: %q", tok)
	}
	if IsWellFormedToken("abc") || IsWellFormedToken(tok+"'") {
		t.Fatal("tokens inválidos aceitos")
	}
	if hashToken(tok) == tok || len(hashToken(tok)) != 64 {
		t.Fatal("hash do token inesperado")
	}
}
