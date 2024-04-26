package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/janmmiranda/chripy/internal/auth"
)

func (cfg *apiConfig) handlerChirpsCreate(w http.ResponseWriter, req *http.Request) {
	type parameters struct {
		Body string `json:"body`
	}

	bearerToken, err := auth.GetBearerToken(req.Header, BEARER)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	userIDString, issuer, err := auth.ValidateJWT(bearerToken, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if issuer == RefreshIssuer {
		respondWithError(w, http.StatusUnauthorized, "refresh token not accepted for updates")
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	cleaned, err := validateChirp(params.Body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	userIdStr, err := strconv.Atoi(userIDString)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	chirp, err := cfg.DB.CreateChirp(cleaned, userIdStr)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create chirp")
		return
	}
	respondWithJSON(w, http.StatusCreated, Chirp{
		ID:       chirp.ID,
		Body:     chirp.Body,
		AuthorId: chirp.AuthorId,
	})
}

func validateChirp(body string) (string, error) {
	if !maxChirpLength(body) {
		return "", errors.New("Chirp is too long")
	}

	return filterChrip(body), nil

}

func maxChirpLength(chrip string) bool {
	const maxChirpLength = 140
	return len(chrip) <= maxChirpLength
}

func filterChrip(chirp string) string {
	filterSet := make(map[string]bool)
	for _, word := range filterWords {
		filterSet[strings.ToLower(word)] = true
	}

	words := strings.Fields(chirp)

	for i, word := range words {
		if filterSet[strings.ToLower(word)] {
			words[i] = filterWord
		}
	}

	return strings.Join(words, " ")
}
