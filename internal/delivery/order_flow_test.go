package delivery_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"cleannest/internal/delivery"
	"cleannest/internal/repository"
	"cleannest/internal/usecase"
)

func newTestServer() http.Handler {
	store := repository.NewMemoryStore()
	userRepo := repository.NewMemoryUserRepository(store)
	customerRepo := repository.NewMemoryCustomerRepository(store)
	serviceRepo := repository.NewMemoryServiceRepository(store)
	orderRepo := repository.NewMemoryOrderRepository(store)
	paymentRepo := repository.NewMemoryPaymentRepository(store)

	router := delivery.NewRouter(
		usecase.NewAuthUsecase(userRepo, customerRepo, "test-secret"),
		usecase.NewCustomerUsecase(customerRepo),
		usecase.NewServiceUsecase(serviceRepo),
		usecase.NewOrderUsecase(orderRepo, serviceRepo, customerRepo),
		usecase.NewPaymentUsecase(paymentRepo, orderRepo),
		usecase.NewReportUsecase(paymentRepo),
	)
	return router.SetupRoutes()
}

func call(t *testing.T, h http.Handler, method, path, token string, body interface{}) (int, map[string]interface{}) {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	var resp map[string]interface{}
	_ = json.Unmarshal(rec.Body.Bytes(), &resp)
	return rec.Code, resp
}

func data(resp map[string]interface{}) map[string]interface{} {
	d, _ := resp["data"].(map[string]interface{})
	return d
}

func TestLoginRejectsWrongPassword(t *testing.T) {
	h := newTestServer()
	code, _ := call(t, h, "POST", "/api/v1/auth/login", "", map[string]string{
		"email": "admin@cleannest.com", "password": "salah",
	})
	if code != http.StatusUnauthorized {
		t.Fatalf("password salah harus 401, dapat %d", code)
	}
}

func TestOrderFlowKiloanDanSatuan(t *testing.T) {
	h := newTestServer()

	// Login kasir
	code, resp := call(t, h, "POST", "/api/v1/auth/login", "", map[string]string{
		"email": "kasir@cleannest.com", "password": "password123",
	})
	if code != http.StatusOK {
		t.Fatalf("login kasir gagal: %d %v", code, resp)
	}
	kasirToken := data(resp)["token"].(string)

	// Order campuran: 2 kg Cuci Setrika Reguler (min 3 kg x 7.000 = 21.000)
	// + 1 pcs Bed Cover (25.000) => total 46.000
	code, resp = call(t, h, "POST", "/api/v1/orders", kasirToken, map[string]interface{}{
		"customer_id": 1,
		"items": []map[string]interface{}{
			{"service_id": 1, "quantity": 2},
			{"service_id": 4, "quantity": 1},
		},
	})
	if code != http.StatusCreated {
		t.Fatalf("buat order gagal: %d %v", code, resp)
	}
	order := data(resp)
	if order["total_price"].(float64) != 46000 {
		t.Fatalf("total harus 46000 (minimum kiloan diterapkan), dapat %v", order["total_price"])
	}
	orderID := int(order["id"].(float64))
	statusURL := fmt.Sprintf("/api/v1/orders/%d/status", orderID)

	// Lompat status tidak boleh
	code, _ = call(t, h, "PATCH", statusURL, kasirToken, map[string]string{"new_status": "siap_diambil"})
	if code == http.StatusOK {
		t.Fatal("transisi diterima -> siap_diambil seharusnya ditolak")
	}

	for _, st := range []string{"diproses", "siap_diambil"} {
		code, resp = call(t, h, "PATCH", statusURL, kasirToken, map[string]string{"new_status": st})
		if code != http.StatusOK {
			t.Fatalf("ubah status ke %s gagal: %d %v", st, code, resp)
		}
	}

	// Belum lunas -> tidak boleh selesai
	code, _ = call(t, h, "PATCH", statusURL, kasirToken, map[string]string{"new_status": "selesai"})
	if code == http.StatusOK {
		t.Fatal("order belum lunas seharusnya tidak bisa selesai")
	}

	code, resp = call(t, h, "POST", "/api/v1/payments", kasirToken, map[string]interface{}{
		"order_id": orderID, "amount": 46000, "method": "qris",
	})
	if code != http.StatusCreated && code != http.StatusOK {
		t.Fatalf("pembayaran gagal: %d %v", code, resp)
	}

	code, resp = call(t, h, "PATCH", statusURL, kasirToken, map[string]string{"new_status": "selesai"})
	if code != http.StatusOK {
		t.Fatalf("order lunas seharusnya bisa selesai: %d %v", code, resp)
	}

	// Pelanggan lain tidak boleh melihat order ini
	code, resp = call(t, h, "POST", "/api/v1/auth/register", "", map[string]string{
		"name": "Siti", "email": "siti@mail.com", "password": "rahasia123", "phone": "0812",
	})
	if code != http.StatusCreated {
		t.Fatalf("registrasi gagal: %d %v", code, resp)
	}
	pelangganToken := data(resp)["token"].(string)
	code, _ = call(t, h, "GET", fmt.Sprintf("/api/v1/orders/%d", orderID), pelangganToken, nil)
	if code != http.StatusNotFound {
		t.Fatalf("pelanggan lain seharusnya tidak bisa melihat order (404), dapat %d", code)
	}
}
