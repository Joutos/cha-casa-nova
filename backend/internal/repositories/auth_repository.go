package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/chadecasanova/backend/internal/models"
	"gorm.io/gorm"
)

type AuthRepository struct{ db *gorm.DB }

func NewAuthRepository(db *gorm.DB) *AuthRepository { return &AuthRepository{db: db} }

func (r *AuthRepository) FindAdminByEmail(ctx context.Context, email string) (*models.Admin, error) {
	var a models.Admin
	err := r.db.WithContext(ctx).Where("email = ?", email).Take(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &a, err
}

func (r *AuthRepository) CountAdmins(ctx context.Context) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&models.Admin{}).Count(&n).Error
	return n, err
}

func (r *AuthRepository) CreateAdmin(ctx context.Context, a *models.Admin) error {
	return r.db.WithContext(ctx).Create(a).Error
}

func (r *AuthRepository) UpdateAdminPassword(ctx context.Context, id, hash string) error {
	return r.db.WithContext(ctx).Model(&models.Admin{}).Where("id = ?", id).
		Updates(map[string]any{"password_hash": hash, "updated_at": time.Now().UTC()}).Error
}

func (r *AuthRepository) CreateSession(ctx context.Context, s *models.Session) error {
	return r.db.WithContext(ctx).Omit("Admin").Create(s).Error
}

// FindValidSession devolve a sessão (com o administrador) se ainda não expirou.
func (r *AuthRepository) FindValidSession(ctx context.Context, tokenHash string, now time.Time) (*models.Session, error) {
	var s models.Session
	err := r.db.WithContext(ctx).Joins("Admin").
		Where("sessions.token_hash = ? AND sessions.expires_at > ?", tokenHash, now).
		Take(&s).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &s, err
}

func (r *AuthRepository) TouchSession(ctx context.Context, id string, now time.Time) error {
	return r.db.WithContext(ctx).Model(&models.Session{}).Where("id = ?", id).Update("last_seen_at", now).Error
}

func (r *AuthRepository) DeleteSessionByHash(ctx context.Context, tokenHash string) error {
	return r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).Delete(&models.Session{}).Error
}

func (r *AuthRepository) DeleteExpiredSessions(ctx context.Context, now time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Where("expires_at <= ?", now).Delete(&models.Session{})
	return res.RowsAffected, res.Error
}

func (r *AuthRepository) DeleteSessionsForAdmin(ctx context.Context, adminID string) error {
	return r.db.WithContext(ctx).Where("admin_id = ?", adminID).Delete(&models.Session{}).Error
}
