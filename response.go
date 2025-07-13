package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/internal/database"
	"strings"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body   string `json:"body"`
		UserID string `json:"user_id"`
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

	if len(params.Body) > 140 {
		msg = "Chirp is too long"
		code = 400
		respondWithError(w, code, msg)
	}

	args := database.CreateChirpParams{
		Body:   params.Body,
		UserID: uuid.MustParse(params.UserID),
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), args)

	if err != nil {
		msg = "Something went wrong"
		code = 500
		respondWithError(w, code, msg)
	} else {
		respBody := returnChirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}

		data, _ := json.Marshal(respBody)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		w.Write(data)
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
		respBody := returnUser{
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

func badWordReplacement(body string) string {
	bodyWords := strings.Split(body, " ")
	for i, word := range bodyWords {
		if strings.ToLower(word) == "kerfuffle" || strings.ToLower(word) == "sharbert" || strings.ToLower(word) == "fornax" {
			bodyWords[i] = "****"
		}
	}
	return strings.Join(bodyWords, " ")
}

func (cfg *apiConfig) handlerGetChirps(w http.ResponseWriter, r *http.Request) {
	msg := ""
	code := 200

	chirps, err := cfg.dbQueries.GetAllChirps(r.Context())

	if err != nil {
		msg = "Something went wrong"
		code = 500
		respondWithError(w, code, msg)
		return
	}
	respBody := []returnChirp{}
	for _, chirp := range chirps {
		respChirp := returnChirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
		respBody = append(respBody, respChirp)
	}

	data, _ := json.Marshal(respBody)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)

}

func (cfg *apiConfig) handlerGetChirp(w http.ResponseWriter, r *http.Request) {
	msg := ""
	code := 200
	chirpID := uuid.MustParse(r.PathValue("chirpID"))

	chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpID)

	if err != nil {
		msg = "Something went wrong"
		code = 404
		respondWithError(w, code, msg)
		return
	}

	respBody := returnChirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	data, _ := json.Marshal(respBody)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}
