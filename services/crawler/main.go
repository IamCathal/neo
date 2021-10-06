package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/iamcathal/neo/services/crawler/endpoints"
	"github.com/iamcathal/neo/services/crawler/util"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	logConfig, err := util.LoadLoggingConfig()
	if err != nil {
		log.Fatal(err)
	}
	endpoints := &endpoints.Endpoints{
		Logger:                 util.InitLogger(logConfig),
		ApplicationStartUpTime: time.Now(),
	}

	router := endpoints.SetupRouter()

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%s", os.Getenv("API_PORT")),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	endpoints.Logger.Info(fmt.Sprintf("crawler start up and serving requsts on %s:%s", util.GetLocalIPAddress(), os.Getenv("API_PORT")))
	log.Fatal(srv.ListenAndServe())
}
