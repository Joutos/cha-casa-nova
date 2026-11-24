package services

import (
	"strings"
	"testing"
)

func TestBuildInviteMessageEscapesHTML(t *testing.T) {
	msg, err := BuildInviteMessage("Sofia e João <convite@example.com>", InviteEmail{
		To: "ana@example.com", InviteURL: "https://festa.example/convite/abc",
		Invite: Invite{GuestName: `<img src=x onerror=alert(1)>`, DateLabel: "23 de Janeiro de 2027", Time: "11h00",
			AddressLine1: "Rua A, 1", AddressLine2: "BH", MapsQuery: "Rua A 1 BH", Signature: "Sofia e João"},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(msg)
	htmlPart := s[strings.Index(s, "Content-Type: text/html"):]
	if strings.Contains(htmlPart, "<img src=x") || !strings.Contains(htmlPart, "&lt;img src=x") {
		t.Fatal("nome do convidado deveria ser escapado no HTML")
	}
	for _, want := range []string{"To: ana@example.com\r\n", "multipart/alternative", "https://www.google.com/maps/search/?api=1&amp;query=Rua&#43;A&#43;1&#43;BH", "11h00"} {
		if !strings.Contains(s, want) {
			t.Errorf("faltando %q", want)
		}
	}
}
