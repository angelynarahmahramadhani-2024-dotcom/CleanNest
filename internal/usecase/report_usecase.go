package usecase

import (
	"context"
	"cleannest/internal/domain"
	"time"
)

type ReportUsecase interface {
	GetDailyReport(ctx context.Context, targetDate time.Time) (*domain.RevenueSummary, error)
	GetMonthlyReport(ctx context.Context, year int, month time.Month) (*domain.RevenueSummary, error)
	GetCustomRangeReport(ctx context.Context, start, end time.Time) (*domain.RevenueSummary, error)
}

type reportUsecase struct {
	paymentRepo domain.PaymentRepository
}

func NewReportUsecase(paymentRepo domain.PaymentRepository) ReportUsecase {
	return &reportUsecase{paymentRepo: paymentRepo}
}

func (u *reportUsecase) GetDailyReport(ctx context.Context, targetDate time.Time) (*domain.RevenueSummary, error) {
	start := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 0, 0, 0, 0, targetDate.Location())
	end := time.Date(targetDate.Year(), targetDate.Month(), targetDate.Day(), 23, 59, 59, 999999999, targetDate.Location())
	return u.paymentRepo.GetRevenueReport(ctx, start, end)
}

func (u *reportUsecase) GetMonthlyReport(ctx context.Context, year int, month time.Month) (*domain.RevenueSummary, error) {
	start := time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0).Add(-time.Nanosecond)
	return u.paymentRepo.GetRevenueReport(ctx, start, end)
}

func (u *reportUsecase) GetCustomRangeReport(ctx context.Context, start, end time.Time) (*domain.RevenueSummary, error) {
	if start.After(end) {
		return nil, domain.ErrInvalidInput
	}
	return u.paymentRepo.GetRevenueReport(ctx, start, end)
}
