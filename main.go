package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"github.com/joho/godotenv"
	"github.com/domifgit/chirpy/internal/database"
	"github.com/domifgit/chirpy/internal/handler"
)

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("database connection error")
	}

	h := handler.New(database.New(db), os.Getenv("PLATFORM"))

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app/", h.MiddlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", h.Healthz)
	mux.HandleFunc("POST /api/chirps", h.InsertChirp)
	mux.HandleFunc("POST /api/users", h.CreateUser)
	mux.HandleFunc("GET /admin/metrics", h.Metrics)
	mux.HandleFunc("POST /admin/reset", h.Reset)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := http.Server{
		Handler:      mux,
		Addr:         ":" + port,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
}
