package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/jenklu/learn-french/internal/config"
	"github.com/jenklu/learn-french/internal/database"
	"github.com/jenklu/learn-french/internal/handlers"
	"github.com/jenklu/learn-french/internal/middleware"
	"github.com/jenklu/learn-french/internal/services"
	"github.com/robfig/cron/v3"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	ctx := context.Background()
	db, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Init(ctx); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize services
	authService := services.NewAuthService(db)
	exerciseService := services.NewExerciseService(db)
	reminderService := services.NewReminderService(db, services.VAPIDConfig{
		PublicKey:  cfg.VAPIDPublicKey,
		PrivateKey: cfg.VAPIDPrivateKey,
	})
	adminService := services.NewAdminService(db)

	// Initialize session store
	store := sessions.NewCookieStore([]byte(cfg.SessionSecret))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 1 week
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: http.SameSiteLaxMode,
	}

	sessionName := "learn_french_session"

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService, store, sessionName)
	exerciseHandler := handlers.NewExerciseHandler(exerciseService, store, sessionName)
	reminderHandler := handlers.NewReminderHandler(reminderService, store, sessionName)
	adminHandler := handlers.NewAdminHandler(adminService, store, sessionName)

	// Create router
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()

	// Auth routes
	api.HandleFunc("/auth/register", authHandler.Register).Methods("POST")
	api.HandleFunc("/auth/login", authHandler.Login).Methods("POST")
	api.HandleFunc("/auth/logout", authHandler.Logout).Methods("POST")
	api.HandleFunc("/auth/me", authHandler.Me).Methods("GET")

	// Exercise routes
	api.HandleFunc("/exercises", exerciseHandler.GetExercises).Methods("GET")
	api.HandleFunc("/exercises/submit", exerciseHandler.SubmitExercise).Methods("POST")
	api.HandleFunc("/sessions/complete", exerciseHandler.CompleteSession).Methods("POST")
	api.HandleFunc("/sessions", exerciseHandler.GetSessions).Methods("GET")
	api.HandleFunc("/sessions/results", exerciseHandler.GetSessionResults).Methods("GET")
	api.HandleFunc("/progress", exerciseHandler.GetProgress).Methods("GET")
	api.HandleFunc("/stats/daily", exerciseHandler.GetDailyStats).Methods("GET")

	// Reminder routes
	api.HandleFunc("/reminders", reminderHandler.GetReminder).Methods("GET")
	api.HandleFunc("/reminders", reminderHandler.UpdateReminder).Methods("POST")
	api.HandleFunc("/push/subscribe", reminderHandler.StorePushSubscription).Methods("POST")
	api.HandleFunc("/push/subscriptions", reminderHandler.GetPushSubscriptions).Methods("GET")
	api.HandleFunc("/push/subscriptions", reminderHandler.DeletePushSubscription).Methods("DELETE")
	api.HandleFunc("/reminders/trigger", reminderHandler.TriggerReminder).Methods("POST")

	// Admin routes
	admin := api.PathPrefix("/admin").Subrouter()
	admin.HandleFunc("/users", adminHandler.GetUsers).Methods("GET")
	admin.HandleFunc("/users", adminHandler.CreateUser).Methods("POST")
	admin.HandleFunc("/users", adminHandler.UpdateUser).Methods("PUT")
	admin.HandleFunc("/users", adminHandler.DeleteUser).Methods("DELETE")
	admin.HandleFunc("/users/password", adminHandler.UpdateUserPassword).Methods("PUT")
	admin.HandleFunc("/users/{id}", adminHandler.GetUser).Methods("GET")

	// CSRF protection
	csrfMiddleware := csrf.Protect(
		[]byte(cfg.CSRFSecret),
		csrf.Path("/"),
		csrf.Secure(os.Getenv("ENV") == "production"),
		csrf.HttpOnly(true),
		csrf.ErrorHandler(http.HandlerFunc(csrfErrorHandler)),
	)

	// Middleware chain
	api.Use(middleware.CORSMiddleware([]string{cfg.FrontendURL}))
	api.Use(middleware.LoggingMiddleware)
	api.Use(middleware.TimezoneMiddleware)
	api.Use(csrfMiddleware)

	// Static files
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("frontend/dist")))

	// Start cron job for reminders
	c := cron.New()
	if _, err := c.AddFunc("*/1 * * * *", func() {
		reminderService.SendReminders(ctx)
	}); err != nil {
		log.Printf("Failed to add cron job: %v", err)
	}
	c.Start()
	defer c.Stop()

	// Start server
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: r,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Printf("Server starting on port %s", cfg.Port)

	// Graceful shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-done
	log.Println("Server is shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped")
}

func csrfErrorHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	w.Write([]byte(`{"error": "CSRF token invalid"}`))
}
