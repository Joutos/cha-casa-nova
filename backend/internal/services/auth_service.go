package services

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"log/slog"
	"time"
	"unicode/utf8"

	"github.com/chadecasanova/backend/internal/apperr"
	"github.com/chadecasanova/backend/internal/models"
	"github.com/chadecasanova/backend/internal/repositories"
	"github.com/google/uuid"
)

const PasswordMinLength = 12

var errInvalidCredentials = apperr.Unauthorized("E-mail ou senha incorretos.")

type AuthService struct {
	repo      *repositories.AuthRepository
	secret    []byte
	ttl       time.Duration
	argon     Argon2Params
	dummyHash string
	now       func() time.Time
}

func NewAuthService(repo *repositories.AuthRepository, secret []byte, ttl time.Duration, argon Argon2Params) (*AuthService, error) {
	// Hash fictício usado quando o e-mail não existe: o tempo de resposta fica
	// igual ao de uma senha errada, evitando enumeração de administradores.
	dummy, err := HashPassword("dummy-password-for-timing", argon)
	if err != nil {
		return nil, err
	}
	return &AuthService{repo: repo, secret: secret, ttl: ttl, argon: argon, dummyHash: dummy,
		now: func() time.Time { return time.Now().UTC() }}, nil
}

type LoginResult struct {
	Token     string
	ExpiresAt time.Time
	Admin     *models.Admin
}

func (s *AuthService) Login(ctx context.Context, rawEmail, password string) (*LoginResult, error) {
	email, msg := NormalizeEmail(rawEmail)
	if msg != "" || password == "" || len(password) > 1024 {
		_, _ = VerifyPassword(password, s.dummyHash)
		return nil, errInvalidCredentials
	}
	admin, err := s.repo.FindAdminByEmail(ctx, email)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if admin == nil {
		_, _ = VerifyPassword(password, s.dummyHash)
		return nil, errInvalidCredentials
	}
	ok, err := VerifyPassword(password, admin.PasswordHash)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if !ok {
		return nil, errInvalidCredentials
	}

	token, err := randomToken(32)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	now := s.now()
	sess := &models.Session{
		ID: uuid.NewString(), AdminID: admin.ID, TokenHash: hashToken(token),
		ExpiresAt: now.Add(s.ttl), CreatedAt: now, LastSeenAt: now,
	}
	if err := s.repo.CreateSession(ctx, sess); err != nil {
		return nil, apperr.Internal(err)
	}
	if _, err := s.repo.DeleteExpiredSessions(ctx, now); err != nil {
		slog.Warn("falha ao limpar sessões expiradas", "error", err)
	}
	return &LoginResult{Token: token, ExpiresAt: sess.ExpiresAt, Admin: admin}, nil
}

// Authenticate valida o token do cookie e devolve o administrador da sessão.
func (s *AuthService) Authenticate(ctx context.Context, token string) (*models.Admin, error) {
	if !IsWellFormedToken(token) {
		return nil, apperr.Unauthorized("Sua sessão expirou. Entre novamente.")
	}
	now := s.now()
	sess, err := s.repo.FindValidSession(ctx, hashToken(token), now)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if sess == nil || sess.Admin.ID == "" {
		return nil, apperr.Unauthorized("Sua sessão expirou. Entre novamente.")
	}
	if now.Sub(sess.LastSeenAt) > time.Minute {
		_ = s.repo.TouchSession(ctx, sess.ID, now)
	}
	admin := sess.Admin
	return &admin, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	if !IsWellFormedToken(token) {
		return nil
	}
	if err := s.repo.DeleteSessionByHash(ctx, hashToken(token)); err != nil {
		return apperr.Internal(err)
	}
	return nil
}

// CSRFToken deriva o token anti-CSRF da sessão via HMAC; nada extra é armazenado.
func (s *AuthService) CSRFToken(sessionToken string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte("csrf:" + sessionToken))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *AuthService) VerifyCSRF(sessionToken, provided string) bool {
	if provided == "" {
		return false
	}
	return hmac.Equal([]byte(s.CSRFToken(sessionToken)), []byte(provided))
}

func (s *AuthService) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	return s.repo.DeleteExpiredSessions(ctx, s.now())
}

func validateNewAdmin(name, email, password string) (string, string, map[string]string) {
	fields := map[string]string{}
	n, msg := NormalizeName(name)
	if msg != "" {
		fields["name"] = msg
	}
	e, msg := NormalizeEmail(email)
	if msg != "" {
		fields["email"] = msg
	}
	if utf8.RuneCountInString(password) < PasswordMinLength {
		fields["password"] = "A senha precisa ter pelo menos 12 caracteres."
	}
	return n, e, fields
}

// CreateAdmin é usado pelo comando CLI e pelo bootstrap; não há rota pública para isso.
func (s *AuthService) CreateAdmin(ctx context.Context, name, email, password string) (*models.Admin, error) {
	n, e, fields := validateNewAdmin(name, email, password)
	if len(fields) > 0 {
		return nil, apperr.Validation(fields)
	}
	existing, err := s.repo.FindAdminByEmail(ctx, e)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	if existing != nil {
		return nil, apperr.Conflict("admin_exists", "Já existe um administrador com este e-mail.")
	}
	hash, err := HashPassword(password, s.argon)
	if err != nil {
		return nil, apperr.Internal(err)
	}
	now := s.now()
	a := &models.Admin{ID: uuid.NewString(), Name: n, Email: e, PasswordHash: hash, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.CreateAdmin(ctx, a); err != nil {
		return nil, apperr.Internal(err)
	}
	return a, nil
}

// ResetPassword troca a senha e encerra todas as sessões do administrador.
func (s *AuthService) ResetPassword(ctx context.Context, email, password string) error {
	e, msg := NormalizeEmail(email)
	if msg != "" {
		return apperr.Validation(map[string]string{"email": msg})
	}
	if utf8.RuneCountInString(password) < PasswordMinLength {
		return apperr.Validation(map[string]string{"password": "A senha precisa ter pelo menos 12 caracteres."})
	}
	a, err := s.repo.FindAdminByEmail(ctx, e)
	if err != nil {
		return apperr.Internal(err)
	}
	if a == nil {
		return apperr.NotFound("Administrador não encontrado.")
	}
	hash, err := HashPassword(password, s.argon)
	if err != nil {
		return apperr.Internal(err)
	}
	if err := s.repo.UpdateAdminPassword(ctx, a.ID, hash); err != nil {
		return apperr.Internal(err)
	}
	return s.repo.DeleteSessionsForAdmin(ctx, a.ID)
}

// EnsureBootstrapAdmin cria o primeiro administrador a partir das variáveis de
// ambiente, somente se ainda não houver nenhum. Depois disso as variáveis são ignoradas.
func (s *AuthService) EnsureBootstrapAdmin(ctx context.Context, name, email, password string) error {
	if email == "" || password == "" {
		return nil
	}
	count, err := s.repo.CountAdmins(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	if _, err := s.CreateAdmin(ctx, name, email, password); err != nil {
		return err
	}
	slog.Info("primeiro administrador criado a partir de ADMIN_EMAIL/ADMIN_PASSWORD; remova ADMIN_PASSWORD do ambiente")
	return nil
}
