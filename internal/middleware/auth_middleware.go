package middleware

import (
	// "log"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/THEGunDevil/GoForBackend/internal/db"
	// gen "github.com/THEGunDevil/GoForBackend/internal/db/gen"
	"github.com/THEGunDevil/GoForBackend/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// AuthMiddleware validates JWT tokens and sets user context
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        log.Println("⏳ AuthMiddleware start")

        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            log.Println("❌ Authorization header missing")
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header missing"})
            return
        }

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
            log.Println("❌ Invalid authorization header format:", authHeader)
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
            return
        }

        tokenString := parts[1]
        token, err := service.VerifyToken(tokenString, false)
        if err != nil {
            log.Println("❌ Invalid or expired token:", err)
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok || !token.Valid {
            log.Println("❌ Invalid token claims")
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
            return
        }

        subStr, ok := claims["sub"].(string)
        if !ok {
            log.Println("❌ Missing sub claim")
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing sub claim"})
            return
        }

        userUUID, err := uuid.Parse(subStr)
        if err != nil {
            log.Println("❌ Invalid user ID:", subStr)
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
            return
        }

        user, err := db.Q.GetUserByID(c.Request.Context(), pgtype.UUID{Bytes: userUUID, Valid: true})
        if err != nil {
            log.Println("❌ User not found:", userUUID)
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
            return
        }

        log.Printf("✅ User found: %s, role: %s, banned: %v", userUUID, user.Role.String, user.IsBanned.Bool)

        // Token version validation
        tokenVersion, _ := claims["token_version"].(float64)
        if int32(tokenVersion) != user.TokenVersion {
            log.Printf("❌ Token revoked. Token version: %d, User version: %d", int(tokenVersion), user.TokenVersion)
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
            return
        }

        // Check banned status
        if user.IsBanned.Bool {
            log.Printf("⚠️ User is banned. Permanent: %v, Temporary until: %v", user.IsPermanentBan.Bool, user.BanUntil.Time)

            // Temporarily banned?
            if user.BanUntil.Valid && user.BanUntil.Time.After(time.Now()) {
                if strings.HasPrefix(c.FullPath(), "/users/user/") {
                    log.Println("✅ Temporarily banned user accessing allowed route:", c.FullPath())
                    c.Set("userID", userUUID)
                    c.Set("role", user.Role.String)
                    c.Set("token_version", int(tokenVersion))
                    c.Set("banned_user", true)
                    c.Next()
                    return
                }

                log.Println("❌ Temporarily banned user blocked from route:", c.FullPath())
                c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
                    "error":  "your account is temporarily banned",
                    "until":  user.BanUntil.Time,
                    "reason": user.BanReason.String,
                })
                return
            }

            // Permanently banned?
            if user.IsPermanentBan.Bool {
                if strings.HasPrefix(c.FullPath(), "/users/user/") {
                    log.Println("✅ Permanently banned user accessing allowed route:", c.FullPath())
                    c.Set("userID", userUUID)
                    c.Set("role", user.Role.String)
                    c.Set("token_version", int(tokenVersion))
                    c.Set("banned_user", true)
                    c.Next()
                    return
                }

                log.Println("❌ Permanently banned user blocked from route:", c.FullPath())
                c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
                    "error":  "your account has been permanently banned",
                    "reason": user.BanReason.String,
                })
                return
            }
        }

        // Default (non-banned user)
        log.Println("✅ Auth passed for non-banned user:", userUUID)
        c.Set("userID", userUUID)
        c.Set("role", user.Role.String)
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
