package service

import (
	"context"
	"errors"
	"time"
	"orynt/internal/models"
	"orynt/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	repo      repository.DBRepository
	jwtSecret []byte
}

// NewAuthService creates a concrete auth service
func NewAuthService(repo repository.DBRepository, secret string) AuthService {
	if secret == "" {
		secret = "orynt-default-secret-hackathon-2026-key"
	}
	return &authService{
		repo:      repo,
		jwtSecret: []byte(secret),
	}
}

// Claims defines JWT token payload
type Claims struct {
	UserID     string `json:"userId"`
	Username   string `json:"username"`
	Role       string `json:"role"`
	Department string `json:"department"`
	jwt.RegisteredClaims
}

func (s *authService) Login(ctx context.Context, username, password string) (*models.TokenResponse, error) {
	user, err := s.repo.GetUserByUsername(ctx, username)
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.New("invalid username or password")
	}

	// Generate Access Token (expires in 15 minutes for security, or 12 hours for ease of hackathon debugging)
	expirationTime := time.Now().Add(12 * time.Hour)
	claims := &Claims{
		UserID:     user.ID,
		Username:   user.Username,
		Role:       user.Role,
		Department: user.Department,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	// Refresh token
	refreshExpiration := time.Now().Add(7 * 24 * time.Hour)
	refClaims := &Claims{
		UserID: user.ID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(refreshExpiration),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refClaims)
	refreshTokenStr, err := refreshToken.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &models.TokenResponse{
		AccessToken:  tokenStr,
		RefreshToken: refreshTokenStr,
		User:         *user,
	}, nil
}

func (s *authService) ValidateToken(tokenStr string) (*models.User, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Double check user exists in repo
	user, err := s.repo.GetUser(context.Background(), claims.UserID)
	if err != nil {
		return nil, errors.New("user session not found")
	}

	return user, nil
}

func (s *authService) CreateUser(ctx context.Context, username, password, role, department string) (*models.User, error) {
	if username == "" || password == "" || role == "" {
		return nil, errors.New("missing fields username/password/role")
	}

	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:         "u_" + username,
		Username:   username,
		Password:   string(hashed),
		Role:       role,
		Department: department,
		CreatedAt:  time.Now(),
	}

	err = s.repo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) ListUsers(ctx context.Context) ([]models.User, error) {
	return s.repo.ListUsers(ctx)
}
