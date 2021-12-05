package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/IamCathal/neo/services/frontend/configuration"
	"github.com/IamCathal/neo/services/frontend/controller"
	"github.com/IamCathal/neo/services/frontend/endpoints"
	"github.com/IamCathal/neo/services/frontend/statsmonitoring"
	"github.com/neosteamfriendgraphing/common/util"
)

func main() {
	err := configuration.InitConfig()
	if err != nil {
		log.Fatalf("failure initialising config: %v", err)
	}

	controller := controller.Cntr{}

	endpoints := &endpoints.Endpoints{
		Cntr: controller,
	}

	go statsmonitoring.CollectAndShipStats()

	router := endpoints.SetupRouter()

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%s", os.Getenv("API_PORT")),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	configuration.Logger.Info(fmt.Sprintf("frontend start up and serving requsts on %s:%s", util.GetLocalIPAddress(), os.Getenv("API_PORT")))
	log.Fatal(srv.ListenAndServe())
}
