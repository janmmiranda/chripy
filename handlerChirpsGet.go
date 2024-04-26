package main

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
)

const ASC = "asc"
const DESC = "desc"

func (cfg *apiConfig) handlerChirpsGet(w http.ResponseWriter, req *http.Request) {
	authorIdStr := req.URL.Query().Get("author_id")
	authorId := 0
	var err error
	if authorIdStr != "" {
		authorId, err = strconv.Atoi(authorIdStr)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't convert authorId")
			return
		}
	}

	sortingOrder := req.URL.Query().Get("sort")
	if sortingOrder == "" {
		sortingOrder = ASC
	}

	dbChirps, err := cfg.DB.GetChirps(authorId)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve chirps")
		return
	}

	chirps := []Chirp{}
	for _, dbChirp := range dbChirps {
		chirps = append(chirps, Chirp{
			ID:   dbChirp.ID,
			Body: dbChirp.Body,
		})
	}

	if sortingOrder == DESC {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID > chirps[j].ID
		})
	} else {
		sort.Slice(chirps, func(i, j int) bool {
			return chirps[i].ID < chirps[j].ID
		})
	}

	if len(chirps) == 1 {
		respondWithJSON(w, http.StatusOK, chirps[0])
		return
	}

	respondWithJSON(w, http.StatusOK, chirps)
}

func (cfg *apiConfig) handlerChirpGet(w http.ResponseWriter, req *http.Request) {
	chirpID := req.PathValue("chirpID")
	iChirpID, err := strconv.Atoi(chirpID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Couldn't parse ID: %v", chirpID))
		return
	}
	dbChirp, err := cfg.DB.GetChirp(iChirpID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, Chirp{
		ID:       dbChirp.ID,
		Body:     dbChirp.Body,
		AuthorId: dbChirp.AuthorId,
	})
}
