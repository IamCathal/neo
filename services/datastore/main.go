package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/controller"
	"github.com/IamCathal/neo/services/datastore/dbmonitor"
	"github.com/IamCathal/neo/services/datastore/endpoints"
	"github.com/IamCathal/neo/services/datastore/statsmonitoring"
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
	go dbmonitor.Monitor(endpoints.Cntr)
	router := endpoints.SetupRouter()

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%s", os.Getenv("API_PORT")),
		WriteTimeout: 20 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	configuration.Logger.Info(fmt.Sprintf("datastore start up and serving requests on %s:%s", util.GetLocalIPAddress(), os.Getenv("API_PORT")))
	log.Fatal(srv.ListenAndServe())
}
