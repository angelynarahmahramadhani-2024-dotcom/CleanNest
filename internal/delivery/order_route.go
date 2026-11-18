package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"cleannest/internal/delivery/middleware"
	"cleannest/internal/domain"
	"cleannest/internal/usecase"
)

// Endpoint pesanan laundry: buat, ubah status, batal, detail, lacak, riwayat.
// Fungsi serveXxx adalah titik masuk delivery yang memanggil usecase (tanpa controller/handler).

func (rt *Router) serveCreateOrder(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		SendError(w, http.StatusUnauthorized, "sesi kasir tidak valid", domain.ErrUnauthorized)
		return
	}

	var req usecase.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "format payload json tidak valid", err)
		return
	}

	order, err := rt.orderUsecase.CreateOrder(r.Context(), claims.UserID, req)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	SendJSON(w, http.StatusCreated, "pesanan laundry berhasil dibuat", order)
}

func (rt *Router) serveUpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, "id tidak valid", err)
		return
	}

	var req usecase.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "format payload json tidak valid", err)
		return
	}

	order, err := rt.orderUsecase.UpdateOrderStatus(r.Context(), id, claims.UserID, req)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidStatusTransition) || errors.Is(err, domain.ErrOrderNotFullyPaid) {
			SendError(w, http.StatusUnprocessableEntity, err.Error(), err)
			return
		}
		SendError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	SendJSON(w, http.StatusOK, "status pesanan berhasil diperbarui", order)
}

func (rt *Router) serveCancelOrder(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, "id tidak valid", err)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	order, err := rt.orderUsecase.CancelOrder(r.Context(), id, claims.UserID, req.Reason)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	SendJSON(w, http.StatusOK, "pesanan berhasil dibatalkan", order)
}

func (rt *Router) serveListOrders(w http.ResponseWriter, r *http.Request) {
	statusQuery := r.URL.Query().Get("status")
	customerIDStr := r.URL.Query().Get("customer_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	var filter domain.OrderFilter
	if statusQuery != "" {
		st := domain.OrderStatus(statusQuery)
		filter.Status = &st
	}
	if customerIDStr != "" {
		if cID, err := strconv.ParseInt(customerIDStr, 10, 64); err == nil {
			filter.CustomerID = &cID
		}
	}
	filter.Limit, _ = strconv.Atoi(limitStr)
	filter.Offset, _ = strconv.Atoi(offsetStr)

	orders, err := rt.orderUsecase.ListOrders(r.Context(), filter)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "gagal mengambil data pesanan", err)
		return
	}

	SendJSON(w, http.StatusOK, "daftar pesanan", orders)
}

func (rt *Router) serveGetOrderByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, "id tidak valid", err)
		return
	}

	order, err := rt.orderUsecase.GetOrderByID(r.Context(), id)
	if err != nil {
		SendError(w, http.StatusNotFound, "pesanan tidak ditemukan", err)
		return
	}

	// Pelanggan hanya boleh melihat pesanan miliknya sendiri.
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		SendError(w, http.StatusUnauthorized, "sesi login tidak valid", domain.ErrUnauthorized)
		return
	}
	if claims.Role == domain.RoleCustomer {
		cust, err := rt.customerUsecase.GetCustomerByUserID(r.Context(), claims.UserID)
		if err != nil || cust.ID != order.CustomerID {
			SendError(w, http.StatusNotFound, "pesanan tidak ditemukan", domain.ErrNotFound)
			return
		}
	}

	SendJSON(w, http.StatusOK, "detail pesanan", order)
}

func (rt *Router) serveTrackOrderByCode(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		SendError(w, http.StatusBadRequest, "parameter code wajib diisi (contoh: ?code=CN-...)", domain.ErrInvalidInput)
		return
	}

	order, err := rt.orderUsecase.GetOrderByCode(r.Context(), code)
	if err != nil {
		SendError(w, http.StatusNotFound, "pesanan dengan kode tersebut tidak ditemukan", err)
		return
	}

	SendJSON(w, http.StatusOK, "informasi pelacakan pesanan", map[string]interface{}{
		"order_code":        order.OrderCode,
		"customer_name":     order.CustomerName,
		"status":            order.Status,
		"total_price":       order.TotalPrice,
		"paid_amount":       order.PaidAmount,
		"payment_status":    order.PaymentStatus,
		"estimated_done_at": order.EstimatedDoneAt,
		"items":             order.Items,
		"status_logs":       order.Logs,
	})
}

func (rt *Router) serveListCustomerOrders(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		SendError(w, http.StatusUnauthorized, "sesi login tidak valid", domain.ErrUnauthorized)
		return
	}

	cust, err := rt.customerUsecase.GetCustomerByUserID(r.Context(), claims.UserID)
	if err != nil {
		SendError(w, http.StatusNotFound, "data profil pelanggan tidak ditemukan", err)
		return
	}

	orders, err := rt.orderUsecase.ListCustomerOrders(r.Context(), cust.ID)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "gagal mengambil riwayat pesanan", err)
		return
	}

	SendJSON(w, http.StatusOK, "riwayat pesanan pelanggan", orders)
}
