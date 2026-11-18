package middleware

import (
	"context"
	"encoding/json"
	"cleannest/internal/domain"
	"cleannest/internal/usecase"
	"net/http"
	"strings"
)

type contextKey string

const (
	UserClaimsContextKey contextKey = "user_claims"
)

type AuthMiddleware struct {
	authUsecase usecase.AuthUsecase
}

func NewAuthMiddleware(authUsecase usecase.AuthUsecase) *AuthMiddleware {
	return &AuthMiddleware{authUsecase: authUsecase}
}

func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			writeUnauthorized(w, "header Authorization tidak ditemukan")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			writeUnauthorized(w, "format token tidak valid, gunakan format 'Bearer <token>'")
			return
		}

		claims, err := m.authUsecase.ValidateToken(parts[1])
		if err != nil {
			writeUnauthorized(w, "token tidak valid atau sudah kedaluwarsa")
			return
		}

		ctx := context.WithValue(r.Context(), UserClaimsContextKey, claims)
		next(w, r.WithContext(ctx))
	}
}

func (m *AuthMiddleware) RequireRole(roles ...domain.Role) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserClaimsContextKey).(*usecase.TokenClaims)
			if !ok || claims == nil {
				writeUnauthorized(w, "sesi login tidak ditemukan")
				return
			}

			allowed := false
			for _, role := range roles {
				if claims.Role == role {
					allowed = true
					break
				}
			}

			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusForbidden)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": "hak akses tidak mencukupi",
				})
				return
			}

			next(w, r)
		}
	}
}

func GetClaimsFromContext(ctx context.Context) *usecase.TokenClaims {
	claims, ok := ctx.Value(UserClaimsContextKey).(*usecase.TokenClaims)
	if !ok {
		return nil
	}
	return claims
}

func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"message": message,
	})
}
