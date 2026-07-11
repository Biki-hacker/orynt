package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"orynt/internal/models"
	"orynt/internal/repository"
	"orynt/internal/service"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCORSMiddleware(t *testing.T) {
	r := gin.New()
	r.Use(CORSMiddleware())
	r.GET("/test", func(c *gin.Context) {
		c.String(200, "ok")
	})

	// Test OPTIONS preflight
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected CORS origin *, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}

	// Test GET request
	req = httptest.NewRequest("GET", "/test", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Expected CORS origin *, got %s", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestAuthMiddleware(t *testing.T) {
	dbRepo := repository.NewFirestoreRepository(false, "", "")
	secret := "test-secret-key-12345"
	authSvc := service.NewAuthService(dbRepo, secret)

	r := gin.New()
	r.Use(AuthMiddleware(authSvc))
	r.GET("/protected", func(c *gin.Context) {
		user, exists := c.Get("currentUser")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "user not found in context"})
			return
		}
		c.JSON(http.StatusOK, user)
	})

	// Case 1: Missing Header
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for missing header, got %d", w.Code)
	}

	// Case 2: Invalid Header Format
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat tokenhere")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid format, got %d", w.Code)
	}

	// Case 3: Invalid Token
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid token, got %d", w.Code)
	}

	// Case 4: Valid Token (Volunteer)
	resp, _ := authSvc.Login(context.Background(), "volunteer1", "staff123")
	req = httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+resp.AccessToken)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200 for valid token, got %d", w.Code)
	}

	var user models.User
	_ = json.Unmarshal(w.Body.Bytes(), &user)
	if user.Username != "volunteer1" {
		t.Errorf("Expected volunteer1, got %s", user.Username)
	}
}

func TestRBACMiddleware(t *testing.T) {
	r := gin.New()
	r.GET("/staff-only", func(c *gin.Context) {
		// Mock AuthMiddleware setting user
		user := &models.User{
			Username: "volunteer1",
			Role:     "volunteer",
		}
		c.Set("currentUser", user)
		c.Next()
	}, RBACMiddleware("volunteer", "ops"), func(c *gin.Context) {
		c.String(200, "success")
	})

	r.GET("/admin-only", func(c *gin.Context) {
		user := &models.User{
			Username: "volunteer1",
			Role:     "volunteer",
		}
		c.Set("currentUser", user)
		c.Next()
	}, RBACMiddleware("admin"), func(c *gin.Context) {
		c.String(200, "success")
	})

	r.GET("/admin-bypass", func(c *gin.Context) {
		// Admin user calling something permitted for only "ops"
		user := &models.User{
			Username: "admin",
			Role:     "admin",
		}
		c.Set("currentUser", user)
		c.Next()
	}, RBACMiddleware("ops"), func(c *gin.Context) {
		c.String(200, "success")
	})

	// Case 1: Allowed role
	req := httptest.NewRequest("GET", "/staff-only", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	// Case 2: Forbidden role
	req = httptest.NewRequest("GET", "/admin-only", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected 403, got %d", w.Code)
	}

	// Case 3: Admin bypasses role restrictions
	req = httptest.NewRequest("GET", "/admin-bypass", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("Expected admin to bypass role restrictions (200), got %d", w.Code)
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	pubSub := repository.NewRedisRepository(false, "")
	r := gin.New()
	r.Use(RateLimitMiddleware(pubSub))
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	// Since maxRequests is 120, we can run a loop of 125 requests in test mode.
	// In-memory sliding window fallback should block the 121st request.
	for i := 0; i < 125; i++ {
		req := httptest.NewRequest("GET", "/ping", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if i < 120 {
			if w.Code != http.StatusOK {
				t.Fatalf("Request %d failed prematurely: expected 200, got %d", i+1, w.Code)
			}
		} else {
			if w.Code != http.StatusTooManyRequests {
				t.Errorf("Request %d should have been rate limited: expected 429, got %d", i+1, w.Code)
			}
		}
	}
}
