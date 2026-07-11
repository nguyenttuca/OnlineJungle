package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/justinas/nosurf"
	"github.com/tuantu/oj-web/internal/auth"
	"github.com/tuantu/oj-web/internal/config"
	"github.com/tuantu/oj-web/internal/database"
	"github.com/tuantu/oj-web/internal/database/sqlcdb"
	"github.com/tuantu/oj-web/internal/dispatcher"
	"github.com/tuantu/oj-web/internal/handlers"
)

func main() {
	cfg := config.Load()

	// 1. Connect to Database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	log.Println("Connected to database successfully")

	// 2. Setup session manager
	auth.InitSession(db)

	// 3. Setup templates
	templatesDir := filepath.Join(".", "templates")
	if err := handlers.InitTemplates(templatesDir); err != nil {
		log.Fatalf("Failed to initialize templates: %v", err)
	}

	queries := sqlcdb.New(db.Pool)

	// 4. Start Background Dispatcher and Health Checker
	bgCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	d := dispatcher.NewDispatcher(queries, bgCtx, 10)
	d.Start()

	go dispatcher.RunHealthChecker(queries)

	// 5. Setup Router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(handlers.SecureHeaders)
	r.Use(handlers.RequestSizeLimit(10 << 20)) // 10MB default body limit
	r.Use(func(next http.Handler) http.Handler {
		csrfHandler := nosurf.New(next)
		csrfHandler.SetBaseCookie(http.Cookie{
			HttpOnly: true,
			Path:     "/",
			Secure:   false, // Set true khi dùng HTTPS
			SameSite: http.SameSiteLaxMode,
		})
		return csrfHandler
	})
	r.Use(handlers.LoadAndSave) // Session middleware

	env := &handlers.Env{
		Pool:    db.Pool,
		Queries: queries,
	}
	r.Use(env.Authenticate) // Authenticate middleware

	// Serve static files
	workDir, _ := filepath.Abs(".")
	staticDir := filepath.Join(workDir, "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Basic Health Check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})

	// App Routes
	handlers.SetupRoutes(r, env)

	// 6. Start Server
	serverAddr := fmt.Sprintf(":%d", cfg.Port)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("Server starting on %s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	sig := <-quit
	log.Printf("Received signal %s, shutting down gracefully...", sig)

	// 1. Stop dispatcher
	cancel()
	log.Println("Dispatcher stopped")

	// 2. Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server forced to shutdown: %v", err)
	}

	// 3. Close DB
	db.Close()
	log.Println("Server stopped gracefully")
}
