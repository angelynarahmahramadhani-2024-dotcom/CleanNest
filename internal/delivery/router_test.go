package delivery_test

import (
	"encoding/json"
	"cleannest/internal/delivery"
	"cleannest/internal/repository"
	"cleannest/internal/usecase"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterEndpoints(t *testing.T) {
	memStore := repository.NewMemoryStore()
	userRepo := repository.NewMemoryUserRepository(memStore)
	customerRepo := repository.NewMemoryCustomerRepository(memStore)
	serviceRepo := repository.NewMemoryServiceRepository(memStore)
	orderRepo := repository.NewMemoryOrderRepository(memStore)
	paymentRepo := repository.NewMemoryPaymentRepository(memStore)

	jwtSecret := "test-secret"
	authUC := usecase.NewAuthUsecase(userRepo, customerRepo, jwtSecret)
	custUC := usecase.NewCustomerUsecase(customerRepo)
	srvUC := usecase.NewServiceUsecase(serviceRepo)
	orderUC := usecase.NewOrderUsecase(orderRepo, serviceRepo, customerRepo)
	payUC := usecase.NewPaymentUsecase(paymentRepo, orderRepo)
	repUC := usecase.NewReportUsecase(paymentRepo)

	router := delivery.NewRouter(authUC, custUC, srvUC, orderUC, payUC, repUC)
	handler := router.SetupRoutes()

	// 0. Test Root Index
	reqRoot := httptest.NewRequest("GET", "/", nil)
	recRoot := httptest.NewRecorder()
	handler.ServeHTTP(recRoot, reqRoot)
	if recRoot.Code != http.StatusOK {
		t.Fatalf("expected 200 for root, got %d", recRoot.Code)
	}

	// 1. Test Health Check
	req := httptest.NewRequest("GET", "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp delivery.APIResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success true, got false")
	}

	// 2. Test Get Services
	req = httptest.NewRequest("GET", "/api/v1/services", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for services, got %d", rec.Code)
	}
}
