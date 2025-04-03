package apiserver

import (
	"fmt"
	"time"

	"github.com/Deepjyoti-Sarmah/fast-api/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type JwtManager struct {
	config *config.Config
}

func NextJwtManager(config *config.Config) *JwtManager {
	return &JwtManager{
		config: config,
	}
}

type TokenPair struct {
	AccessToken  *jwt.Token
	RefreshToken *jwt.Token
}

type CustomClaims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func (j *JwtManager) GenerateTokenPair(userId uuid.UUID) (*TokenPair, error) {
	now := time.Now()
	issuer := "http://" + j.config.ApiServerHost + ":" + j.config.ApiServerPort

	// AccessToken
	jwtAccessToken := jwt.NewWithClaims(jwt.SigningMethodES256,
		CustomClaims{
			TokenType: "access",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   userId.String(),
				Issuer:    issuer,
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Minute * 15)),
				IssuedAt:  jwt.NewNumericDate(now),
			},
		})

	key := []byte(j.config.JwtSecret)

	var err error
	jwtAccessToken.Raw, err = jwtAccessToken.SignedString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// RefreshToken
	jwtRefreshToken := jwt.NewWithClaims(jwt.SigningMethodES256,
		CustomClaims{
			TokenType: "refresh",
			RegisteredClaims: jwt.RegisteredClaims{
				Subject:   userId.String(),
				Issuer:    issuer,
				ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * 24 * 30)),
				IssuedAt:  jwt.NewNumericDate(now),
			},
		})

	jwtRefreshToken.Raw, err = jwtRefreshToken.SignedString(key)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  jwtAccessToken,
		RefreshToken: jwtRefreshToken,
	}, nil
}
