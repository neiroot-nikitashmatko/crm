package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"proclients/backend/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UserID string `json:"sub"`
	Phone  string `json:"phone"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret []byte
	ttl    time.Duration
}

func NewManager(secret string, ttl time.Duration) (*Manager, error) {
	trimmed := strings.TrimSpace(secret)
	if len(trimmed) < 32 {
		return nil, errors.New("JWT_SECRET must be at least 32 characters")
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &Manager{
		secret: []byte(trimmed),
		ttl:    ttl,
	}, nil
}

func (m *Manager) Issue(user *model.AuthUser) (string, error) {
	if user == nil || strings.TrimSpace(user.ID) == "" {
		return "", errors.New("user id is required")
	}

	now := time.Now()
	claims := Claims{
		UserID: user.ID,
		Phone:  user.Phone,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign token: %w", err)
	}
	return signed, nil
}

func (m *Manager) Parse(tokenString string) (*Claims, error) {
	tokenString = strings.TrimSpace(tokenString)
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	parsed, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := parsed.Claims.(*Claims)
	if !ok || !parsed.Valid || strings.TrimSpace(claims.UserID) == "" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
