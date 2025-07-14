package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"server/internal/auth"
	"server/internal/database"
	"time"
)

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

func (cfg *apiConfig) handlerMakeUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	msg := ""
	code := 201

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		msg = "Something went wrong"
		code = 500
		respondWithError(w, code, msg)
		return
	}

	hash, err := auth.HashPassword(params.Password)
	if err != nil {
		msg = "Something went wrong with password"
		code = 500
		respondWithError(w, code, msg)
		return
	}

	args := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hash,
	}

	user, err := cfg.dbQueries.CreateUser(r.Context(), args)
	if err != nil {
		msg = "Something went wrong"
		code = 500
		respondWithError(w, code, msg)
		return
	}

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

func (cfg *apiConfig) handlerUserLogin(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Password  string `json:"password"`
		Email     string `json:"email"`
		ExpiresIn int    `json:"expires_in_seconds"`
	}
	msg := ""
	code := 200

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		msg = "Something went wrong"
		code = 500
		respondWithError(w, code, msg)
		return
	}

	expiresIn := 3600
	if params.ExpiresIn != 0 {
		expiresIn = params.ExpiresIn
		if expiresIn > 3600 {
			expiresIn = 3600
		}
	}

	user, err := cfg.dbQueries.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		msg = "Incorrect email or password"
		code = 401
		respondWithError(w, code, msg)
		return
	}

	err = auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil {
		msg = "Incorrect email or password"
		code = 401
		respondWithError(w, code, msg)
		return
	}

	jwtToken, err := auth.MakeJWT(user.ID, cfg.secret, time.Hour)
	if err != nil {
		msg = "Something went wrong"
		code = 500
		respondWithError(w, code, msg)
		return
	}

	refresh_token, err := auth.MakeRefreshToken()
	if err != nil {
		msg = "Something went wrong"
		code = 500
		respondWithError(w, code, msg)
		return
	}

	args := database.CreateRefreshTokenParams{
		Token:     refresh_token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(60 * 24 * time.Hour),
	}
	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), args)
	if err != nil {
		msg = "Something went wrong"
		code = 500
		respondWithError(w, code, msg)
		return
	}

	respBody := returnUser{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        jwtToken,
		RefreshToken: refresh_token,
	}

	data, _ := json.Marshal(respBody)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(data)

}
