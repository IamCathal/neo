package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/IamCathal/neo/services/frontend/util"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func HomeHandler(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, fmt.Sprintf("%s/pages/index.html", os.Getenv("STATIC_CONTENT_DIR_NAME")))
}

func ServeGraph(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	isValidFormatGraphID, err := util.IsValidFormatGraphID(vars["graphID"])
	if err != nil || !isValidFormatGraphID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		response := struct {
			Error string `json:"error"`
		}{
			"invalid graphID given",
		}
		json.NewEncoder(w).Encode(response)
		return
	}

	http.ServeFile(w, req, fmt.Sprintf("%s/pages/%s.html", os.Getenv("STATIC_CONTENT_DIR_NAME"), vars["graphID"]))
}

func setupRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler).Methods("GET")
	r.HandleFunc("/graph/{graphID}", ServeGraph).Methods("GET")
	return r
}

func disallowFileBrowsing(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// logConfig := util.LoadLoggingConfig()
	// logger := util.InitLogger(logConfig)
	// logger.Info("Hello world!")

	router := setupRouter()
	router.Handle("/static", http.NotFoundHandler())
	fs := http.FileServer(http.Dir(os.Getenv("STATIC_CONTENT_DIR")))
	router.PathPrefix("/").Handler(http.StripPrefix("/static", disallowFileBrowsing(fs)))

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%s", os.Getenv("API_PORT")),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
