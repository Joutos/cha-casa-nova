package models

import "time"

type Guest struct {
	ID                   string     `gorm:"primaryKey" json:"id"`
	Name                 string     `json:"name"`
	Email                string     `json:"email"`
	Attending            bool       `json:"attending"`
	WantsToContribute    bool       `json:"wants_to_contribute"`
	InviteToken          *string    `json:"-"`
	AttendingConfirmedAt *time.Time `json:"attending_confirmed_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	Messages             []Message  `gorm:"foreignKey:GuestID" json:"messages,omitempty"`
}

type Message struct {
	ID        string    `gorm:"primaryKey" json:"id"`
	GuestID   string    `json:"-"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}

type Admin struct {
	ID           string    `gorm:"primaryKey" json:"id"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"-"`
	UpdatedAt    time.Time `json:"-"`
}

type Session struct {
	ID         string `gorm:"primaryKey"`
	AdminID    string
	Admin      Admin `gorm:"foreignKey:AdminID"`
	TokenHash  string
	ExpiresAt  time.Time
	CreatedAt  time.Time
	LastSeenAt time.Time
}

// GuestSummary é a linha da listagem administrativa.
type GuestSummary struct {
	ID                string    `json:"id"`
	Name              string    `json:"name"`
	Email             string    `json:"email"`
	Attending         bool      `json:"attending"`
	WantsToContribute bool      `json:"wants_to_contribute"`
	MessageCount      int       `json:"message_count"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Stats struct {
	Total       int64 `json:"total"`
	Attending   int64 `json:"attending"`
	Contribute  int64 `json:"contribute"`
	WithMessage int64 `json:"with_message"`
}
