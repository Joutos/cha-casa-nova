// Package repositories concentra o acesso ao banco. Toda consulta usa
// parâmetros (placeholders) do GORM; colunas de ordenação vêm de uma lista fixa.
package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/chadecasanova/backend/internal/models"
	"gorm.io/gorm"
)

type GuestFilter string

const (
	FilterAll        GuestFilter = ""
	FilterAttending  GuestFilter = "attending"
	FilterContribute GuestFilter = "contribute"
	FilterMessage    GuestFilter = "message"
)

type ListParams struct {
	Query    string
	Filter   GuestFilter
	Sort     string // chave de sortColumns
	Desc     bool
	Page     int
	PageSize int
}

var sortColumns = map[string]string{
	"created_at": "guests.created_at",
	"name":       "guests.name COLLATE NOCASE",
	"email":      "guests.email",
}

func IsValidSort(key string) bool { _, ok := sortColumns[key]; return ok }

type GuestRepository struct{ db *gorm.DB }

func NewGuestRepository(db *gorm.DB) *GuestRepository { return &GuestRepository{db: db} }

// Transaction executa fn numa transação com um repositório ligado a ela.
func (r *GuestRepository) Transaction(ctx context.Context, fn func(tx *GuestRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&GuestRepository{db: tx})
	})
}

func (r *GuestRepository) FindByEmail(ctx context.Context, email string) (*models.Guest, error) {
	var g models.Guest
	err := r.db.WithContext(ctx).Where("email = ?", email).Take(&g).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &g, err
}

func (r *GuestRepository) FindByInviteToken(ctx context.Context, token string) (*models.Guest, error) {
	var g models.Guest
	err := r.db.WithContext(ctx).Where("invite_token = ? AND attending = 1", token).Take(&g).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &g, err
}

func (r *GuestRepository) Get(ctx context.Context, id string) (*models.Guest, error) {
	var g models.Guest
	err := r.db.WithContext(ctx).
		Preload("Messages", func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC") }).
		Where("id = ?", id).Take(&g).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &g, err
}

func (r *GuestRepository) Create(ctx context.Context, g *models.Guest) error {
	return r.db.WithContext(ctx).Omit("Messages").Create(g).Error
}

func (r *GuestRepository) Save(ctx context.Context, g *models.Guest) error {
	return r.db.WithContext(ctx).Omit("Messages", "CreatedAt").Save(g).Error
}

func (r *GuestRepository) AddMessage(ctx context.Context, m *models.Message) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *GuestRepository) CountMessages(ctx context.Context, guestID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.Message{}).Where("guest_id = ?", guestID).Count(&n).Error
	return n, err
}

func (r *GuestRepository) Delete(ctx context.Context, id string) (bool, error) {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Guest{})
	return res.RowsAffected > 0, res.Error
}

// escapeLike neutraliza os curingas do LIKE no termo digitado.
func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

func (r *GuestRepository) filtered(ctx context.Context, p ListParams) *gorm.DB {
	q := r.db.WithContext(ctx).Model(&models.Guest{})
	if term := strings.TrimSpace(p.Query); term != "" {
		like := "%" + escapeLike(term) + "%"
		q = q.Where(`(guests.name LIKE ? ESCAPE '\' OR guests.email LIKE ? ESCAPE '\'
			OR EXISTS (SELECT 1 FROM messages m WHERE m.guest_id = guests.id AND m.body LIKE ? ESCAPE '\'))`,
			like, like, like)
	}
	switch p.Filter {
	case FilterAttending:
		q = q.Where("guests.attending = 1")
	case FilterContribute:
		q = q.Where("guests.wants_to_contribute = 1")
	case FilterMessage:
		q = q.Where("EXISTS (SELECT 1 FROM messages m WHERE m.guest_id = guests.id)")
	}
	return q
}

func (r *GuestRepository) orderClause(p ListParams) string {
	col, ok := sortColumns[p.Sort]
	if !ok {
		col = sortColumns["created_at"]
	}
	dir := " ASC"
	if p.Desc {
		dir = " DESC"
	}
	return col + dir + ", guests.id ASC"
}

const summaryColumns = `guests.id, guests.name, guests.email, guests.attending, guests.wants_to_contribute,
	guests.created_at, guests.updated_at,
	(SELECT COUNT(*) FROM messages m WHERE m.guest_id = guests.id) AS message_count`

func (r *GuestRepository) List(ctx context.Context, p ListParams) ([]models.GuestSummary, int64, error) {
	var total int64
	if err := r.filtered(ctx, p).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	rows := make([]models.GuestSummary, 0, p.PageSize)
	err := r.filtered(ctx, p).
		Select(summaryColumns).
		Order(r.orderClause(p)).
		Limit(p.PageSize).
		Offset((p.Page - 1) * p.PageSize).
		Scan(&rows).Error
	return rows, total, err
}

// ListForExport devolve todos os convidados do filtro com suas mensagens.
func (r *GuestRepository) ListForExport(ctx context.Context, p ListParams) ([]models.Guest, error) {
	var guests []models.Guest
	err := r.filtered(ctx, p).
		Preload("Messages", func(db *gorm.DB) *gorm.DB { return db.Order("created_at ASC") }).
		Order(r.orderClause(p)).
		Find(&guests).Error
	return guests, err
}

func (r *GuestRepository) Stats(ctx context.Context) (models.Stats, error) {
	var s models.Stats
	err := r.db.WithContext(ctx).Raw(`SELECT
		COUNT(*) AS total,
		COALESCE(SUM(attending), 0) AS attending,
		COALESCE(SUM(wants_to_contribute), 0) AS contribute,
		COALESCE(SUM(EXISTS (SELECT 1 FROM messages m WHERE m.guest_id = guests.id)), 0) AS with_message
		FROM guests`).Scan(&s).Error
	return s, err
}
