package middleware

import (
	// "log"
	"log"
	"net/http"
	"strings"
	// "time"

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
        log.Println("🔹 AuthMiddleware started")

        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            log.Println("❌ Authorization header missing")
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization header missing"})
            return
        }
        log.Println("✅ Authorization header found")

        parts := strings.SplitN(authHeader, " ", 2)
        if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
            log.Printf("❌ Invalid auth header format: %v\n", authHeader)
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
            return
        }

        tokenString := parts[1]
        token, err := service.VerifyToken(tokenString, false)
        if err != nil {
            log.Printf("❌ Token verification failed: %v\n", err)
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
            return
        }
        log.Println("✅ Token verified successfully")

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok || !token.Valid {
            log.Println("❌ Invalid token claims")
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
            return
        }
        log.Printf("✅ Token claims: %+v\n", claims)

        subStr, ok := claims["sub"].(string)
        if !ok {
            log.Println("❌ Missing sub claim")
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing sub claim"})
            return
        }

        userUUID, err := uuid.Parse(subStr)
        if err != nil {
            log.Printf("❌ Invalid user UUID: %v\n", err)
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid user ID"})
            return
        }
        log.Printf("✅ User UUID: %v\n", userUUID)

        user, err := db.Q.GetUserByID(c.Request.Context(), pgtype.UUID{Bytes: userUUID, Valid: true})
        if err != nil {
            log.Printf("❌ User not found: %v\n", err)
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not found"})
            return
        }
        log.Printf("✅ User fetched: %s %s\n", user.FirstName, user.LastName)

        tokenVersion, _ := claims["token_version"].(float64)
        if int32(tokenVersion) != user.TokenVersion {
            log.Printf("❌ Token version mismatch: token=%v, user=%v\n", int32(tokenVersion), user.TokenVersion)
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token has been revoked"})
            return
        }
        log.Println("✅ Token version validated")

        // Handle banned users
        if user.IsBanned.Bool {
            log.Println("⚠️ User is banned")
            if !strings.HasPrefix(c.FullPath(), "/users/user/") {
                log.Printf("❌ Banned user tried to access: %s\n", c.FullPath())
                c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
                    "error":  "your account is banned",
                    "reason": user.BanReason.String,
                })
                return
            }
            log.Println("✅ Banned user accessing allowed route /users/user/:id")
            c.Set("banned_user", true)
        }

        // Set context for downstream handlers
        c.Set("userID", userUUID)
        c.Set("role", user.Role.String)
        c.Set("token_version", int(tokenVersion))
        c.Set("isBanned", user.IsBanned.Bool)
        c.Set("isPermanentBan", user.IsPermanentBan.Bool)
        c.Set("banReason", user.BanReason.String)
        c.Set("banUntil", user.BanUntil.Time)

        log.Println("✅ Context set for downstream handlers")
        c.Next()
        log.Println("🔹 AuthMiddleware finished")
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
