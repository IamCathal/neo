package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func HomeHandler(w http.ResponseWriter, req *http.Request) {
	http.ServeFile(w, req, fmt.Sprintf("%s/pages/index.html", os.Getenv("STATIC_CONTENT_DIR_NAME")))
}

func setupRouter() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler).Methods("GET")
	return r
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

	fs := http.FileServer(http.Dir(os.Getenv("STATIC_CONTENT_DIR")))
	router.PathPrefix("/").Handler(http.StripPrefix("/static", fs))

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf("127.0.0.1:%s", os.Getenv("API_PORT")),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
