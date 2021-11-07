package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/IamCathal/neo/services/frontend/endpoints"
	"github.com/IamCathal/neo/services/frontend/statsmonitoring"
	"github.com/joho/godotenv"
	"github.com/neosteamfriendgraphing/common/util"
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

	go statsmonitoring.CollectAndShipStats()

	router := endpoints.SetupRouter()

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%s", os.Getenv("API_PORT")),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	endpoints.Logger.Info(fmt.Sprintf("frontend start up and serving requsts on %s:%s", util.GetLocalIPAddress(), os.Getenv("API_PORT")))
	log.Fatal(srv.ListenAndServe())
}
