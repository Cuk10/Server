package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/internal/database"
	"strings"
	"time"

	"github.com/google/uuid"
)

func handlerDecode(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}
	msg := ""
	code := 200

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		msg = "Something went wrong"
		code = 500
	}

	if len(params.Body) > 140 {
		err = fmt.Errorf("chirp is too long")
		msg = "Chirp is too long"
		code = 400
	}

	if err != nil {
		respondWithError(w, code, msg)
	} else {
		payload := badWordReplacement(params.Body)
		respondWithJSON(w, code, payload)
	}
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf(`<html>
  <body>
    <h1>Welcome, Chirpy Admin</h1>
    <p>Chirpy has been visited %d times!</p>
  </body>
</html>`, cfg.fileserverHits.Load())))
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, r *http.Request) {
	if cfg.platform != "dev" {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(403)
		w.Write([]byte("Forbidden"))
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(""))
	cfg.fileserverHits.Store(0)
	err := cfg.dbQueries.ResetUsers(r.Context())
	if err != nil {
		fmt.Println(err)
	}
}

func (cfg *apiConfig) handlerUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}
	msg := ""
	code := 201

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		msg = "Something went wrong"
		code = 500
	}
	user, err := cfg.dbQueries.CreateUser(r.Context(), params.Email)
	if err != nil {
		msg = "Something went wrong"
		code = 500
	}

	if err != nil {
		respondWithError(w, code, msg)
	} else {
		respondWithUser(w, code, user)
	}
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	type returnVals struct {
		Error string `json:"error"`
	}

	respBody := returnVals{
		Error: msg,
	}

	data, _ := json.Marshal(respBody)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func respondWithJSON(w http.ResponseWriter, code int, payload string) {
	type returnVals struct {
		ClBody string `json:"cleaned_body"`
	}

	respBody := returnVals{
		ClBody: payload,
	}

	data, _ := json.Marshal(respBody)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func badWordReplacement(body string) string {
	bodyWords := strings.Split(body, " ")
	for i, word := range bodyWords {
		if strings.ToLower(word) == "kerfuffle" || strings.ToLower(word) == "sharbert" || strings.ToLower(word) == "fornax" {
			bodyWords[i] = "****"
		}
	}
	return strings.Join(bodyWords, " ")
}

func respondWithUser(w http.ResponseWriter, code int, user database.User) {
	type returnVals struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}

	respBody := returnVals{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	data, _ := json.Marshal(respBody)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}
