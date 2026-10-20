package auth

import (
	"errors"
	"time"

	"klyniva/internal/config"
	"klyniva/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken   = errors.New("invalid token")
	ErrExpiredToken   = errors.New("token has expired")
	ErrWrongTokenType = errors.New("wrong token type")
)

// Token types stored in the "typ" claim. Keeping them distinct stops a
// long-lived refresh token from being accepted as an access token.
const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type JWTManager struct {
	secretKey     string
	tokenExpiry   time.Duration
	refreshExpiry time.Duration
}

func NewJWTManager(cfg *config.Config) *JWTManager {
	return &JWTManager{
		secretKey:     cfg.JWT.Secret,
		tokenExpiry:   cfg.JWT.Expiry,
		refreshExpiry: cfg.JWT.RefreshExpiry,
	}
}

// GenerateTokenPair generates both access and refresh tokens
func (j *JWTManager) GenerateTokenPair(user *models.User) (*models.LoginResponse, error) {
	// Generate access token
	accessToken, err := j.generateToken(user, TokenTypeAccess, j.tokenExpiry)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := j.generateToken(user, TokenTypeRefresh, j.refreshExpiry)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

// generateToken creates a JWT token of the given type with the given expiry
func (j *JWTManager) generateToken(user *models.User, tokenType string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := &models.TokenClaims{
		UserID:   user.ID,
		Username: user.Username,
		UserType: user.UserType,
		Exp:      now.Add(expiry).Unix(),
		Iat:      now.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":   claims.UserID,
		"username":  claims.Username,
		"user_type": claims.UserType,
		"typ":       tokenType,
		"exp":       claims.Exp,
		"iat":       claims.Iat,
	})

	return token.SignedString([]byte(j.secretKey))
}

// ValidateToken validates an access token and returns the claims.
// Refresh tokens are rejected so they cannot be used to call the API.
func (j *JWTManager) ValidateToken(tokenString string) (*models.TokenClaims, error) {
	return j.validate(tokenString, TokenTypeAccess)
}

// ValidateRefreshToken validates a refresh token and returns the claims.
func (j *JWTManager) ValidateRefreshToken(tokenString string) (*models.TokenClaims, error) {
	return j.validate(tokenString, TokenTypeRefresh)
}

// validate parses the token, checks the signature, expiry and token type.
func (j *JWTManager) validate(tokenString, wantType string) (*models.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Check expiration
	if exp, ok := claims["exp"].(float64); ok {
		if time.Now().Unix() > int64(exp) {
			return nil, ErrExpiredToken
		}
	}

	if typ, _ := claims["typ"].(string); typ != wantType {
		return nil, ErrWrongTokenType
	}

	// Extract claims
	userID, ok := claims["user_id"].(float64)
	if !ok {
		return nil, ErrInvalidToken
	}

	username, ok := claims["username"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	userType, ok := claims["user_type"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	return &models.TokenClaims{
		UserID:   uint(userID),
		Username: username,
		UserType: models.UserRole(userType),
		Exp:      int64(claims["exp"].(float64)),
		Iat:      int64(claims["iat"].(float64)),
	}, nil
}
