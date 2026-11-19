package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"cleannest/internal/domain"
	"cleannest/pkg/security"
	"strings"
	"time"
)

type AuthUsecase interface {
	RegisterCustomer(ctx context.Context, req RegisterCustomerRequest) (*domain.User, *domain.Customer, string, error)
	Login(ctx context.Context, email, password string) (string, *domain.User, error)
	CreateCashier(ctx context.Context, req CreateCashierRequest) (*domain.User, error)
	ListCashiers(ctx context.Context) ([]domain.User, error)
	DeleteCashier(ctx context.Context, id int64) error
	ValidateToken(tokenString string) (*TokenClaims, error)
}

type TokenClaims struct {
	UserID int64       `json:"user_id"`
	Email  string      `json:"email"`
	Name   string      `json:"name"`
	Role   domain.Role `json:"role"`
	Exp    int64       `json:"exp"`
}

type RegisterCustomerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Address  string `json:"address"`
}

type CreateCashierRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type authUsecase struct {
	userRepo     domain.UserRepository
	customerRepo domain.CustomerRepository
	jwtSecret    []byte
}

func NewAuthUsecase(userRepo domain.UserRepository, customerRepo domain.CustomerRepository, jwtSecret string) AuthUsecase {
	if jwtSecret == "" {
		jwtSecret = "cleannest-super-secret-key-32-chars!"
	}
	return &authUsecase{
		userRepo:     userRepo,
		customerRepo: customerRepo,
		jwtSecret:    []byte(jwtSecret),
	}
}

func (u *authUsecase) generateToken(user *domain.User) (string, error) {
	claims := TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		Name:   user.Name,
		Role:   user.Role,
		Exp:    time.Now().Add(24 * time.Hour).Unix(),
	}

	payloadBytes, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := header + "." + payload

	mac := hmac.New(sha256.New, u.jwtSecret)
	mac.Write([]byte(signingInput))
	signature := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + signature, nil
}

func (u *authUsecase) ValidateToken(tokenString string) (*TokenClaims, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, domain.ErrUnauthorized
	}

	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, u.jwtSecret)
	mac.Write([]byte(signingInput))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return nil, domain.ErrUnauthorized
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, domain.ErrUnauthorized
	}

	var claims TokenClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, domain.ErrUnauthorized
	}

	if claims.Exp < time.Now().Unix() {
		return nil, domain.ErrUnauthorized
	}

	return &claims, nil
}

func (u *authUsecase) RegisterCustomer(ctx context.Context, req RegisterCustomerRequest) (*domain.User, *domain.Customer, string, error) {
	if req.Email == "" || req.Password == "" || req.Name == "" || req.Phone == "" {
		return nil, nil, "", domain.ErrInvalidInput
	}

	existing, _ := u.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, nil, "", domain.ErrEmailAlreadyExists
	}

	passwordHash, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, nil, "", err
	}

	user := &domain.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         domain.RoleCustomer,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, nil, "", err
	}

	customer := &domain.Customer{
		UserID:  &user.ID,
		Name:    req.Name,
		Phone:   req.Phone,
		Address: req.Address,
	}
	if err := u.customerRepo.Create(ctx, customer); err != nil {
		return nil, nil, "", err
	}

	token, err := u.generateToken(user)
	if err != nil {
		return nil, nil, "", err
	}

	return user, customer, token, nil
}

func (u *authUsecase) Login(ctx context.Context, email, password string) (string, *domain.User, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", nil, domain.ErrInvalidCredentials
	}

	if !security.CheckPassword(password, user.PasswordHash) {
		return "", nil, domain.ErrInvalidCredentials
	}

	token, err := u.generateToken(user)
	if err != nil {
		return "", nil, err
	}

	return token, user, nil
}

func (u *authUsecase) CreateCashier(ctx context.Context, req CreateCashierRequest) (*domain.User, error) {
	if req.Email == "" || req.Password == "" || req.Name == "" {
		return nil, domain.ErrInvalidInput
	}

	existing, _ := u.userRepo.GetByEmail(ctx, req.Email)
	if existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	passwordHash, err := security.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         domain.RoleCashier,
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (u *authUsecase) ListCashiers(ctx context.Context) ([]domain.User, error) {
	return u.userRepo.ListCashiers(ctx)
}

func (u *authUsecase) DeleteCashier(ctx context.Context, id int64) error {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if user.Role != domain.RoleCashier {
		return domain.ErrForbidden
	}
	return u.userRepo.Delete(ctx, id)
}
