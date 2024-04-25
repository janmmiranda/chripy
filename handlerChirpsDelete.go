package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/janmmiranda/chripy/internal/auth"
)

func (cfg *apiConfig) handlerChirpsDelete(w http.ResponseWriter, req *http.Request) {
	bearerToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	userIDString, issuer, err := auth.ValidateJWT(bearerToken, cfg.JWTSecret)
	if err != nil {
		fmt.Printf("unable to validate token %v\n", bearerToken)
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if issuer == RefreshIssuer {
		respondWithError(w, http.StatusUnauthorized, "refresh token not accepted for updates")
		return
	}

	userId, errr := strconv.Atoi(userIDString)
	if errr != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't conver userID")
		return
	}
	chirpID := req.PathValue("chirpID")
	iChirpID, err := strconv.Atoi(chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't parse ID: %v", chirpID))
		return
	}

	isDeleted, err := cfg.DB.DeleteChirp(iChirpID, userId)
	if err != nil {
		respondWithError(w, http.StatusForbidden, err.Error())
		return
	}
	if !isDeleted {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, response{})
}
