package middleware

import (
	"net/http"
	"strings"

	"business_process_efficiency/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTMiddleware(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing or invalid Authorization header"})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		_, claims, err := auth.ParseToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user_claims", claims)

		if uid, ok := claims["user_id"].(float64); ok {
			c.Set("user_id", uint(uid))
		}

		c.Next()
	}
}

func RequireAccess(auth *service.AuthService, codes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claimsVal, exists := c.Get("user_claims")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "No claims"})
			c.Abort()
			return
		}

		claims, ok := claimsVal.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid claims"})
			c.Abort()
			return
		}

		for _, code := range codes {
			if auth.HasAccess(claims, code) {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		c.Abort()
	}
}
