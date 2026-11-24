package services

import (
	"context"
	"encoding/csv"
	"errors"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/chadecasanova/backend/internal/apperr"
	"github.com/chadecasanova/backend/internal/config"
	"github.com/chadecasanova/backend/internal/models"
	"github.com/chadecasanova/backend/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrSpam indica que o honeypot foi preenchido. O handler responde como se
// tivesse dado certo, para não ensinar o robô, mas nada é gravado.
var ErrSpam = errors.New("spam detectado")

type RegistrationService struct {
	repo        *repositories.GuestRepository
	mailer      Mailer
	event       config.Event
	baseURL     string
	maxMessages int
	now         func() time.Time
}

func NewRegistrationService(repo *repositories.GuestRepository, mailer Mailer, event config.Event, baseURL string, maxMessages int) *RegistrationService {
	if mailer == nil {
		mailer = NoopMailer{}
	}
	return &RegistrationService{repo: repo, mailer: mailer, event: event, baseURL: baseURL,
		maxMessages: maxMessages, now: func() time.Time { return time.Now().UTC() }}
}

type AttendanceInput struct {
	Name       string
	Email      string
	Contribute bool
	Honeypot   string
}

type MessageInput struct {
	Name       string
	Email      string
	Body       string
	Contribute bool
	Honeypot   string
}

func (s *RegistrationService) ConfirmAttendance(ctx context.Context, in AttendanceInput) (*models.Guest, error) {
	if strings.TrimSpace(in.Honeypot) != "" {
		return nil, ErrSpam
	}
	fields := map[string]string{}
	name, msg := NormalizeName(in.Name)
	if msg != "" {
		fields["name"] = msg
	}
	email, msg := NormalizeEmail(in.Email)
	if msg != "" {
		fields["email"] = msg
	}
	if len(fields) > 0 {
		return nil, apperr.Validation(fields)
	}

	var guest *models.Guest
	err := s.repo.Transaction(ctx, func(tx *repositories.GuestRepository) error {
		existing, err := tx.FindByEmail(ctx, email)
		if err != nil {
			return err
		}
		now := s.now()
		token, err := randomToken(32)
		if err != nil {
			return err
		}
		if existing != nil {
			if existing.Attending {
				return apperr.Conflict("already_confirmed",
					"Este e-mail já confirmou presença. Se precisar mudar algo, fale com a gente.")
			}
			existing.Name = name
			existing.Attending = true
			existing.WantsToContribute = existing.WantsToContribute || in.Contribute
			existing.AttendingConfirmedAt = &now
			if existing.InviteToken == nil {
				existing.InviteToken = &token
			}
			existing.UpdatedAt = now
			guest = existing
			return tx.Save(ctx, existing)
		}
		guest = &models.Guest{
			ID: uuid.NewString(), Name: name, Email: email, Attending: true,
			WantsToContribute: in.Contribute, InviteToken: &token, AttendingConfirmedAt: &now,
			CreatedAt: now, UpdatedAt: now,
		}
		return tx.Create(ctx, guest)
	})
	if err != nil {
		return nil, wrap(err)
	}

	s.sendInviteAsync(guest)
	return guest, nil
}

func (s *RegistrationService) LeaveMessage(ctx context.Context, in MessageInput) (*models.Message, error) {
	if strings.TrimSpace(in.Honeypot) != "" {
		return nil, ErrSpam
	}
	fields := map[string]string{}
	name, msg := NormalizeName(in.Name)
	if msg != "" {
		fields["name"] = msg
	}
	email, msg := NormalizeEmail(in.Email)
	if msg != "" {
		fields["email"] = msg
	}
	body, msg := NormalizeMessage(in.Body)
	if msg != "" {
		fields["message"] = msg
	}
	if len(fields) > 0 {
		return nil, apperr.Validation(fields)
	}

	var message *models.Message
	err := s.repo.Transaction(ctx, func(tx *repositories.GuestRepository) error {
		now := s.now()
		guest, err := tx.FindByEmail(ctx, email)
		if err != nil {
			return err
		}
		if guest == nil {
			guest = &models.Guest{ID: uuid.NewString(), Name: name, Email: email,
				WantsToContribute: in.Contribute, CreatedAt: now, UpdatedAt: now}
			if err := tx.Create(ctx, guest); err != nil {
				return err
			}
		} else {
			count, err := tx.CountMessages(ctx, guest.ID)
			if err != nil {
				return err
			}
			if int(count) >= s.maxMessages {
				return apperr.Conflict("message_limit",
					"Você já nos deixou várias mensagens. Obrigado por tanto carinho!")
			}
			if in.Contribute && !guest.WantsToContribute {
				guest.WantsToContribute = true
				guest.UpdatedAt = now
				if err := tx.Save(ctx, guest); err != nil {
					return err
				}
			}
		}
		message = &models.Message{ID: uuid.NewString(), GuestID: guest.ID, Body: body, CreatedAt: now}
		return tx.AddMessage(ctx, message)
	})
	if err != nil {
		return nil, wrap(err)
	}
	return message, nil
}

type Invite struct {
	GuestName    string `json:"guest_name"`
	DateLabel    string `json:"date_label"`
	Time         string `json:"time"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	MapsQuery    string `json:"maps_query"`
	Signature    string `json:"signature"`
}

// GetInvite devolve os dados do convite (inclusive o endereço) apenas para quem
// tem o token pessoal enviado por e-mail.
func (s *RegistrationService) GetInvite(ctx context.Context, token string) (*Invite, error) {
	if !IsWellFormedToken(token) {
		return nil, apperr.NotFound("Convite não encontrado.")
	}
	g, err := s.repo.FindByInviteToken(ctx, token)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if g == nil {
		return nil, apperr.NotFound("Convite não encontrado.")
	}
	return s.inviteFor(g), nil
}

func (s *RegistrationService) inviteFor(g *models.Guest) *Invite {
	return &Invite{GuestName: g.Name, DateLabel: s.event.DateLabel, Time: s.event.Time,
		AddressLine1: s.event.AddressLine1, AddressLine2: s.event.AddressLine2,
		MapsQuery: s.event.MapsQuery, Signature: s.event.Signature}
}

func (s *RegistrationService) InviteURL(g *models.Guest) string {
	if g.InviteToken == nil {
		return ""
	}
	return s.baseURL + "/convite/" + *g.InviteToken
}

func (s *RegistrationService) sendInviteAsync(g *models.Guest) {
	if !s.mailer.Enabled() {
		return
	}
	email := InviteEmail{To: g.Email, Invite: *s.inviteFor(g), InviteURL: s.InviteURL(g)}
	guestID := g.ID
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.mailer.SendInvite(ctx, email); err != nil {
			slog.Error("falha ao enviar convite por e-mail", "guest_id", guestID, "error", err)
			return
		}
		slog.Info("convite enviado", "guest_id", guestID)
	}()
}

// ---------- Administração ----------

type AdminListResult struct {
	Items    []models.GuestSummary `json:"items"`
	Total    int64                 `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

type GuestDetail struct {
	models.Guest
	InviteURL string `json:"invite_url,omitempty"`
}

type GuestPatch struct {
	Name              *string `json:"name"`
	Attending         *bool   `json:"attending"`
	WantsToContribute *bool   `json:"wants_to_contribute"`
}

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// ListQuery são os parâmetros crus da query string administrativa.
type ListQuery struct {
	Q, Type, Sort, Order string
	Page, PageSize       int
}

func (q ListQuery) Normalize() (repositories.ListParams, error) {
	return NormalizeListParams(q.Q, q.Type, q.Sort, q.Order, q.Page, q.PageSize)
}

// NormalizeListParams aplica valores padrão e rejeita filtros/ordenações desconhecidos.
func NormalizeListParams(q, filter, sort, order string, page, pageSize int) (repositories.ListParams, error) {
	fields := map[string]string{}
	p := repositories.ListParams{Query: strings.TrimSpace(q), Page: page, PageSize: pageSize, Sort: sort, Desc: true}
	if len([]rune(p.Query)) > 100 {
		fields["q"] = "A busca pode ter no máximo 100 caracteres."
	}
	switch repositories.GuestFilter(filter) {
	case repositories.FilterAll, repositories.FilterAttending, repositories.FilterContribute, repositories.FilterMessage:
		p.Filter = repositories.GuestFilter(filter)
	default:
		fields["type"] = "Filtro desconhecido."
	}
	if p.Sort == "" {
		p.Sort = "created_at"
	}
	if !repositories.IsValidSort(p.Sort) {
		fields["sort"] = "Ordenação desconhecida."
	}
	switch order {
	case "", "desc":
		p.Desc = true
	case "asc":
		p.Desc = false
	default:
		fields["order"] = "Use asc ou desc."
	}
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize < 1 {
		p.PageSize = DefaultPageSize
	}
	if p.PageSize > MaxPageSize {
		p.PageSize = MaxPageSize
	}
	if len(fields) > 0 {
		return p, apperr.Validation(fields)
	}
	return p, nil
}

func (s *RegistrationService) List(ctx context.Context, p repositories.ListParams) (*AdminListResult, error) {
	items, total, err := s.repo.List(ctx, p)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	return &AdminListResult{Items: items, Total: total, Page: p.Page, PageSize: p.PageSize}, nil
}

func (s *RegistrationService) Stats(ctx context.Context) (models.Stats, error) {
	st, err := s.repo.Stats(ctx)
	if err != nil {
		return st, apperr.Internal(err)
	}
	return st, nil
}

func (s *RegistrationService) Get(ctx context.Context, id string) (*GuestDetail, error) {
	if !IsUUID(id) {
		return nil, apperr.NotFound("Inscrição não encontrada.")
	}
	g, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if g == nil {
		return nil, apperr.NotFound("Inscrição não encontrada.")
	}
	if g.Messages == nil {
		g.Messages = []models.Message{}
	}
	d := &GuestDetail{Guest: *g}
	if g.Attending {
		d.InviteURL = s.InviteURL(g)
	}
	return d, nil
}

func (s *RegistrationService) Update(ctx context.Context, id string, patch GuestPatch) (*GuestDetail, error) {
	if !IsUUID(id) {
		return nil, apperr.NotFound("Inscrição não encontrada.")
	}
	if patch.Name == nil && patch.Attending == nil && patch.WantsToContribute == nil {
		return nil, apperr.BadRequest("empty_patch", "Nada para atualizar.")
	}
	err := s.repo.Transaction(ctx, func(tx *repositories.GuestRepository) error {
		g, err := tx.Get(ctx, id)
		if err != nil {
			return err
		}
		if g == nil {
			return apperr.NotFound("Inscrição não encontrada.")
		}
		now := s.now()
		if patch.Name != nil {
			name, msg := NormalizeName(*patch.Name)
			if msg != "" {
				return apperr.Validation(map[string]string{"name": msg})
			}
			g.Name = name
		}
		if patch.Attending != nil && *patch.Attending != g.Attending {
			g.Attending = *patch.Attending
			if g.Attending {
				g.AttendingConfirmedAt = &now
				if g.InviteToken == nil {
					t, err := randomToken(32)
					if err != nil {
						return err
					}
					g.InviteToken = &t
				}
			} else {
				g.AttendingConfirmedAt = nil
			}
		}
		if patch.WantsToContribute != nil {
			g.WantsToContribute = *patch.WantsToContribute
		}
		g.UpdatedAt = now
		return tx.Save(ctx, g)
	})
	if err != nil {
		return nil, wrap(err)
	}
	return s.Get(ctx, id)
}

func (s *RegistrationService) Delete(ctx context.Context, id string) error {
	if !IsUUID(id) {
		return apperr.NotFound("Inscrição não encontrada.")
	}
	ok, err := s.repo.Delete(ctx, id)
	if err != nil {
		return apperr.Internal(err)
	}
	if !ok {
		return apperr.NotFound("Inscrição não encontrada.")
	}
	return nil
}

// ExportCSV escreve um CSV (UTF-8 com BOM, separador ';', compatível com o Excel em pt-BR).
func (s *RegistrationService) ExportCSV(ctx context.Context, p repositories.ListParams, w io.Writer) error {
	guests, err := s.repo.ListForExport(ctx, p)
	if err != nil {
		return apperr.Internal(err)
	}
	if _, err := w.Write([]byte("\ufeff")); err != nil {
		return err
	}
	cw := csv.NewWriter(w)
	cw.Comma = ';'
	loc, _ := time.LoadLocation("America/Sao_Paulo")
	if loc == nil {
		loc = time.UTC
	}
	_ = cw.Write([]string{"Nome", "E-mail", "Presença", "Contribuição", "Mensagens", "Inscrito em"})
	for _, g := range guests {
		msgs := make([]string, 0, len(g.Messages))
		for _, m := range g.Messages {
			msgs = append(msgs, m.Body)
		}
		if err := cw.Write([]string{
			SafeCSVCell(g.Name), SafeCSVCell(g.Email), yesNo(g.Attending), yesNo(g.WantsToContribute),
			SafeCSVCell(strings.Join(msgs, " | ")), g.CreatedAt.In(loc).Format("02/01/2006 15:04"),
		}); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}

// SafeCSVCell evita injeção de fórmulas ao abrir o CSV em planilhas.
func SafeCSVCell(v string) string {
	if v == "" {
		return v
	}
	switch v[0] {
	case '=', '+', '-', '@', '\t', '\r':
		return "'" + v
	}
	return v
}

func yesNo(b bool) string {
	if b {
		return "Sim"
	}
	return "Não"
}

func wrap(err error) error {
	var ae *apperr.Error
	if errors.As(err, &ae) {
		return ae
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return apperr.Conflict("already_confirmed", "Este e-mail já foi registrado agora há pouco.")
	}
	return apperr.Internal(err)
}
