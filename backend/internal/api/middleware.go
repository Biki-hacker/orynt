package api

import (
	"net/http"
	"orynt/internal/models"
	"orynt/internal/repository"
	"orynt/internal/service"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware handles CORS requests
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// AuthMiddleware parses JWT tokens and authenticates requests
func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header must be Bearer token"})
			c.Abort()
			return
		}

		tokenStr := parts[1]
		user, err := authService.ValidateToken(tokenStr)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session token"})
			c.Abort()
			return
		}

		// Store user context
		c.Set("currentUser", user)
		c.Next()
	}
}

// RBACMiddleware checks if user has one of the allowed roles
func RBACMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, exists := c.Get("currentUser")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User context not established"})
			c.Abort()
			return
		}

		// Import user models to unpack context
		userTypeCheck(userVal, func(role string) {
			hasRole := false
			for _, r := range allowedRoles {
				if r == role || role == "admin" { // admin bypasses all role constraints
					hasRole = true
					break
				}
			}

			if !hasRole {
				c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: insufficient permissions"})
				c.Abort()
				return
			}
		}, func() {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user model in context"})
			c.Abort()
		})

		c.Next()
	}
}

// Helper to check user type and role without import loops
func userTypeCheck(userVal interface{}, success func(role string), fail func()) {
	if u, ok := userVal.(*models.User); ok {
		success(u.Role)
	} else {
		fail()
	}
}

// RateLimiter represents a thread-safe sliding rate limiter (Simple Token Bucket)
type RateLimiter struct {
	mu           sync.Mutex
	ips          map[string][]time.Time
	limitSeconds int
	maxRequests  int
}

// RateLimitMiddleware blocks spamming IPs, using Redis when active or falling back to in-memory sliding window
func RateLimitMiddleware(pubSub repository.PubSubRepository) gin.HandlerFunc {
	limiter := &RateLimiter{
		ips:          make(map[string][]time.Time),
		limitSeconds: 60,
		maxRequests:  120, // 120 requests per minute
	}

	return func(c *gin.Context) {
		ip := c.ClientIP()
		key := "rate:ip:" + ip
		now := time.Now()

		// Try Redis rate limiting first
		count, err := pubSub.IncrLimit(c.Request.Context(), key, time.Duration(limiter.limitSeconds)*time.Second)
		if err == nil && count > 0 {
			if count > int64(limiter.maxRequests) {
				c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Please slow down."})
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// Fallback to in-memory sliding window limiter
		limiter.mu.Lock()
		times, exists := limiter.ips[ip]
		validTimes := make([]time.Time, 0)
		if exists {
			for _, t := range times {
				if now.Sub(t) < time.Duration(limiter.limitSeconds)*time.Second {
					validTimes = append(validTimes, t)
				}
			}
		}

		if len(validTimes) >= limiter.maxRequests {
			limiter.mu.Unlock()
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests. Please slow down."})
			c.Abort()
			return
		}

		limiter.ips[ip] = append(validTimes, now)
		limiter.mu.Unlock()
		c.Next()
	}
}
