package main

import (
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

const port = "8080"

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
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(200)
	w.Write([]byte(fmt.Sprintf("Hits: %d", a.fileserverHits.Load())))
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

func main() {
	apiCfg := apiConfig{}
	mux := http.NewServeMux()

	mux.Handle("/app/", apiCfg.metricsIncMiddleware(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))
	mux.HandleFunc("/healthz", readinessHandler)
	mux.HandleFunc("/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("/reset", apiCfg.metricsResetHandler)

	server := http.Server {
		Addr: ":" + port,
		Handler: mux,
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Could not start server: %s", err)
	}
}
