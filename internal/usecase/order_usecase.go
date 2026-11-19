package usecase

import (
	"context"
	"fmt"
	"cleannest/internal/domain"
	"math/rand"
	"time"
)

type OrderUsecase interface {
	CreateOrder(ctx context.Context, cashierID int64, req CreateOrderRequest) (*domain.Order, error)
	GetOrderByID(ctx context.Context, id int64) (*domain.Order, error)
	GetOrderByCode(ctx context.Context, code string) (*domain.Order, error)
	ListOrders(ctx context.Context, filter domain.OrderFilter) ([]domain.Order, error)
	ListCustomerOrders(ctx context.Context, customerID int64) ([]domain.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, changedBy int64, req UpdateOrderStatusRequest) (*domain.Order, error)
	CancelOrder(ctx context.Context, orderID int64, cancelledBy int64, reason string) (*domain.Order, error)
}

type OrderItemInput struct {
	ServiceID int64   `json:"service_id"`
	Quantity  float64 `json:"quantity"` // kg, pcs, atau m2 sesuai satuan layanan
}

type CreateOrderRequest struct {
	CustomerID int64            `json:"customer_id"`
	Notes      string           `json:"notes"`
	Items      []OrderItemInput `json:"items"`
}

type UpdateOrderStatusRequest struct {
	NewStatus domain.OrderStatus `json:"new_status"`
	Note      string             `json:"note"`
}

type orderUsecase struct {
	orderRepo    domain.OrderRepository
	serviceRepo  domain.ServiceRepository
	customerRepo domain.CustomerRepository
}

func NewOrderUsecase(
	orderRepo domain.OrderRepository,
	serviceRepo domain.ServiceRepository,
	customerRepo domain.CustomerRepository,
) OrderUsecase {
	return &orderUsecase{
		orderRepo:    orderRepo,
		serviceRepo:  serviceRepo,
		customerRepo: customerRepo,
	}
}

// calculateSubtotal menghitung biaya satu item: quantity x harga per satuan.
// Jika quantity di bawah minimum layanan (mis. kiloan min. 3 kg), yang ditagih adalah minimumnya.
func calculateSubtotal(quantity float64, service *domain.Service) float64 {
	billedQty := quantity
	if service.MinQty > 0 && billedQty < service.MinQty {
		billedQty = service.MinQty
	}
	return billedQty * service.PricePerUnit
}

func generateOrderCode() string {
	now := time.Now()
	randomNum := rand.Intn(9000) + 1000
	return fmt.Sprintf("CN-%s-%04d", now.Format("200601021504"), randomNum)
}

func (u *orderUsecase) CreateOrder(ctx context.Context, cashierID int64, req CreateOrderRequest) (*domain.Order, error) {
	if req.CustomerID <= 0 || len(req.Items) == 0 {
		return nil, domain.ErrInvalidInput
	}

	_, err := u.customerRepo.GetByID(ctx, req.CustomerID)
	if err != nil {
		return nil, err
	}

	var totalPrice float64
	var maxEstimatedDays int = 1
	var orderItems []domain.OrderItem

	for _, itemInput := range req.Items {
		if itemInput.Quantity <= 0 {
			return nil, domain.ErrInvalidInput
		}

		service, err := u.serviceRepo.GetByID(ctx, itemInput.ServiceID)
		if err != nil {
			return nil, err
		}
		if !service.IsActive {
			return nil, domain.ErrServiceInactive
		}

		subtotal := calculateSubtotal(itemInput.Quantity, service)
		totalPrice += subtotal

		if service.EstimatedDays > maxEstimatedDays {
			maxEstimatedDays = service.EstimatedDays
		}

		orderItems = append(orderItems, domain.OrderItem{
			ServiceID:    service.ID,
			ServiceName:  service.Name,
			Quantity:     itemInput.Quantity,
			Unit:         service.Unit,
			PricePerUnit: service.PricePerUnit,
			Subtotal:     subtotal,
		})
	}

	estimatedDoneAt := time.Now().AddDate(0, 0, maxEstimatedDays)

	order := &domain.Order{
		OrderCode:       generateOrderCode(),
		CustomerID:      req.CustomerID,
		CashierID:       cashierID,
		Status:          domain.StatusDiterima,
		TotalPrice:      totalPrice,
		PaidAmount:      0,
		PaymentStatus:   domain.PaymentBelum,
		Notes:           req.Notes,
		EstimatedDoneAt: estimatedDoneAt,
		Items:           orderItems,
	}

	if err := u.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	return order, nil
}

