package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"server/internal/database"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	platform       string
	secret         string
	polkaKey       string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	platform := os.Getenv("PLATFORM")
	secret := os.Getenv("SECRET")
	polkaKey := os.Getenv("POLKA_KEY")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		fmt.Println(err)
	}

	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		dbQueries:      database.New(db),
		platform:       platform,
		secret:         secret,
		polkaKey:       polkaKey,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(200)
		w.Write([]byte("OK"))
	})

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)

	mux.HandleFunc("POST /admin/reset", apiCfg.handlerReset)

	mux.HandleFunc("POST /api/users", apiCfg.handlerMakeUser)

	mux.HandleFunc("PUT /api/users", apiCfg.handlerUpdateUser)

	mux.HandleFunc("POST /api/login", apiCfg.handlerUserLogin)

	mux.HandleFunc("POST /api/refresh", apiCfg.handlerRefresh)

	mux.HandleFunc("POST /api/revoke", apiCfg.handlerRevoke)

	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetChirps)

	mux.HandleFunc("POST /api/chirps", apiCfg.handlerPostChirp)

	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirp)

	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCfg.handlerDeleteChirp)

	mux.HandleFunc("POST /api/polka/webhooks", apiCfg.handlerRedUser)

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	server.ListenAndServe()
}
