package auth

import (
	"errors"
	"time"

	"care-backend/internal/config"
	"care-backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
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
	accessToken, err := j.generateToken(user, j.tokenExpiry)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := j.generateToken(user, j.refreshExpiry)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Access:  accessToken,
		Refresh: refreshToken,
	}, nil
}

// generateToken creates a JWT token with the given expiry
func (j *JWTManager) generateToken(user *models.User, expiry time.Duration) (string, error) {
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
		"exp":       claims.Exp,
		"iat":       claims.Iat,
	})

	return token.SignedString([]byte(j.secretKey))
}

// ValidateToken validates a JWT token and returns the claims
func (j *JWTManager) ValidateToken(tokenString string) (*models.TokenClaims, error) {
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

// RefreshToken generates a new access token from a valid refresh token
func (j *JWTManager) RefreshToken(refreshTokenString string) (*models.LoginResponse, error) {
	claims, err := j.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, err
	}

	// Create a temporary user object for token generation
	user := &models.User{
		ID:       claims.UserID,
		Username: claims.Username,
		UserType: claims.UserType,
	}

	// Generate new token pair
	return j.GenerateTokenPair(user)
}