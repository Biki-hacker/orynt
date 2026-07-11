package service

import (
	"context"
	"orynt/internal/repository"
	"testing"
)

func TestAuthService_Login_Success(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewFirestoreRepository(false, "", "")
	secret := "test-jwt-secret-key-12345"
	svc := NewAuthService(repo, secret)

	resp, err := svc.Login(ctx, "admin", "admin123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if resp == nil {
		t.Fatal("Expected token response to be not nil")
	}

	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Errorf("Expected tokens to be populated, got access=%s, refresh=%s", resp.AccessToken, resp.RefreshToken)
	}

	if resp.User.Username != "admin" || resp.User.Role != "admin" {
		t.Errorf("Expected user to be admin, got username=%s, role=%s", resp.User.Username, resp.User.Role)
	}
}

func TestAuthService_Login_Failure(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewFirestoreRepository(false, "", "")
	secret := "test-jwt-secret-key-12345"
	svc := NewAuthService(repo, secret)

	// Wrong password
	_, err := svc.Login(ctx, "admin", "wrong_password")
	if err == nil {
		t.Error("Expected error for wrong password, got nil")
	}

	// Non-existent user
	_, err = svc.Login(ctx, "nonexistent", "password")
	if err == nil {
		t.Error("Expected error for non-existent user, got nil")
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewFirestoreRepository(false, "", "")
	secret := "test-jwt-secret-key-12345"
	svc := NewAuthService(repo, secret)

	resp, err := svc.Login(ctx, "volunteer1", "staff123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Validate valid token
	user, err := svc.ValidateToken(resp.AccessToken)
	if err != nil {
		t.Fatalf("Expected token to be valid, got error: %v", err)
	}

	if user.Username != "volunteer1" {
		t.Errorf("Expected user volunteer1, got %s", user.Username)
	}

	// Validate invalid token
	_, err = svc.ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("Expected validation error for invalid token string, got nil")
	}
}

func TestAuthService_CreateUser(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewFirestoreRepository(false, "", "")
	secret := "test-jwt-secret-key-12345"
	svc := NewAuthService(repo, secret)

	// Create user
	user, err := svc.CreateUser(ctx, "new_staff", "securepass", "security", "Gate Operations")
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}

	if user.Username != "new_staff" || user.Role != "security" || user.Department != "Gate Operations" {
		t.Errorf("Incorrect user fields: %+v", user)
	}

	// Try logging in with the newly created user
	resp, err := svc.Login(ctx, "new_staff", "securepass")
	if err != nil {
		t.Fatalf("Login failed for new user: %v", err)
	}

	if resp.User.Username != "new_staff" {
		t.Errorf("Expected new_staff, got %s", resp.User.Username)
	}

	// Create user with invalid fields
	_, err = svc.CreateUser(ctx, "", "securepass", "security", "")
	if err == nil {
		t.Error("Expected error for empty username, got nil")
	}
}

func TestAuthService_ListUsers(t *testing.T) {
	ctx := context.Background()
	repo := repository.NewFirestoreRepository(false, "", "")
	secret := "test-jwt-secret-key-12345"
	svc := NewAuthService(repo, secret)

	users, err := svc.ListUsers(ctx)
	if err != nil {
		t.Fatalf("Failed to list users: %v", err)
	}

	// By default seedData creates 6 users
	if len(users) < 6 {
		t.Errorf("Expected at least 6 users from seed data, got %d", len(users))
	}
}
