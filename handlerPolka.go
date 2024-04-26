package main

import (
	"encoding/json"
	"net/http"

	"github.com/janmmiranda/chripy/internal/auth"
)

const USER_UPGRADED = "user.upgraded"
const APIKEY = "ApiKey"

func (cfg *apiConfig) handlerPolkaWebhooks(w http.ResponseWriter, req *http.Request) {
	type data struct {
		UserId int `json:"user_id"`
	}
	type polkaRequest struct {
		Event string `json:"event"`
		Data  data   `json:"data"`
	}

	apiToken, err := auth.GetBearerToken(req.Header, APIKEY)
	if err != nil || apiToken != cfg.PolkaKey {
		respondWithError(w, http.StatusUnauthorized, err.Error())
		return
	}

	decoder := json.NewDecoder(req.Body)
	params := polkaRequest{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't decode parameters")
		return
	}
	event := params.Event
	if event != USER_UPGRADED {
		respondWithJSON(w, http.StatusOK, response{})
		return
	}

	userId := params.Data.UserId
	upgraded, err := cfg.DB.UpgradeUser(userId)
	if err != nil || !upgraded {
		respondWithError(w, http.StatusNotFound, "user not found")
		return
	}
	respondWithJSON(w, http.StatusOK, response{})
}
