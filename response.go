package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
