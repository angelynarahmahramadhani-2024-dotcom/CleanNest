package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"

	"cleannest/internal/usecase"
)

// Endpoint data pelanggan untuk kasir/admin.
// Fungsi serveXxx adalah titik masuk delivery yang memanggil usecase (tanpa controller/handler).

func (rt *Router) serveCreateCustomer(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "format payload json tidak valid", err)
		return
	}

	cust, err := rt.customerUsecase.CreateCustomer(r.Context(), req)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	SendJSON(w, http.StatusCreated, "data pelanggan berhasil disimpan", cust)
}

func (rt *Router) serveListCustomers(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	customers, err := rt.customerUsecase.ListCustomers(r.Context(), limit, offset)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "gagal mengambil data pelanggan", err)
		return
	}

	SendJSON(w, http.StatusOK, "daftar pelanggan", customers)
}

func (rt *Router) serveGetCustomerByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, "id tidak valid", err)
		return
	}

	cust, err := rt.customerUsecase.GetCustomerByID(r.Context(), id)
	if err != nil {
		SendError(w, http.StatusNotFound, "pelanggan tidak ditemukan", err)
		return
	}

	SendJSON(w, http.StatusOK, "detail data pelanggan", cust)
}
