package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
)

const port = "8080"

var forbiddenWords = [3]string{"kerfuffle", "sharbert", "fornax"}

func cleanChirpBody(sentence string) string {
	words := strings.Split(sentence, " ")
	cleaned := make([]string, 0)
	for i := range(words) {
		forbidden := false
		for j := range(forbiddenWords) {
			if strings.ToLower(words[i]) == forbiddenWords[j] {
				forbidden = true
			}
		}
		if forbidden {
			cleaned = append(cleaned, "****")
		} else {
			cleaned = append(cleaned, words[i])
		}
	}
	return strings.Join(cleaned, " ")
}

type apiConfig struct {
	fileserverHits atomic.Int32
}

func (a *apiConfig) metricsIncMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		a.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (a *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf(`
		<html>
  		<body>
    	<h1>Welcome, Chirpy Admin</h1>
    		<p>Chirpy has been visited %d times!</p>
  		</body>
		</html>
		`, a.fileserverHits.Load())))
}

func (a *apiConfig) metricsResetHandler(w http.ResponseWriter, r *http.Request) {
	a.fileserverHits.Store(0)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("Metrics have been reset"))
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte("OK"))
}

func validateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	type returnVals struct {
		Error string `json:"error"`
		Cleaned string `json:"cleaned_body"`
	}

	// decode json request
	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
	}

	isValidChirp := len(params.Body) <= 140
	var status int

	// encode json response
	respBody := returnVals{}
	if isValidChirp {
		respBody.Error = ""
		respBody.Cleaned = cleanChirpBody(params.Body)
		status = 200
	} else {
		respBody.Error = "Chirp is too long"
		respBody.Cleaned = cleanChirpBody(params.Body)
		status = 400
	}
	dat, err := json.Marshal(respBody)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(dat)
}

func main() {
	apiCfg := apiConfig{}
	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.metricsIncMiddleware(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.metricsResetHandler)
	mux.HandleFunc("POST /api/validate_chirp", validateChirpHandler)

	server := http.Server {
		Addr: ":" + port,
		Handler: mux,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Could not start server: %s", err)
	}
}
