package main

import (
	"log"
	"net/http"
)

type apiHandler struct{}

func (apiHandler) ServeHTTP(http.ResponseWriter, *http.Request) {}

func main() {
	mux := http.NewServeMux()
	// mux.Handle("/api/", apiHandler{})
	// mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
	// 	if req.URL.Path != "/" {
	// 		if req.URL.Path != "/" {
	// 			http.NotFound(w, req)
	// 			return
	// 		}
	// 		fmt.Fprintf(w, "Welcome to the home page!")
	// 	}
	// })

	corsMux := middlewareCors(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: corsMux,
	}

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
