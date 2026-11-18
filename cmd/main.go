package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cleannest/config"
	"cleannest/internal/delivery"
	"cleannest/internal/domain"
	"cleannest/internal/repository"
	"cleannest/internal/usecase"

	_ "github.com/lib/pq" // driver PostgreSQL untuk database/sql
)

func main() {
	fmt.Println("==================================================")
	fmt.Println("   CLEANNEST - SISTEM MANAJEMEN LAUNDRY (KILOAN & SATUAN)")
	fmt.Println("   Clean Architecture Backend (Golang)           ")
	fmt.Println("==================================================")

	// 1. Load Configurations
	cfg := config.LoadConfig()
	log.Printf("[CONFIG] Environment: %s | Port: %s\n", cfg.AppEnv, cfg.AppPort)

	// 2. Initialize Repositories (SQL or In-Memory fallback)
	var (
		userRepo     domain.UserRepository
		customerRepo domain.CustomerRepository
		serviceRepo  domain.ServiceRepository
		orderRepo    domain.OrderRepository
		paymentRepo  domain.PaymentRepository
	)

	// Attempt database connection if configured
	var db *sql.DB
	var dbErr error

	if os.Getenv("USE_SQL_DB") == "true" {
		db, dbErr = sql.Open("postgres", cfg.GetPostgresDSN())
		if dbErr == nil && db.Ping() == nil {
			log.Println("[DATABASE] Terhubung ke PostgreSQL Database:", cfg.DBName)
			userRepo = repository.NewSQLUserRepository(db)
			customerRepo = repository.NewSQLCustomerRepository(db)
			serviceRepo = repository.NewSQLServiceRepository(db)
			orderRepo = repository.NewSQLOrderRepository(db)
			paymentRepo = repository.NewSQLPaymentRepository(db)
		} else {
			log.Printf("[DATABASE WARNING] Gagal konek ke PostgreSQL (%v). Mengaktifkan In-Memory Repository.\n", dbErr)
		}
	}

	if userRepo == nil {
		log.Println("[INFO] Menggunakan In-Memory Data Store (demo/development mode siap pakai).")
		memStore := repository.NewMemoryStore()
		userRepo = repository.NewMemoryUserRepository(memStore)
		customerRepo = repository.NewMemoryCustomerRepository(memStore)
		serviceRepo = repository.NewMemoryServiceRepository(memStore)
		orderRepo = repository.NewMemoryOrderRepository(memStore)
		paymentRepo = repository.NewMemoryPaymentRepository(memStore)
		log.Println("[INFO] Data awal otomatis di-seed: Akun Admin, Kasir, Layanan, dan Contoh Pelanggan.")
	}

	// 3. Initialize Usecases
	authUsecase := usecase.NewAuthUsecase(userRepo, customerRepo, cfg.JWTSecret)
	customerUsecase := usecase.NewCustomerUsecase(customerRepo)
	serviceUsecase := usecase.NewServiceUsecase(serviceRepo)
	orderUsecase := usecase.NewOrderUsecase(orderRepo, serviceRepo, customerRepo)
	paymentUsecase := usecase.NewPaymentUsecase(paymentRepo, orderRepo)
	reportUsecase := usecase.NewReportUsecase(paymentRepo)

	// 4. Initialize Delivery Router (No Controller/Handler)
	router := delivery.NewRouter(
		authUsecase,
		customerUsecase,
		serviceUsecase,
		orderUsecase,
		paymentUsecase,
		reportUsecase,
	)

	routesHandler := router.SetupRoutes()

	// 5. Start HTTP Server
	server := &http.Server{
		Addr:         ":" + cfg.AppPort,
		Handler:      routesHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[SERVER] CleanNest Backend berjalan di http://localhost:%s\n", cfg.AppPort)
		log.Println("[SERVER] Health check: GET http://localhost:" + cfg.AppPort + "/api/v1/health")
		log.Println("[SERVER] Katalog Layanan: GET http://localhost:" + cfg.AppPort + "/api/v1/services")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[SERVER FATAL] Gagal menjalankan server: %v\n", err)
		}
	}()

	// 6. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("[SERVER] Mematikan server secara graceful...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if db != nil {
		_ = db.Close()
	}

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("[SERVER FATAL] Shutdown paksa: %v\n", err)
	}
	log.Println("[SERVER] Server CleanNest berhenti dengan aman.")
}
