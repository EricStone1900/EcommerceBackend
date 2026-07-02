package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/EricStone1900/ecommerce-backend/internal/domain/entity"
	bizerr "github.com/EricStone1900/ecommerce-backend/pkg/errors"
)

// Claims extends jwt.RegisteredClaims with application-specific fields.
type Claims struct {
	UserID uint        `json:"user_id"`
	Role   entity.Role `json:"role"`
	Type   string      `json:"type"`
	jwt.RegisteredClaims
}

const (
	tokenTypeAccess  = "access"
	tokenTypeRefresh = "refresh"
)

// JWTService handles JWT token generation and validation.
type JWTService struct {
	secret        []byte
	accessExpire  time.Duration
	refreshExpire time.Duration
	issuer        string
}

// NewJWTService creates a new JWTService with the given configuration.
func NewJWTService(secret string, accessExpire, refreshExpire time.Duration) *JWTService {
	return &JWTService{
		secret:        []byte(secret),
		accessExpire:  accessExpire,
		refreshExpire: refreshExpire,
		issuer:        "ecommerce-backend",
	}
}

// GenerateAccessToken creates a short-lived access token containing userID and role.
func (s *JWTService) GenerateAccessToken(userID uint, role entity.Role) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Role:   role,
		Type:   tokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.accessExpire)),
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secret)
}

// GenerateRefreshToken creates a long-lived refresh token with a unique jti for revocation.
// Returns the token string and the tokenID (jti) separately so the caller can store it.
func (s *JWTService) GenerateRefreshToken(userID uint, role entity.Role) (tokenString string, tokenID string, err error) {
	tokenID = uuid.New().String()
	now := time.Now()
	claims := &Claims{
		UserID: userID,
		Role:   role,
		Type:   tokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(s.refreshExpire)),
			Subject:   fmt.Sprintf("%d", userID),
			ID:        tokenID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString(s.secret)
	if err != nil {
		return "", "", fmt.Errorf("failed to sign refresh token: %w", err)
	}
	return tokenString, tokenID, nil
}

// ValidateToken parses and validates a JWT token string.
// Returns the claims on success, or one of: ErrUnauthorized, ErrTokenExpired.
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secret, nil
	})

	if err != nil {
		// Check if the error is due to token expiration
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, bizerr.ErrTokenExpired
		}
		return nil, bizerr.ErrUnauthorized
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, bizerr.ErrUnauthorized
	}

	return claims, nil
}

// IsAccessToken checks whether the claims represent an access token.
func (s *JWTService) IsAccessToken(claims *Claims) bool {
	return claims.Type == tokenTypeAccess
}

// IsRefreshToken checks whether the claims represent a refresh token.
func (s *JWTService) IsRefreshToken(claims *Claims) bool {
	return claims.Type == tokenTypeRefresh
}
