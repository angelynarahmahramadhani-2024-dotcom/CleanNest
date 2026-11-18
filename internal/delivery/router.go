package delivery

import (
	"net/http"

	"cleannest/internal/delivery/middleware"
	"cleannest/internal/domain"
	"cleannest/internal/usecase"
)

type Router struct {
	authUsecase     usecase.AuthUsecase
	customerUsecase usecase.CustomerUsecase
	serviceUsecase  usecase.ServiceUsecase
	orderUsecase    usecase.OrderUsecase
	paymentUsecase  usecase.PaymentUsecase
	reportUsecase   usecase.ReportUsecase
	authMiddleware  *middleware.AuthMiddleware
}

func NewRouter(
	authUsecase usecase.AuthUsecase,
	customerUsecase usecase.CustomerUsecase,
	serviceUsecase usecase.ServiceUsecase,
	orderUsecase usecase.OrderUsecase,
	paymentUsecase usecase.PaymentUsecase,
	reportUsecase usecase.ReportUsecase,
) *Router {
	return &Router{
		authUsecase:     authUsecase,
		customerUsecase: customerUsecase,
		serviceUsecase:  serviceUsecase,
		orderUsecase:    orderUsecase,
		paymentUsecase:  paymentUsecase,
		reportUsecase:   reportUsecase,
		authMiddleware:  middleware.NewAuthMiddleware(authUsecase),
	}
}

func (rt *Router) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// 0. Root Index / API Explorer
	mux.HandleFunc("GET /{$}", rt.serveRoot)

	// 1. Health check & Info
	mux.HandleFunc("GET /api/v1/health", rt.serveHealthCheck)

	// 2. Authentication (Public)
	mux.HandleFunc("POST /api/v1/auth/register", rt.serveRegisterCustomer)
	mux.HandleFunc("POST /api/v1/auth/login", rt.serveLogin)

	// 3. Public Order Tracking
	mux.HandleFunc("GET /api/v1/orders/track", rt.serveTrackOrderByCode)

	// 4. Public / Shared Services Catalog
	mux.HandleFunc("GET /api/v1/services", rt.serveListServices)

	// 5. Admin - Cashier Management
	adminAuth := rt.authMiddleware.RequireRole(domain.RoleAdmin)
	mux.HandleFunc("POST /api/v1/admin/cashiers", rt.authMiddleware.Authenticate(adminAuth(rt.serveCreateCashier)))
	mux.HandleFunc("GET /api/v1/admin/cashiers", rt.authMiddleware.Authenticate(adminAuth(rt.serveListCashiers)))
	mux.HandleFunc("DELETE /api/v1/admin/cashiers/{id}", rt.authMiddleware.Authenticate(adminAuth(rt.serveDeleteCashier)))

	// 6. Admin - Service Management
	mux.HandleFunc("POST /api/v1/admin/services", rt.authMiddleware.Authenticate(adminAuth(rt.serveCreateService)))
	mux.HandleFunc("PUT /api/v1/admin/services/{id}", rt.authMiddleware.Authenticate(adminAuth(rt.serveUpdateService)))
	mux.HandleFunc("DELETE /api/v1/admin/services/{id}", rt.authMiddleware.Authenticate(adminAuth(rt.serveDeleteService)))

	// 7. Admin - Reports
	mux.HandleFunc("GET /api/v1/admin/reports/daily", rt.authMiddleware.Authenticate(adminAuth(rt.serveDailyReport)))
	mux.HandleFunc("GET /api/v1/admin/reports/monthly", rt.authMiddleware.Authenticate(adminAuth(rt.serveMonthlyReport)))

	// 8. Kasir & Admin - Customers
	cashierAuth := rt.authMiddleware.RequireRole(domain.RoleAdmin, domain.RoleCashier)
	mux.HandleFunc("POST /api/v1/customers", rt.authMiddleware.Authenticate(cashierAuth(rt.serveCreateCustomer)))
	mux.HandleFunc("GET /api/v1/customers", rt.authMiddleware.Authenticate(cashierAuth(rt.serveListCustomers)))
	mux.HandleFunc("GET /api/v1/customers/{id}", rt.authMiddleware.Authenticate(cashierAuth(rt.serveGetCustomerByID)))

	// 9. Kasir & Admin - Orders
	mux.HandleFunc("POST /api/v1/orders", rt.authMiddleware.Authenticate(cashierAuth(rt.serveCreateOrder)))
	mux.HandleFunc("PATCH /api/v1/orders/{id}/status", rt.authMiddleware.Authenticate(cashierAuth(rt.serveUpdateOrderStatus)))
	mux.HandleFunc("POST /api/v1/orders/{id}/cancel", rt.authMiddleware.Authenticate(cashierAuth(rt.serveCancelOrder)))
	mux.HandleFunc("GET /api/v1/orders", rt.authMiddleware.Authenticate(cashierAuth(rt.serveListOrders)))

	// 10. Kasir & Admin - Payments
	mux.HandleFunc("POST /api/v1/payments", rt.authMiddleware.Authenticate(cashierAuth(rt.serveRecordPayment)))
	mux.HandleFunc("GET /api/v1/payments/order/{orderId}", rt.authMiddleware.Authenticate(cashierAuth(rt.serveListOrderPayments)))

	// 11. Pelanggan - View Own Orders
	customerAuth := rt.authMiddleware.RequireRole(domain.RoleCustomer)
	mux.HandleFunc("GET /api/v1/customer/orders", rt.authMiddleware.Authenticate(customerAuth(rt.serveListCustomerOrders)))

	// 12. General Authenticated Order Detail
	mux.HandleFunc("GET /api/v1/orders/{id}", rt.authMiddleware.Authenticate(rt.serveGetOrderByID))

	return mux
}
