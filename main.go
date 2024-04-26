package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

const filepathRoot = "."
const port = "8080"
const serverFailed = "Something went wrong"
const filterWord = "****"
const dbFilename = "database.json"

var filterWords = []string{"kerfuffle", "sharbert", "fornax"}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	secretKey := os.Getenv("JWT_SECRET")
	polkaKey := os.Getenv("POLKA_KEY")
	dbg := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()
	if *dbg {
		DeleteDB(dbFilename)
	}

	db, err := NewDB(dbFilename)
	if err != nil {
		log.Fatal(err)
	}

	apiConfig := apiConfig{
		fileServerHits: 0,
		DB:             db,
		JWTSecret:      secretKey,
		PolkaKey:       polkaKey,
	}

	mux := http.NewServeMux()
	mux.Handle("/app/*", apiConfig.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", handlerHealth)
	mux.HandleFunc("GET /admin/metrics", apiConfig.handlerMetrics)
	mux.HandleFunc("GET /api/reset", apiConfig.handlerReset)

	mux.HandleFunc("POST /api/chirps", apiConfig.handlerChirpsCreate)
	mux.HandleFunc("GET /api/chirps", apiConfig.handlerChirpsGet)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiConfig.handlerChirpGet)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiConfig.handlerChirpsDelete)

	mux.HandleFunc("POST /api/users", apiConfig.handlerUsersCreate)
	mux.HandleFunc("POST /api/login", apiConfig.handlerUsersLogin)
	mux.HandleFunc("PUT /api/users", apiConfig.handlerUsersUpdate)

	mux.HandleFunc("POST /api/refresh", apiConfig.handlerRefreshToken)
	mux.HandleFunc("POST /api/revoke", apiConfig.handlerRevokeToken)

	mux.HandleFunc("POST /api/polka/webhooks", apiConfig.handlerPolkaWebhooks)

	corsMux := middlewareLog(middlewareCors(mux))

	server := &http.Server{
		Addr:    ":" + port,
		Handler: corsMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}

func middlewareCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func handlerHealth(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(http.StatusText(http.StatusOK)))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileServerHits += 1
		next.ServeHTTP(w, r)
	})
}

func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) handlerReset(w http.ResponseWriter, req *http.Request) {
	cfg.fileServerHits = 0
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hits reset to 0"))
}

func (cfg *apiConfig) handlerMetrics(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`
	<html>

	<body>
		<h1>Welcome, Chirpy Admin</h1>
		<p>Chirpy has been visited %d times!</p>
	</body>

	</html>

	`, cfg.fileServerHits)))
}

func respondWithError(w http.ResponseWriter, code int, msg string) {
	if code > 499 {
		log.Printf("Responding with 5XX error: %s", msg)
	}
	type errorResponse struct {
		Error string `json:"error"`
	}
	respondWithJSON(w, code, errorResponse{
		Error: msg,
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		w.Write([]byte(serverFailed))
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}
