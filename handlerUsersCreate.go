package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/janmmiranda/chripy/internal/auth"
)

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
	User
	Token string `json:"token"`
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
	userIDString, err := auth.ValidateJWT(bearerToken, cfg.JWTSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
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
		User: User{
			ID:    user.ID,
			Email: user.Email,
		},
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

	defaultExpireTime := 60 * 60 * 24
	if p.ExpiresInSeconds == 0 {
		p.ExpiresInSeconds = defaultExpireTime
	} else if p.ExpiresInSeconds > defaultExpireTime {
		p.ExpiresInSeconds = defaultExpireTime
	}
	tokenStr, err := auth.MakeJWT(user.ID, cfg.JWTSecret, time.Duration(p.ExpiresInSeconds)*time.Second)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}
	respondWithJSON(w, http.StatusOK, response{
		User: User{
			ID:    user.ID,
			Email: user.Email,
		},
		Token: tokenStr,
	})
}
