package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/internal/auth"
	"strings"
	"time"
)

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

func (cfg *apiConfig) handlerRefresh(w http.ResponseWriter, r *http.Request) {
	msg := ""
	code := 200

	refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		code = 401
		respondWithError(w, code, msg)
		return
	}

	err = cfg.ValidateRefreshToken(refresh_token, r)
	if err != nil {
		code = 401
		respondWithError(w, code, msg)
		return
	}

	user, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), refresh_token)
	if err != nil {
		code = 500
		respondWithError(w, code, msg)
		return
	}

	token, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		code = 500
		respondWithError(w, code, msg)
		return
	}

	respBody := returnUser{
		Token: token,
	}

	data, _ := json.Marshal(respBody)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)
}

func (cfg *apiConfig) handlerRevoke(w http.ResponseWriter, r *http.Request) {
	msg := ""
	code := 204

	refresh_token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		code = 401
		respondWithError(w, code, msg)
		return
	}
	err = cfg.ValidateRefreshToken(refresh_token, r)
	if err != nil {
		code = 401
		respondWithError(w, code, msg)
		return
	}

	err = cfg.dbQueries.RevokeRefreshToken(r.Context(), refresh_token)
	if err != nil {
		code = 500
		respondWithError(w, code, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
}

func (cfg *apiConfig) ValidateRefreshToken(token string, r *http.Request) error {
	refToken, err := cfg.dbQueries.GetRefreshToken(r.Context(), token)
	if err != nil {
		return err
	}

	if refToken.RevokedAt.Valid {
		return fmt.Errorf("token revoked")
	}
	return nil
}
