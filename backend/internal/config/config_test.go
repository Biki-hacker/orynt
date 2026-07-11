package config

import (
	"os"
	"testing"
)

func TestLoadConfig_Defaults(t *testing.T) {
	// Clear relevant env variables to test defaults
	os.Unsetenv("PORT")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("GEMINI_API_KEY")
	os.Unsetenv("USE_FIRESTORE")
	os.Unsetenv("FIREBASE_PROJECT_ID")
	os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
	os.Unsetenv("USE_REDIS")
	os.Unsetenv("REDIS_URL")
	os.Unsetenv("REDIS_ADDR")

	cfg := LoadConfig()

	if cfg.Port != "8080" {
		t.Errorf("Expected default Port to be 8080, got %s", cfg.Port)
	}

	expectedSecret := "orynt-default-secret-hackathon-2026-key"
	if cfg.JWTSecret != expectedSecret {
		t.Errorf("Expected default JWTSecret to be %s, got %s", expectedSecret, cfg.JWTSecret)
	}

	if cfg.UseFirestore != false {
		t.Errorf("Expected default UseFirestore to be false, got %t", cfg.UseFirestore)
	}

	if cfg.UseRedis != false {
		t.Errorf("Expected default UseRedis to be false, got %t", cfg.UseRedis)
	}

	if cfg.RedisURL != "redis://localhost:6379" {
		t.Errorf("Expected default RedisURL to be redis://localhost:6379, got %s", cfg.RedisURL)
	}
}

func TestLoadConfig_CustomEnv(t *testing.T) {
	// Set custom env variables
	os.Setenv("PORT", "9090")
	os.Setenv("JWT_SECRET", "custom-secret")
	os.Setenv("GEMINI_API_KEY", "custom-gemini-key")
	os.Setenv("USE_FIRESTORE", "true")
	os.Setenv("FIREBASE_PROJECT_ID", "test-project-id")
	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/path/to/creds.json")
	os.Setenv("USE_REDIS", "1")
	os.Setenv("REDIS_URL", "redis://redis-server:6379")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("GEMINI_API_KEY")
		os.Unsetenv("USE_FIRESTORE")
		os.Unsetenv("FIREBASE_PROJECT_ID")
		os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		os.Unsetenv("USE_REDIS")
		os.Unsetenv("REDIS_URL")
	}()

	cfg := LoadConfig()

	if cfg.Port != "9090" {
		t.Errorf("Expected Port to be 9090, got %s", cfg.Port)
	}

	if cfg.JWTSecret != "custom-secret" {
		t.Errorf("Expected JWTSecret to be custom-secret, got %s", cfg.JWTSecret)
	}

	if cfg.GeminiAPIKey != "custom-gemini-key" {
		t.Errorf("Expected GeminiAPIKey to be custom-gemini-key, got %s", cfg.GeminiAPIKey)
	}

	if cfg.UseFirestore != true {
		t.Errorf("Expected UseFirestore to be true, got %t", cfg.UseFirestore)
	}

	if cfg.FirebaseProjectID != "test-project-id" {
		t.Errorf("Expected FirebaseProjectID to be test-project-id, got %s", cfg.FirebaseProjectID)
	}

	if cfg.FirebaseCredentialsFile != "/path/to/creds.json" {
		t.Errorf("Expected FirebaseCredentialsFile to be /path/to/creds.json, got %s", cfg.FirebaseCredentialsFile)
	}

	if cfg.UseRedis != true {
		t.Errorf("Expected UseRedis to be true, got %t", cfg.UseRedis)
	}

	if cfg.RedisURL != "redis://redis-server:6379" {
		t.Errorf("Expected RedisURL to be redis://redis-server:6379, got %s", cfg.RedisURL)
	}
}
