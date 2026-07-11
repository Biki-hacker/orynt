package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"orynt/internal/repository"
	"orynt/internal/service"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupTestServer() (*gin.Engine, service.AuthService) {
	// Initialize repositories in-memory
	dbRepo := repository.NewFirestoreRepository(false, "", "")
	pubSubRepo := repository.NewRedisRepository(false, "")

	// Initialize services
	authSvc := service.NewAuthService(dbRepo, "test-jwt-secret-key-12345")
	tournSvc := service.NewTournamentService(dbRepo, pubSubRepo)
	stadSvc := service.NewStadiumService(dbRepo, pubSubRepo)
	opsSvc := service.NewOperationsService(dbRepo, pubSubRepo)
	aiSvc := service.NewAIService(dbRepo, pubSubRepo)

	// Initialize WebSockets Hub (required for router)
	wsHub := NewWSHub()
	go wsHub.Run()

	// Connect PubSub events to WebSockets hub
	StartPubSubListener(pubSubRepo, wsHub)

	// Setup Gin handler delivery layer
	handler := NewAPIHandler(authSvc, tournSvc, stadSvc, opsSvc, aiSvc)
	router := SetupRouter(handler, wsHub, pubSubRepo)

	return router, authSvc
}

func TestHealthCheckHandler(t *testing.T) {
	router, _ := setupTestServer()

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected health status 200, got %d", w.Code)
	}

	var resp map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "UP" {
		t.Errorf("Expected status 'UP', got %s", resp["status"])
	}
}

func TestLoginHandler(t *testing.T) {
	router, _ := setupTestServer()

	// Successful login
	loginData := map[string]string{
		"username": "admin",
		"password": "admin123",
	}
	body, _ := json.Marshal(loginData)

	req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected login status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	var tokenResp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &tokenResp)
	if tokenResp["accessToken"] == "" {
		t.Error("Expected accessToken in login response")
	}

	// Failed login
	loginData = map[string]string{
		"username": "admin",
		"password": "wrong_password",
	}
	body, _ = json.Marshal(loginData)

	req = httptest.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected failed login status 401, got %d", w.Code)
	}
}

func TestGetStadiumHandler(t *testing.T) {
	router, _ := setupTestServer()

	req := httptest.NewRequest("GET", "/api/stadium", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var stadium map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &stadium)
	if stadium["name"] != "Orynt Arena" {
		t.Errorf("Expected stadium name 'Orynt Arena', got %v", stadium["name"])
	}
}

func TestGetParkingHandler(t *testing.T) {
	router, _ := setupTestServer()

	req := httptest.NewRequest("GET", "/api/parking", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var list []interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) == 0 {
		t.Error("Expected seeded parking zones list")
	}
}

func TestGetTransportHandler(t *testing.T) {
	router, _ := setupTestServer()

	req := httptest.NewRequest("GET", "/api/transport", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var list []interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) == 0 {
		t.Error("Expected seeded transport lines list")
	}
}

func TestLostFoundHandlers(t *testing.T) {
	router, _ := setupTestServer()

	// POST Report
	itemData := map[string]string{
		"itemName":     "Keys",
		"description":  "Silver keychain",
		"category":     "Personal Items",
		"contactName":  "Jack",
		"contactPhone": "555-555",
	}
	body, _ := json.Marshal(itemData)

	req := httptest.NewRequest("POST", "/api/lost-found", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d", w.Code)
	}

	// GET List
	req = httptest.NewRequest("GET", "/api/lost-found", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var list []interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &list)
	if len(list) == 0 {
		t.Error("Expected reported item to be in list")
	}
}

func TestAIChatHandler(t *testing.T) {
	router, _ := setupTestServer()

	chatQuery := map[string]interface{}{
		"message": "Where is Parking Lot A?",
		"history": []interface{}{},
	}
	body, _ := json.Marshal(chatQuery)

	req := httptest.NewRequest("POST", "/api/ai/chat", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := newCloseNotifyingRecorder()
	router.ServeHTTP(w, req)

	// In test mode, SSE response should stream back status 200
	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	// Verify headers for SSE (Server-Sent Events)
	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Errorf("Expected content-type text/event-stream, got %s", w.Header().Get("Content-Type"))
	}
	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Errorf("Expected cache-control no-cache, got %s", w.Header().Get("Cache-Control"))
	}
}

type closeNotifyingRecorder struct {
	*httptest.ResponseRecorder
	closed chan bool
}

func (c *closeNotifyingRecorder) CloseNotify() <-chan bool {
	return c.closed
}

func newCloseNotifyingRecorder() *closeNotifyingRecorder {
	return &closeNotifyingRecorder{
		ResponseRecorder: httptest.NewRecorder(),
		closed:           make(chan bool, 1),
	}
}
