package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidToken = errors.New("invalid token")

type Service struct {
	tokenSecret []byte
	tokenTTL    time.Duration
}

func NewService(tokenSecret string, tokenTTL time.Duration) *Service {
	return &Service{
		tokenSecret: []byte(tokenSecret),
		tokenTTL:    tokenTTL,
	}
}

func (s *Service) VerifyPassword(password, passwordHash string) error {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
}

func (s *Service) GenerateToken(user User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.tokenTTL)
	claims := TokenClaims{
		Sub:      user.ID,
		Username: user.Username,
		Email:    user.Email,
		FullName: user.FullName,
		Roles:    user.RoleCodes,
		Exp:      expiresAt.Unix(),
	}

	payload, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("marshal auth token claims: %w", err)
	}

	payloadEncoded := base64.RawURLEncoding.EncodeToString(payload)
	signatureEncoded := s.sign(payloadEncoded)

	return payloadEncoded + "." + signatureEncoded, expiresAt, nil
}

func (s *Service) ParseToken(token string) (*TokenClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return nil, ErrInvalidToken
	}

	expectedSignature := s.sign(parts[0])
	if !hmac.Equal([]byte(expectedSignature), []byte(parts[1])) {
		return nil, ErrInvalidToken
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}

	var claims TokenClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, ErrInvalidToken
	}

	if claims.Exp <= time.Now().Unix() {
		return nil, ErrInvalidToken
	}

	return &claims, nil
}

func (s *Service) sign(payload string) string {
	mac := hmac.New(sha256.New, s.tokenSecret)
	_, _ = mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
