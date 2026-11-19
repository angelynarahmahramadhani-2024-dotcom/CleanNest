package usecase

import (
	"context"
	"cleannest/internal/domain"
	"time"
)

type PaymentUsecase interface {
	RecordPayment(ctx context.Context, receivedBy int64, req RecordPaymentRequest) (*domain.Payment, *domain.Order, error)
	GetPaymentByID(ctx context.Context, id int64) (*domain.Payment, error)
	ListOrderPayments(ctx context.Context, orderID int64) ([]domain.Payment, error)
}

type RecordPaymentRequest struct {
	OrderID int64                `json:"order_id"`
	Amount  float64              `json:"amount"`
	Method  domain.PaymentMethod `json:"method"`
}

type paymentUsecase struct {
	paymentRepo domain.PaymentRepository
	orderRepo   domain.OrderRepository
}

func NewPaymentUsecase(paymentRepo domain.PaymentRepository, orderRepo domain.OrderRepository) PaymentUsecase {
	return &paymentUsecase{
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
	}
}

func (u *paymentUsecase) RecordPayment(ctx context.Context, receivedBy int64, req RecordPaymentRequest) (*domain.Payment, *domain.Order, error) {
	if req.OrderID <= 0 || req.Amount <= 0 {
		return nil, nil, domain.ErrInvalidInput
	}

	if req.Method != domain.MethodCash && req.Method != domain.MethodTransfer && req.Method != domain.MethodQRIS {
		return nil, nil, domain.ErrInvalidInput
	}

	order, err := u.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		return nil, nil, err
	}

	if order.Status == domain.StatusDibatalkan {
		return nil, nil, domain.ErrOrderAlreadyCancelled
	}

	remaining := order.TotalPrice - order.PaidAmount
	if req.Amount > remaining {
		return nil, nil, domain.ErrPaymentExceedsTotal
	}

	newPaidAmount := order.PaidAmount + req.Amount
	var newPaymentStatus domain.PaymentStatus

	if newPaidAmount >= order.TotalPrice {
		newPaymentStatus = domain.PaymentLunas
	} else if newPaidAmount > 0 {
		newPaymentStatus = domain.PaymentSebagian
	} else {
		newPaymentStatus = domain.PaymentBelum
	}

	payment := &domain.Payment{
		OrderID:    order.ID,
		Amount:     req.Amount,
		Method:     req.Method,
		PaidAt:     time.Now(),
		ReceivedBy: receivedBy,
	}

	if err := u.paymentRepo.Create(ctx, payment); err != nil {
		return nil, nil, err
	}

	if err := u.orderRepo.UpdatePayment(ctx, order.ID, newPaidAmount, newPaymentStatus); err != nil {
		return nil, nil, err
	}

	updatedOrder, err := u.orderRepo.GetByID(ctx, order.ID)
	if err != nil {
		return nil, nil, err
	}

	return payment, updatedOrder, nil
}

func (u *paymentUsecase) GetPaymentByID(ctx context.Context, id int64) (*domain.Payment, error) {
	return u.paymentRepo.GetByID(ctx, id)
}

func (u *paymentUsecase) ListOrderPayments(ctx context.Context, orderID int64) ([]domain.Payment, error) {
	return u.paymentRepo.ListByOrderID(ctx, orderID)
}
