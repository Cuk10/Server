package main

import (
	"encoding/json"
	"net/http"
	"server/internal/auth"
	"server/internal/database"

	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerPostChirp(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
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

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		msg = "Something went wrong"
		code = 401
		respondWithError(w, code, msg)
		return
	}

	id, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		code = 401
		msg = "Unauthorized"
		respondWithError(w, code, msg)
		return
	}

	args := database.CreateChirpParams{
		Body:   params.Body,
		UserID: id,
	}

	chirp, err := cfg.dbQueries.CreateChirp(r.Context(), args)
	if err != nil {
		msg = "Something went wrong when creating chirp"
		code = 500
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

func (cfg *apiConfig) handlerDeleteChirp(w http.ResponseWriter, r *http.Request) {
	msg := ""
	code := 204
	chirpID := uuid.MustParse(r.PathValue("chirpID"))

	chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		msg = "Something went wrong"
		code = 404
		respondWithError(w, code, msg)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		code = 401
		respondWithError(w, code, msg)
		return
	}

	user_id, err := auth.ValidateJWT(token, cfg.secret)
	if err != nil {
		code = 401
		respondWithError(w, code, msg)
		return
	}

	if user_id != chirp.UserID {
		code = 403
		respondWithError(w, code, msg)
		return
	}

	err = cfg.dbQueries.DeleteChirp(r.Context(), chirp.ID)
	if err != nil {
		msg = "Something went wrong"
		code = 500
		respondWithError(w, code, msg)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
}