func (u *orderUsecase) GetOrderByID(ctx context.Context, id int64) (*domain.Order, error) {
	return u.orderRepo.GetByID(ctx, id)
}

func (u *orderUsecase) GetOrderByCode(ctx context.Context, code string) (*domain.Order, error) {
	return u.orderRepo.GetByCode(ctx, code)
}

func (u *orderUsecase) ListOrders(ctx context.Context, filter domain.OrderFilter) ([]domain.Order, error) {
	return u.orderRepo.List(ctx, filter)
}

func (u *orderUsecase) ListCustomerOrders(ctx context.Context, customerID int64) ([]domain.Order, error) {
	return u.orderRepo.List(ctx, domain.OrderFilter{
		CustomerID: &customerID,
		Limit:      50,
	})
}

func isValidStatusTransition(current, next domain.OrderStatus) bool {
	switch current {
	case domain.StatusDiterima:
		return next == domain.StatusDiproses
	case domain.StatusDiproses:
		return next == domain.StatusSiapDiambil
	case domain.StatusSiapDiambil:
		return next == domain.StatusSelesai
	default:
		return false
	}
}

func (u *orderUsecase) UpdateOrderStatus(ctx context.Context, orderID int64, changedBy int64, req UpdateOrderStatusRequest) (*domain.Order, error) {
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order.Status == domain.StatusDibatalkan {
		return nil, domain.ErrOrderAlreadyCancelled
	}

	if !isValidStatusTransition(order.Status, req.NewStatus) {
		return nil, domain.ErrInvalidStatusTransition
	}

	if req.NewStatus == domain.StatusSelesai && order.PaymentStatus != domain.PaymentLunas {
		return nil, domain.ErrOrderNotFullyPaid
	}

	if err := u.orderRepo.UpdateStatus(ctx, order.ID, req.NewStatus); err != nil {
		return nil, err
	}

	note := req.Note
	if note == "" {
		note = fmt.Sprintf("Status diubah dari %s ke %s", order.Status, req.NewStatus)
	}

	statusLog := &domain.OrderStatusLog{
		OrderID:   order.ID,
		OldStatus: order.Status,
		NewStatus: req.NewStatus,
		ChangedBy: changedBy,
		Note:      note,
		ChangedAt: time.Now(),
	}
	if err := u.orderRepo.CreateStatusLog(ctx, statusLog); err != nil {
		return nil, err
	}

	return u.orderRepo.GetByID(ctx, order.ID)
}

func (u *orderUsecase) CancelOrder(ctx context.Context, orderID int64, cancelledBy int64, reason string) (*domain.Order, error) {
	order, err := u.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if order.Status == domain.StatusDibatalkan {
		return nil, domain.ErrOrderAlreadyCancelled
	}

	if order.Status == domain.StatusSiapDiambil || order.Status == domain.StatusSelesai {
		return nil, domain.ErrInvalidStatusTransition
	}

	if err := u.orderRepo.UpdateStatus(ctx, order.ID, domain.StatusDibatalkan); err != nil {
		return nil, err
	}

	logNote := "Pesanan dibatalkan"
	if reason != "" {
		logNote = fmt.Sprintf("Pesanan dibatalkan. Alasan: %s", reason)
	}

	statusLog := &domain.OrderStatusLog{
		OrderID:   order.ID,
		OldStatus: order.Status,
		NewStatus: domain.StatusDibatalkan,
		ChangedBy: cancelledBy,
		Note:      logNote,
		ChangedAt: time.Now(),
	}
	if err := u.orderRepo.CreateStatusLog(ctx, statusLog); err != nil {
		return nil, err
	}

	return u.orderRepo.GetByID(ctx, order.ID)
}
