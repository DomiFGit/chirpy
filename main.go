package main

import _ "github.com/lib/pq"

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"fmt"
	"sync/atomic"
	"strings"
	"regexp"
	"github.com/joho/godotenv"
	"database/sql"
	"github.com/domifgit/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries *database.Queries
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func filter_chirp(text string) string {
	badWords := []string{"kerfuffle", "sharbert", "fornax"}
	pattern := "(?i)" + strings.Join(badWords, "|")
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(text, "****")
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("database connection error")
	}

	cfg := &apiConfig{}
	cfg.dbQueries = database.New(db)

	mux := http.NewServeMux()
	mux.Handle("/app/", http.StripPrefix("/app/", cfg.middlewareMetricsInc(http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", func (w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	mux.HandleFunc("GET /admin/metrics", func (w http.ResponseWriter, r *http.Request){
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		message := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", cfg.fileserverHits.Load())
		w.Write([]byte(message))
	})
	mux.HandleFunc("POST /admin/reset", func (w http.ResponseWriter, r *http.Request){
		cfg.fileserverHits.Store(0)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		message := fmt.Sprintf("Hits reset to: %d", cfg.fileserverHits.Load())
		w.Write([]byte(message))
	})
	mux.HandleFunc("POST /api/validate_chirp", func (w http.ResponseWriter, r *http.Request){
		type chirp struct {
			Body string `json:"body"`
		}
		type errorResp struct {
			Error string `json:"error"`
		}
		type validResp struct {
			CleanedBody string `json:"cleaned_body"`
		}

		decoder := json.NewDecoder(r.Body)
		var c chirp
		if err := decoder.Decode(&c); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(500)
			resp, _ := json.Marshal(errorResp{Error: "Could not decode response"})
			w.Write(resp)
			return
		}
		if len(c.Body) > 140 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(400)
			resp, _ := json.Marshal(errorResp{Error: "Chirp is too long"})
			w.Write(resp)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		clean := filter_chirp(c.Body)
		resp, _ := json.Marshal(validResp{CleanedBody: clean})
		w.Write(resp)

	})
	

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
