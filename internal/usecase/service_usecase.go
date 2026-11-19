package usecase

import (
	"context"
	"cleannest/internal/domain"
)

type ServiceUsecase interface {
	CreateService(ctx context.Context, req CreateServiceRequest) (*domain.Service, error)
	GetServiceByID(ctx context.Context, id int64) (*domain.Service, error)
	ListServices(ctx context.Context, activeOnly bool) ([]domain.Service, error)
	UpdateService(ctx context.Context, id int64, req UpdateServiceRequest) (*domain.Service, error)
	DeleteService(ctx context.Context, id int64) error
}

type CreateServiceRequest struct {
	Name          string                 `json:"name"`
	Category      domain.ServiceCategory `json:"category"`
	Unit          domain.ServiceUnit     `json:"unit"`
	PricePerUnit  float64                `json:"price_per_unit"`
	MinQty        float64                `json:"min_qty"`
	EstimatedDays int                    `json:"estimated_days"`
	IsActive      *bool                  `json:"is_active,omitempty"`
}

type UpdateServiceRequest struct {
	Name          string                  `json:"name"`
	Category      *domain.ServiceCategory `json:"category,omitempty"`
	Unit          *domain.ServiceUnit     `json:"unit,omitempty"`
	PricePerUnit  *float64                `json:"price_per_unit,omitempty"`
	MinQty        *float64                `json:"min_qty,omitempty"`
	EstimatedDays *int                    `json:"estimated_days,omitempty"`
	IsActive      *bool                   `json:"is_active,omitempty"`
}

type serviceUsecase struct {
	serviceRepo domain.ServiceRepository
}

func NewServiceUsecase(serviceRepo domain.ServiceRepository) ServiceUsecase {
	return &serviceUsecase{serviceRepo: serviceRepo}
}

func (u *serviceUsecase) CreateService(ctx context.Context, req CreateServiceRequest) (*domain.Service, error) {
	if req.Name == "" || req.PricePerUnit <= 0 || req.EstimatedDays <= 0 || req.MinQty < 0 {
		return nil, domain.ErrInvalidInput
	}
	if !req.Category.IsValid() || !req.Unit.IsValid() {
		return nil, domain.ErrInvalidInput
	}

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	srv := &domain.Service{
		Name:          req.Name,
		Category:      req.Category,
		Unit:          req.Unit,
		PricePerUnit:  req.PricePerUnit,
		MinQty:        req.MinQty,
		EstimatedDays: req.EstimatedDays,
		IsActive:      isActive,
	}

	if err := u.serviceRepo.Create(ctx, srv); err != nil {
		return nil, err
	}

	return srv, nil
}

func (u *serviceUsecase) GetServiceByID(ctx context.Context, id int64) (*domain.Service, error) {
	return u.serviceRepo.GetByID(ctx, id)
}

func (u *serviceUsecase) ListServices(ctx context.Context, activeOnly bool) ([]domain.Service, error) {
	return u.serviceRepo.List(ctx, activeOnly)
}

func (u *serviceUsecase) UpdateService(ctx context.Context, id int64, req UpdateServiceRequest) (*domain.Service, error) {
	srv, err := u.serviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		srv.Name = req.Name
	}
	if req.Category != nil {
		if !req.Category.IsValid() {
			return nil, domain.ErrInvalidInput
		}
		srv.Category = *req.Category
	}
	if req.Unit != nil {
		if !req.Unit.IsValid() {
			return nil, domain.ErrInvalidInput
		}
		srv.Unit = *req.Unit
	}
	if req.PricePerUnit != nil {
		if *req.PricePerUnit <= 0 {
			return nil, domain.ErrInvalidInput
		}
		srv.PricePerUnit = *req.PricePerUnit
	}
	if req.MinQty != nil {
		if *req.MinQty < 0 {
			return nil, domain.ErrInvalidInput
		}
		srv.MinQty = *req.MinQty
	}
	if req.EstimatedDays != nil {
		if *req.EstimatedDays <= 0 {
			return nil, domain.ErrInvalidInput
		}
		srv.EstimatedDays = *req.EstimatedDays
	}
	if req.IsActive != nil {
		srv.IsActive = *req.IsActive
	}

	if err := u.serviceRepo.Update(ctx, srv); err != nil {
		return nil, err
	}

	return srv, nil
}

func (u *serviceUsecase) DeleteService(ctx context.Context, id int64) error {
	return u.serviceRepo.Delete(ctx, id)
}
