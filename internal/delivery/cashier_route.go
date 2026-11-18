package delivery

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"cleannest/internal/domain"
	"cleannest/internal/usecase"
)

// Endpoint admin untuk manajemen akun kasir.
// Fungsi serveXxx adalah titik masuk delivery yang memanggil usecase (tanpa controller/handler).

func (rt *Router) serveCreateCashier(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateCashierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "format payload json tidak valid", err)
		return
	}

	cashier, err := rt.authUsecase.CreateCashier(r.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			SendError(w, http.StatusConflict, err.Error(), err)
			return
		}
		SendError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	SendJSON(w, http.StatusCreated, "akun kasir berhasil dibuat", cashier)
}

func (rt *Router) serveListCashiers(w http.ResponseWriter, r *http.Request) {
	cashiers, err := rt.authUsecase.ListCashiers(r.Context())
	if err != nil {
		SendError(w, http.StatusInternalServerError, "gagal mengambil data kasir", err)
		return
	}
	SendJSON(w, http.StatusOK, "daftar kasir berhasil diambil", cashiers)
}

func (rt *Router) serveDeleteCashier(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		SendError(w, http.StatusBadRequest, "id tidak valid", err)
		return
	}

	if err := rt.authUsecase.DeleteCashier(r.Context(), id); err != nil {
		SendError(w, http.StatusNotFound, "kasir tidak ditemukan", err)
		return
	}

	SendJSON(w, http.StatusOK, "akun kasir berhasil dihapus", nil)
}
