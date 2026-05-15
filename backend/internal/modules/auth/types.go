package auth

import "time"

type User struct {
	ID           uint64     `json:"id"`
	Username     string     `json:"username"`
	Email        string     `json:"email,omitempty"`
	IsActive     bool       `json:"is_active"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	FullName     string     `json:"full_name,omitempty"`
	Phone        string     `json:"phone,omitempty"`
	PhotoURL     string     `json:"photo_url,omitempty"`
	Roles        []Role     `json:"roles"`
	RoleCodes    []string   `json:"role_codes"`
	PasswordHash string     `json:"-"`
}

type Role struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
}

type LoginRequest struct {
	Identifier string `json:"identifier"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

type TokenClaims struct {
	Sub      uint64   `json:"sub"`
	Username string   `json:"username"`
	Email    string   `json:"email,omitempty"`
	FullName string   `json:"full_name,omitempty"`
	Roles    []string `json:"roles"`
	Exp      int64    `json:"exp"`
}

type LoginResponse struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
	User        User      `json:"user"`
}
