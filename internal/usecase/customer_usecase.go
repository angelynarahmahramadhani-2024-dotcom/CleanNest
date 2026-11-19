package usecase

import (
	"context"
	"cleannest/internal/domain"
)

type CustomerUsecase interface {
	CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*domain.Customer, error)
	GetCustomerByID(ctx context.Context, id int64) (*domain.Customer, error)
	GetCustomerByUserID(ctx context.Context, userID int64) (*domain.Customer, error)
	ListCustomers(ctx context.Context, limit, offset int) ([]domain.Customer, error)
	UpdateCustomer(ctx context.Context, id int64, req UpdateCustomerRequest) (*domain.Customer, error)
}

type CreateCustomerRequest struct {
	UserID  *int64 `json:"user_id,omitempty"`
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type UpdateCustomerRequest struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

type customerUsecase struct {
	customerRepo domain.CustomerRepository
}

func NewCustomerUsecase(customerRepo domain.CustomerRepository) CustomerUsecase {
	return &customerUsecase{customerRepo: customerRepo}
}

func (u *customerUsecase) CreateCustomer(ctx context.Context, req CreateCustomerRequest) (*domain.Customer, error) {
	if req.Name == "" || req.Phone == "" {
		return nil, domain.ErrInvalidInput
	}

	cust := &domain.Customer{
		UserID:  req.UserID,
		Name:    req.Name,
		Phone:   req.Phone,
		Address: req.Address,
	}

	if err := u.customerRepo.Create(ctx, cust); err != nil {
		return nil, err
	}

	return cust, nil
}

func (u *customerUsecase) GetCustomerByID(ctx context.Context, id int64) (*domain.Customer, error) {
	return u.customerRepo.GetByID(ctx, id)
}

func (u *customerUsecase) GetCustomerByUserID(ctx context.Context, userID int64) (*domain.Customer, error) {
	return u.customerRepo.GetByUserID(ctx, userID)
}

func (u *customerUsecase) ListCustomers(ctx context.Context, limit, offset int) ([]domain.Customer, error) {
	return u.customerRepo.List(ctx, limit, offset)
}

func (u *customerUsecase) UpdateCustomer(ctx context.Context, id int64, req UpdateCustomerRequest) (*domain.Customer, error) {
	cust, err := u.customerRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		cust.Name = req.Name
	}
	if req.Phone != "" {
		cust.Phone = req.Phone
	}
	if req.Address != "" {
		cust.Address = req.Address
	}

	if err := u.customerRepo.Update(ctx, cust); err != nil {
		return nil, err
	}

	return cust, nil
}
