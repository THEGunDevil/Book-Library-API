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

// AuthMiddleware validates JWT and extracts user info.
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

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unexpected signing method"})
				return nil, nil
			}
			return service.JwtSecret, nil
		})
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
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

		// 🔒 Check if user is banned
		user, err := db.Q.GetUserByID(c.Request.Context(), pgtype.UUID{Bytes: userUUID, Valid: true})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
			return
		}

		// If banned permanently or temporarily active
		if user.IsBanned.Bool {
			// If permanent ban
			if user.IsPermanentBan.Bool {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error":  "your account has been permanently banned",
					"reason": user.BanReason.String,
				})
				return
			}

			// If temporary ban still active
			if user.BanUntil.Valid && user.BanUntil.Time.After(time.Now()) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error":  "your account is temporarily banned",
					"until":  user.BanUntil.Time,
					"reason": user.BanReason.String,
				})
				return
			}
		}

		// Ban expired? automatically treat as unbanned (optional)
		// if user.BanUntil.Valid && user.BanUntil.Time.Before(time.Now()) {
		//     // You could trigger background unban logic if desired
		// }

		role, _ := claims["role"].(string)
		tokenVersion, _ := claims["token_version"].(float64)

		c.Set("userID", userUUID)
		c.Set("role", role)
		c.Set("token_version", int(tokenVersion))
		c.Next()
	}
}

// AdminOnly ensures the request is from an admin user.
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

// CORSMiddleware configures CORS headers.
func CORSMiddleware(allowedOrigins ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowOrigin := "*"

		// If specific origins are defined, allow only those
		if len(allowedOrigins) > 0 {
			for _, o := range allowedOrigins {
				if origin == o {
					allowOrigin = o
					break
				}
			}
			if allowOrigin != origin && origin != "" {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "CORS origin not allowed"})
				return
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigin)
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
