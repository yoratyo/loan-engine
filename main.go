package main

import (
	"context"
	"database/sql"
	"loan-engine/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"loan-engine/handler"
	"loan-engine/notification"
	"loan-engine/repository"
	"loan-engine/service"

	customMiddleware "loan-engine/middleware"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg := config.LoadConfig()

	// Database connection
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	// Initialize components
	loanRepo := repository.NewLoanRepository(db)
	emailSvc := notification.NewSendGridService(cfg.SendgridAPIKey)
	loanSvc := service.NewLoanService(loanRepo, emailSvc)
	loanHandler := handler.NewLoanHandler(loanSvc)

	// Router setup
	r := chi.NewRouter()

	// Middleware
	r.Use(customMiddleware.MetricsMiddleware)

	// Metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// API routes with authentication
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(customMiddleware.BasicAuth)

		r.Post("/loans", loanHandler.CreateLoan)
		r.Route("/loans/{id}", func(r chi.Router) {
			r.Patch("/approve", loanHandler.ApproveLoan)
			r.Post("/investments", loanHandler.AddInvestment)
			r.Patch("/disburse", loanHandler.DisburseLoan)
		})
	})

	// HTTP server configuration
	server := &http.Server{
		Addr:    ":8081",
		Handler: r,
	}

	// Graceful shutdown setup
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)

	go func() {
		log.Printf("Starting server on %s", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-done
	log.Println("Shutting down server...")

	// Graceful shutdown with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}

// Key Feature:
// - State Machine,
// - State Movement Tracker,
// - Optimistic Locking,
// - Atomic Transaction
// - Basic auth
// - Migration DB script
// - metric monitoring, prometheus
// - email notification, sendgrid
// - file generation, gopdf
// - file hosting, file.io

// Out of scope:
// - Handling proof file for approval and disbursement process, assumptions if Client already sent valid URL
// - Rate & ROI calculation, assumptions if it's already calculate when proposed a loan
// - Not handling master data borrower, investor and employee, just add identifier in loan transaction as reference
