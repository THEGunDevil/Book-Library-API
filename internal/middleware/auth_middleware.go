package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	"github.com/THEGunDevil/GoForBackend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header missing"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		tokenString := parts[1]
		token, err := service.VerifyToken(tokenString, false)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		subStr, ok := claims["sub"].(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing sub claim"})
			return
		}

		userUUID, err := uuid.Parse(subStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
			return
		}

		user, err := db.Q.GetUserByID(c.Request.Context(), pgtype.UUID{Bytes: userUUID, Valid: true})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		tokenVersion, _ := claims["token_version"].(float64)
		if int32(tokenVersion) != user.TokenVersion {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
			return
		}

		if user.IsBanned.Bool {
			if user.IsPermanentBan.Bool {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "your account has been permanently banned", "reason": user.BanReason.String})
				return
			}
			if user.BanUntil.Valid && user.BanUntil.Time.After(time.Now()) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "your account is temporarily banned", "until": user.BanUntil.Time, "reason": user.BanReason.String})
				return
			}
		}

		role, _ := claims["role"].(string)
		c.Set("userID", userUUID)
		c.Set("role", role)
		c.Set("token_version", int(tokenVersion))
		c.Next()
	}
}

// AdminOnly ensures the request is from an admin
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}
		c.Next()
	}
}

// CORSMiddleware configures CORS headers
// func CORSMiddleware(allowedOrigins ...string) gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		origin := c.GetHeader("Origin")
// 		allowOrigin := ""

// 		// Match only allowed origins
// 		for _, o := range allowedOrigins {
// 			if origin == o || (o == "http://localhost:3000" && strings.HasPrefix(origin, "http://localhost:")) {
// 				allowOrigin = origin
// 				break
// 			}
// 		}

// 		if allowOrigin == "" {
// 			// If no matching origin, proceed without setting CORS headers
// 			// This avoids breaking non-CORS requests
// 			c.Next()
// 			return
// 		}

// 		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
// 		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
// 		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Accept")
// 		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
// 		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 24 hours

// 		if c.Request.Method == "OPTIONS" {
// 			c.AbortWithStatus(http.StatusNoContent)
// 			return
// 		}

// 		c.Next()
// 	}
// }