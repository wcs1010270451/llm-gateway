package service

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"llm-gateway/backend/internal/entity"
	"llm-gateway/backend/internal/repository"
)

type AuthService struct {
	users  *repository.UserRepository
	secret string
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResult struct {
	Token string      `json:"token"`
	User  entity.User `json:"user"`
}

type AuthClaims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Exp    int64  `json:"exp"`
}

func NewAuthService(users *repository.UserRepository, secret string) *AuthService {
	return &AuthService{users: users, secret: secret}
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (LoginResult, error) {
	email := strings.ToLower(strings.TrimSpace(input.Email))
	if email == "" || strings.TrimSpace(input.Password) == "" {
		return LoginResult{}, validationError("email and password are required")
	}

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return LoginResult{}, err
	}
	if user.Status != "active" {
		return LoginResult{}, validationError("user is disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return LoginResult{}, validationError("invalid email or password")
	}

	token, err := s.sign(AuthClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		Exp:    time.Now().Add(7 * 24 * time.Hour).Unix(),
	})
	if err != nil {
		return LoginResult{}, err
	}

	return LoginResult{Token: token, User: user}, nil
}

func (s *AuthService) CurrentUser(ctx context.Context, token string) (entity.User, error) {
	claims, err := s.Parse(token)
	if err != nil {
		return entity.User{}, err
	}
	user, err := s.users.Get(ctx, claims.UserID)
	if err != nil {
		return entity.User{}, err
	}
	if user.Status != "active" {
		return entity.User{}, validationError("user is disabled")
	}
	return user, nil
}

func (s *AuthService) Parse(token string) (AuthClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return AuthClaims{}, validationError("invalid token")
	}
	signingInput := parts[0] + "." + parts[1]
	expected := s.signature(signingInput)
	if !hmac.Equal([]byte(expected), []byte(parts[2])) {
		return AuthClaims{}, validationError("invalid token")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return AuthClaims{}, validationError("invalid token")
	}
	var claims AuthClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return AuthClaims{}, validationError("invalid token")
	}
	if claims.Exp <= time.Now().Unix() {
		return AuthClaims{}, validationError("token is expired")
	}
	return claims, nil
}

func (s *AuthService) sign(claims AuthClaims) (string, error) {
	header, err := json.Marshal(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	signingInput := base64.RawURLEncoding.EncodeToString(header) + "." + base64.RawURLEncoding.EncodeToString(payload)
	return signingInput + "." + s.signature(signingInput), nil
}

func (s *AuthService) signature(signingInput string) string {
	mac := hmac.New(sha256.New, []byte(s.secret))
	mac.Write([]byte(signingInput))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
