package delivery

import (
	"encoding/json"
	"errors"
	"net/http"

	"cleannest/internal/domain"
	"cleannest/internal/usecase"
)

// Endpoint otentikasi: registrasi pelanggan dan login.
// Fungsi serveXxx adalah titik masuk delivery yang memanggil usecase (tanpa controller/handler).

func (rt *Router) serveRegisterCustomer(w http.ResponseWriter, r *http.Request) {
	var req usecase.RegisterCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "format payload json tidak valid", err)
		return
	}

	user, customer, token, err := rt.authUsecase.RegisterCustomer(r.Context(), req)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			SendError(w, http.StatusConflict, err.Error(), err)
			return
		}
		SendError(w, http.StatusBadRequest, err.Error(), err)
		return
	}

	SendJSON(w, http.StatusCreated, "registrasi pelanggan berhasil", map[string]interface{}{
		"user":     user,
		"customer": customer,
		"token":    token,
	})
}

func (rt *Router) serveLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, "format payload json tidak valid", err)
		return
	}

	token, user, err := rt.authUsecase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		SendError(w, http.StatusUnauthorized, "email atau password salah", err)
		return
	}

	SendJSON(w, http.StatusOK, "login berhasil", map[string]interface{}{
		"token": token,
		"user":  user,
	})
}
