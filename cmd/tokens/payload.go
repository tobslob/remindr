package tokens

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	uuid "github.com/google/uuid"
)

var (
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidToken = errors.New("token is invalid")
)

type Payload struct {
	UserID uuid.UUID `json:"user_id"`
	jwt.RegisteredClaims
}

func NewPayload(userId uuid.UUID, duration time.Duration) (*Payload, error) {
	sessionId, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	now := time.Now()

	payload := &Payload{
		UserID: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        sessionId.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(duration)),
		},
	}

	return payload, nil
}
