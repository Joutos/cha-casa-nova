package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/chadecasanova/backend/internal/apperr"
	"github.com/chadecasanova/backend/internal/repositories"
	"github.com/chadecasanova/backend/internal/services"
	"github.com/chadecasanova/backend/internal/testutil"
)

func code(err error) string {
	var e *apperr.Error
	if errors.As(err, &e) {
		return e.Code
	}
	return ""
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	for i := 0; i < 100; i++ {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condição não atingida")
}

func TestConfirmAttendance(t *testing.T) {
	app := testutil.NewApp(t)
	ctx := context.Background()
	s := app.Registrations

	g, err := s.ConfirmAttendance(ctx, services.AttendanceInput{Name: " Ana  Lima ", Email: "ANA@Example.com", Contribute: true})
	if err != nil {
		t.Fatal(err)
	}
	if g.Name != "Ana Lima" || g.Email != "ana@example.com" || !g.Attending || !g.WantsToContribute || g.InviteToken == nil {
		t.Fatalf("convidado inesperado: %+v", g)
	}
	waitFor(t, func() bool { return app.Mailer.Count() == 1 })
	if sent := app.Mailer.Sent[0]; sent.To != "ana@example.com" || sent.InviteURL != "https://festa.example/convite/"+*g.InviteToken {
		t.Fatalf("e-mail inesperado: %+v", sent)
	}

	_, err = s.ConfirmAttendance(ctx, services.AttendanceInput{Name: "Ana", Email: "ana@EXAMPLE.com"})
	if code(err) != "already_confirmed" {
		t.Fatalf("esperava conflito, veio %v", err)
	}
}

func TestConfirmAttendanceValidationAndSpam(t *testing.T) {
	s := testutil.NewApp(t).Registrations
	_, err := s.ConfirmAttendance(context.Background(), services.AttendanceInput{Name: "A", Email: "x"})
	var e *apperr.Error
	if !errors.As(err, &e) || e.Fields["name"] == "" || e.Fields["email"] == "" {
		t.Fatalf("esperava erros de validação, veio %v", err)
	}
	_, err = s.ConfirmAttendance(context.Background(), services.AttendanceInput{Name: "Robô", Email: "bot@spam.com", Honeypot: "http://spam"})
	if !errors.Is(err, services.ErrSpam) {
		t.Fatalf("honeypot deveria acusar spam, veio %v", err)
	}
	res, _ := s.List(context.Background(), repositories.ListParams{Page: 1, PageSize: 10, Sort: "created_at"})
	if res.Total != 0 {
		t.Fatal("nada deveria ter sido gravado")
	}
}

func TestMessageThenAttendance(t *testing.T) {
	app := testutil.NewApp(t)
	ctx := context.Background()
	s := app.Registrations

	if _, err := s.LeaveMessage(ctx, services.MessageInput{Name: "Bia", Email: "bia@example.com", Body: "Felicidades!"}); err != nil {
		t.Fatal(err)
	}
	// Quem deixou mensagem ainda pode confirmar presença depois.
	g, err := s.ConfirmAttendance(ctx, services.AttendanceInput{Name: "Bia Souza", Email: "bia@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	d, err := s.Get(ctx, g.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !d.Attending || d.Name != "Bia Souza" || len(d.Messages) != 1 || d.Messages[0].Body != "Felicidades!" || d.InviteURL == "" {
		t.Fatalf("detalhe inesperado: %+v", d)
	}
}

func TestMessageLimit(t *testing.T) {
	s := testutil.NewApp(t).Registrations // limite = 3
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if _, err := s.LeaveMessage(ctx, services.MessageInput{Name: "Caio", Email: "caio@example.com", Body: "Oi"}); err != nil {
			t.Fatal(err)
		}
	}
	_, err := s.LeaveMessage(ctx, services.MessageInput{Name: "Caio", Email: "caio@example.com", Body: "Oi de novo"})
	if code(err) != "message_limit" {
		t.Fatalf("esperava limite, veio %v", err)
	}
}

func TestListFilterSortAndPaging(t *testing.T) {
	app := testutil.NewApp(t)
	ctx := context.Background()
	s := app.Registrations
	_, _ = s.ConfirmAttendance(ctx, services.AttendanceInput{Name: "Carla", Email: "carla@example.com"})
	_, _ = s.ConfirmAttendance(ctx, services.AttendanceInput{Name: "Bruno", Email: "bruno@example.com", Contribute: true})
	_, _ = s.LeaveMessage(ctx, services.MessageInput{Name: "Alice", Email: "alice@example.com", Body: "Bolo de 100% chocolate"})

	p, _ := services.NormalizeListParams("", "", "name", "asc", 1, 2)
	res, err := s.List(ctx, p)
	if err != nil {
		t.Fatal(err)
	}
	if res.Total != 3 || len(res.Items) != 2 || res.Items[0].Name != "Alice" || res.Items[1].Name != "Bruno" {
		t.Fatalf("página 1 inesperada: %+v", res)
	}
	if res.Items[0].MessageCount != 1 {
		t.Fatal("contagem de mensagens errada")
	}

	p, _ = services.NormalizeListParams("", "attending", "", "", 0, 0)
	if res, _ := s.List(ctx, p); res.Total != 2 || res.PageSize != services.DefaultPageSize {
		t.Fatalf("filtro presença: %+v", res)
	}
	p, _ = services.NormalizeListParams("", "contribute", "", "", 1, 10)
	if res, _ := s.List(ctx, p); res.Total != 1 || res.Items[0].Name != "Bruno" {
		t.Fatalf("filtro contribuição: %+v", res)
	}
	// Busca pelo texto da mensagem e escape de curingas do LIKE.
	p, _ = services.NormalizeListParams("100%", "", "", "", 1, 10)
	if res, _ := s.List(ctx, p); res.Total != 1 || res.Items[0].Name != "Alice" {
		t.Fatalf("busca na mensagem: %+v", res)
	}
	p, _ = services.NormalizeListParams("%", "", "", "", 1, 10)
	if res, _ := s.List(ctx, p); res.Total != 1 {
		t.Fatalf("'%%' deveria ser literal, total=%d", res.Total)
	}

	if _, err := services.NormalizeListParams("", "nope", "id; DROP TABLE guests", "sideways", 1, 10); code(err) != "validation_error" {
		t.Fatal("parâmetros inválidos deveriam ser rejeitados")
	}

	st, _ := s.Stats(ctx)
	if st.Total != 3 || st.Attending != 2 || st.Contribute != 1 || st.WithMessage != 1 {
		t.Fatalf("stats: %+v", st)
	}
}

func TestUpdateAndDelete(t *testing.T) {
	app := testutil.NewApp(t)
	s := app.Registrations
	ctx := context.Background()
	if _, err := s.LeaveMessage(ctx, services.MessageInput{Name: "Duda", Email: "duda@example.com", Body: "Oi"}); err != nil {
		t.Fatal(err)
	}
	p, _ := services.NormalizeListParams("", "", "", "", 1, 10)
	res, _ := s.List(ctx, p)
	id := res.Items[0].ID

	yes := true
	name := "  Duda Reis "
	d, err := s.Update(ctx, id, services.GuestPatch{Name: &name, Attending: &yes})
	if err != nil {
		t.Fatal(err)
	}
	if d.Name != "Duda Reis" || !d.Attending || d.AttendingConfirmedAt == nil || d.InviteURL == "" {
		t.Fatalf("atualização inesperada: %+v", d)
	}
	if _, err := s.Update(ctx, id, services.GuestPatch{}); code(err) != "empty_patch" {
		t.Fatal("patch vazio deveria falhar")
	}
	bad := "x"
	if _, err := s.Update(ctx, id, services.GuestPatch{Name: &bad}); code(err) != "validation_error" {
		t.Fatal("nome inválido deveria falhar")
	}
	if err := s.Delete(ctx, id); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, id); code(err) != "not_found" {
		t.Fatal("segunda exclusão deveria dar not_found")
	}
	// Mensagens são removidas em cascata (chave estrangeira ON DELETE CASCADE).
	var n int64
	app.DB.Raw("SELECT COUNT(*) FROM messages").Scan(&n)
	if n != 0 {
		t.Fatalf("mensagens órfãs: %d", n)
	}
	if err := s.Delete(ctx, "not-a-uuid"); code(err) != "not_found" {
		t.Fatal("id inválido deveria dar not_found")
	}
}

func TestInvite(t *testing.T) {
	s := testutil.NewApp(t).Registrations
	ctx := context.Background()
	g, _ := s.ConfirmAttendance(ctx, services.AttendanceInput{Name: "Eva", Email: "eva@example.com"})
	inv, err := s.GetInvite(ctx, *g.InviteToken)
	if err != nil || inv.GuestName != "Eva" || inv.AddressLine1 != "Rua Teste, 1" {
		t.Fatalf("convite inesperado: %+v %v", inv, err)
	}
	if _, err := s.GetInvite(ctx, "invalido"); code(err) != "not_found" {
		t.Fatal("token inválido deveria dar not_found")
	}
}
