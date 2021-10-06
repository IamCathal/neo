package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/IamCathal/neo/services/frontend/endpoints"
	"github.com/IamCathal/neo/services/frontend/util"
	"github.com/joho/godotenv"
)

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

	logConfig := util.LoadLoggingConfig()
	// logger := util.InitLogger(logConfig)
	// logger.Info("Hello world!")

	endpoints := &endpoints.Endpoints{
		Logger:                 util.InitLogger(logConfig),
		ApplicationStartUpTime: time.Now(),
	}

	router := endpoints.SetupRouter()
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
