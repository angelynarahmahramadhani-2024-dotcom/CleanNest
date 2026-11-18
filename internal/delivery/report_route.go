package delivery

import (
	"net/http"
	"strconv"
	"time"
)

// Endpoint laporan pendapatan untuk admin.
// Fungsi serveXxx adalah titik masuk delivery yang memanggil usecase (tanpa controller/handler).

func (rt *Router) serveDailyReport(w http.ResponseWriter, r *http.Request) {
	dateStr := r.URL.Query().Get("date")
	targetDate := time.Now()
	if dateStr != "" {
		if parsed, err := time.Parse("2006-01-02", dateStr); err == nil {
			targetDate = parsed
		}
	}

	report, err := rt.reportUsecase.GetDailyReport(r.Context(), targetDate)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "gagal membuat laporan pendapatan harian", err)
		return
	}

	SendJSON(w, http.StatusOK, "laporan pendapatan harian", report)
}

func (rt *Router) serveMonthlyReport(w http.ResponseWriter, r *http.Request) {
	yearStr := r.URL.Query().Get("year")
	monthStr := r.URL.Query().Get("month")

	now := time.Now()
	year := now.Year()
	month := now.Month()

	if yearStr != "" {
		if y, err := strconv.Atoi(yearStr); err == nil {
			year = y
		}
	}
	if monthStr != "" {
		if m, err := strconv.Atoi(monthStr); err == nil && m >= 1 && m <= 12 {
			month = time.Month(m)
		}
	}

	report, err := rt.reportUsecase.GetMonthlyReport(r.Context(), year, month)
	if err != nil {
		SendError(w, http.StatusInternalServerError, "gagal membuat laporan pendapatan bulanan", err)
		return
	}

	SendJSON(w, http.StatusOK, "laporan pendapatan bulanan", report)
}
