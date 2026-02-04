package hashing

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"math/big"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stivo-m/api.kodiflow.com/pkg/config"
)

const (
	RefreshTokenBytes = 32
)

var AccessTokenExpiry = time.Now().Add(time.Minute * 15) // 15 minutes
var RefreshTokenTtl = time.Hour * 24                     // 1 day

// Generate otp code

func GenerateOTP(length int) (string, error) {
	const digits = "0123456789"
	bytes := make([]byte, length)

	for i := range bytes {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		bytes[i] = digits[n.Int64()]
	}

	return string(bytes), nil
}

// Generates a new jwt token
func CreateJwtToken(userId uuid.UUID) (string, error) {
	secretKey := []byte(config.MustGetEnv("JWT_SECRET_KEY"))

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"sub": userId.String(),
			"exp": AccessTokenExpiry.Unix(),
		})

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken parses and validates the JWT, returning the user ID
func ValidateToken(tokenStr string) (string, error) {
	secretKey := []byte(config.MustGetEnv("JWT_SECRET_KEY"))
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	if err != nil {
		return "", fmt.Errorf("invalid token: %w", err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		sub, ok := claims["sub"].(string)
		if !ok {
			return "", fmt.Errorf("user ID (sub) claim not found")
		}
		return sub, nil
	}

	return "", fmt.Errorf("invalid token claims")
}

// Generate a base64url token for the client
func NewRawToken() (string, error) {
	b := make([]byte, RefreshTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// HMAC-SHA256(token, key) -> hex or base64
func HashToken(token string, key []byte) string {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(token))
	return fmt.Sprintf("%x", mac.Sum(nil))
}

// GenerateInviteToken returns a cryptographically secure, URL-safe random token.
// It never returns an error. If random generation fails (very rare), it returns an empty string.
func GenerateInviteToken(length int) string {
	if length <= 0 {
		length = 32
	}

	byteLen := (length * 3) / 4
	if byteLen == 0 {
		byteLen = 24
	}

	b := make([]byte, byteLen)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}

	token := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(b)

	if len(token) > length {
		token = token[:length]
	}

	return token
}

// Generates a refresh token that expires based on the given ttl
type RefreshToken struct {
	RawToken    string
	HashedToken string
	ExpiresAt   time.Time
}

func GenerateRefreshToken(ttl time.Duration) (*RefreshToken, error) {
	token, err := NewRawToken()
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	hmacKey := config.MustGetEnv("APP_ENCRYPTION_KEY")
	tokenHash := HashToken(token, []byte(hmacKey))
	expiresAt := time.Now().Add(ttl)

	return &RefreshToken{
		RawToken:    token,
		HashedToken: tokenHash,
		ExpiresAt:   expiresAt,
	}, nil
}

func HashRefreshToken(token string) string {

	hmacKey := config.MustGetEnv("APP_ENCRYPTION_KEY")
	hashed := HashToken(token, []byte(hmacKey))
	return hashed
}
