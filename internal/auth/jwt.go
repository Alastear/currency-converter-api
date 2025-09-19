package auth

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID    uint `json:"uid"`
	SessionID uint `json:"sid"`
	jwt.RegisteredClaims
}

func MakeToken(secret string, userID uint, sessionID uint, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID:    userID,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(tokenStr, secret string) (*Claims, error) {
	tk, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil {
		return nil, err
	}
	claims, ok := tk.Claims.(*Claims)
	if !ok || !tk.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}

func BearerToken(c *gin.Context) (string, error) {
	h := c.GetHeader("Authorization")
	if !strings.HasPrefix(strings.ToLower(h), "bearer ") {
		return "", errors.New("missing bearer token")
	}
	return strings.TrimSpace(h[7:]), nil
}

func AuthMiddleware(secret string, sessionChecker func(userID, sessionID uint) bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		tok, err := BearerToken(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		claims, err := ParseToken(tok, secret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
			return
		}
		if !sessionChecker(claims.UserID, claims.SessionID) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session_revoked"})
			return
		}
		c.Set("uid", claims.UserID)
		c.Set("sid", claims.SessionID)
		c.Next()
	}
}
