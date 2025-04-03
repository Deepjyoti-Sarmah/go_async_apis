package store

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type RefreshTokenStore struct {
	db *sqlx.DB
}

func NewRefreshTokenStore(db *sql.DB) *RefreshTokenStore {
	return &RefreshTokenStore{
		db: sqlx.NewDb(db, "postgres"),
	}
}

type RefreshToken struct {
	UserId      uuid.UUID `db:"user_id"`
	HashedToken string    `db:"hashed_token"`
	CreatedAt   time.Time `db:"created_at"`
	ExpiredAt   time.Time `db:"expired_at"`
}

func (s *RefreshTokenStore) Create(ctx context.Context, userId uuid.UUID, token *jwt.Token) (*RefreshToken, error) {
	const insert = `INSERT INTO refresh_token (user_id, hashed_token, expired_at) VALUES ($1, $2, $3)`
	hashedToken, err := bcrypt.GenerateFromPassword([]byte(token.Raw), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("bcrypting the refresh token failed: %w", err)
	}

	base64TokenHash := base64.StdEncoding.EncodeToString(hashedToken)

	var refreshToken RefreshToken
	if err := s.db.GetContext(ctx, &refreshToken, insert, userId, base64TokenHash); err != nil {
		return nil, fmt.Errorf("failed to create refresh token record: %w", err)
	}

	return &refreshToken, nil
}
