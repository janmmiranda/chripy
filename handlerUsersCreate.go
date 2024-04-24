package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/janmmiranda/chripy/internal/auth"
)

const AccessIssuer = "chirpy-access"
const RefreshIssuer = "chirpy-refresh"
const AccessDuration = 60 * 60
const RefreshDuration = 60 * 60 * 24 * 60

type parameters struct {
	Email    string `json:"email`
	Password string `json:"password`
}

type userResponse struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
	Token string `json:"token"`
}

type response struct {
	ID           int    `json:"id"`
	Email        string `json:"email"`
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (cfg *apiConfig) handlerUsersCreate(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	hashedPwd, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := cfg.DB.CreateUser(params.Email, hashedPwd)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}
	respondWithJSON(w, http.StatusCreated, userResponse{
		ID:    user.ID,
		Email: user.Email,
	})
}

func (cfg *apiConfig) handlerUsersUpdate(w http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	bearerToken, err := auth.GetBearerToken(req.Header)
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
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	userId, errr := strconv.Atoi(userIDString)
	if errr != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	hashedPwd, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	user, err := cfg.DB.UpdateUser(userId, params.Email, hashedPwd)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create user")
		return
	}
	respondWithJSON(w, http.StatusOK, response{
		ID:    user.ID,
		Email: user.Email,
	})
}

func (cfg *apiConfig) handlerUsersLogin(w http.ResponseWriter, req *http.Request) {
	type params struct {
		Email            string `json:"email`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(req.Body)
	p := params{}
	err := decoder.Decode(&p)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}

	user, err := cfg.DB.FindUserByEmail(p.Email)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	err = auth.CheckPasswordHash(p.Password, user.Password)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	accessToken, err := auth.MakeJWT(user.ID, cfg.JWTSecret, time.Duration(AccessDuration)*time.Second, AccessIssuer)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	refreshToken, err := auth.MakeJWT(user.ID, cfg.JWTSecret, time.Duration(RefreshDuration)*time.Second, RefreshIssuer)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, response{
		ID:           user.ID,
		Email:        user.Email,
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}

func (cfg *apiConfig) handlerRefreshToken(w http.ResponseWriter, req *http.Request) {
	bearerToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	userIDString, issuer, err := auth.ValidateJWT(bearerToken, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if issuer != RefreshIssuer {
		respondWithError(w, http.StatusUnauthorized, "access token not accepted for refresh")
		return
	}
	isRevoked, err := cfg.DB.CheckRefreshToken(bearerToken)
	if err != nil || isRevoked {
		respondWithError(w, http.StatusUnauthorized, "refresh token is revoked")
		return
	}
	type response struct {
		Token string `json:"token"`
	}
	userID, err := strconv.Atoi(userIDString)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	tokenStr, err := auth.MakeJWT(userID, cfg.JWTSecret, time.Hour, AccessIssuer)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, response{
		Token: tokenStr,
	})
}

func (cfg *apiConfig) handlerRevokeToken(w http.ResponseWriter, req *http.Request) {
	bearerToken, err := auth.GetBearerToken(req.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	_, issuer, err := auth.ValidateJWT(bearerToken, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	if issuer != RefreshIssuer {
		respondWithError(w, http.StatusUnauthorized, "access token not accepted for refresh")
		return
	}
	err = cfg.DB.RevokeRefreshToken(bearerToken)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, response{})
}
