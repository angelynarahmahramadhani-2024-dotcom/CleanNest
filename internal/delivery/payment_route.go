package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"

	"cleannest/internal/delivery/middleware"
	"cleannest/internal/domain"
	"cleannest/internal/usecase"
)

// Endpoint pencatatan dan riwayat pembayaran.
// Fungsi serveXxx adalah titik masuk delivery yang memanggil usecase (tanpa controller/handler).

func (rt *Router) serveRecordPayment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaimsFromContext(r.Context())
	if claims == nil {
		SendError(w, http.StatusUnauthorized, "sesi kasir tidak valid", domain.ErrUnauthorized)
		return
	}

	var req usecase.RecordPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "format payload json tidak valid", err)
		return
	}

	payment, updatedOrder, err := rt.paymentUsecase.RecordPayment(r.Context(), claims.UserID, req)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	SendJSON(w, http.StatusCreated, "pembayaran berhasil dicatat", map[string]interface{}{
		"payment": payment,
		"order":   updatedOrder,
	})
}

func (rt *Router) serveListOrderPayments(w http.ResponseWriter, r *http.Request) {
	orderIDStr := r.PathValue("orderId")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, "id pesanan tidak valid", err)
		return
	}

	payments, err := rt.paymentUsecase.ListOrderPayments(r.Context(), orderID)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "gagal mengambil riwayat pembayaran", err)
		return
	}

	SendJSON(w, http.StatusOK, "riwayat pembayaran pesanan", payments)
}
