package delivery

import (
	"encoding/json"
	"net/http"
	"strconv"

	"cleannest/internal/usecase"
)

// Endpoint katalog layanan laundry (kiloan, satuan, khusus).
// Fungsi serveXxx adalah titik masuk delivery yang memanggil usecase (tanpa controller/handler).

func (rt *Router) serveCreateService(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "format payload json tidak valid", err)
		return
	}

	srv, err := rt.serviceUsecase.CreateService(r.Context(), req)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	SendJSON(w, http.StatusCreated, "layanan laundry berhasil ditambahkan", srv)
}

func (rt *Router) serveUpdateService(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, "id tidak valid", err)
		return
	}

	var req usecase.UpdateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "format payload json tidak valid", err)
		return
	}

	srv, err := rt.serviceUsecase.UpdateService(r.Context(), id, req)
	if err != nil {
		SendError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	SendJSON(w, http.StatusOK, "layanan laundry berhasil diperbarui", srv)
}

func (rt *Router) serveDeleteService(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, "id tidak valid", err)
		return
	}

	if err := rt.serviceUsecase.DeleteService(r.Context(), id); err != nil {
		SendError(w, http.StatusNotFound, "layanan tidak ditemukan", err)
		return
	}

	SendJSON(w, http.StatusOK, "layanan laundry berhasil dihapus", nil)
}

func (rt *Router) serveListServices(w http.ResponseWriter, r *http.Request) {
	activeOnlyStr := r.URL.Query().Get("active_only")
	activeOnly := activeOnlyStr != "false"

	services, err := rt.serviceUsecase.ListServices(r.Context(), activeOnly)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "gagal mengambil data layanan", err)
		return
	}

	SendJSON(w, http.StatusOK, "daftar layanan laundry", services)
}
