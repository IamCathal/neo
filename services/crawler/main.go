package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/iamcathal/neo/services/crawler/apikeymanager"
	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/endpoints"
	"github.com/iamcathal/neo/services/crawler/statsmonitoring"
	"github.com/iamcathal/neo/services/crawler/worker"
	commonUtil "github.com/neosteamfriendgraphing/common/util"
)

func main() {
	err := configuration.InitConfig()
	if err != nil {
		panic(fmt.Sprintf("failure initialising config: %v", err))
	}

	controller := controller.Cntr{}

	endpoints := &endpoints.Endpoints{
		Cntr: controller,
	}

	var waitG sync.WaitGroup
	waitG.Add(2)
	go apikeymanager.InitApiKeys(&waitG)
	go worker.StartUpWorkers(controller, &waitG)
	waitG.Wait()

	go statsmonitoring.CollectAndShipStats()
	router := endpoints.SetupRouter()

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%s", os.Getenv("API_PORT")),
		WriteTimeout: 20 * time.Second,
		ReadTimeout:  20 * time.Second,
	}
	configuration.Logger.Info(fmt.Sprintf("crawler start up and serving requests on %s:%s", commonUtil.GetLocalIPAddress(), os.Getenv("API_PORT")))
	log.Fatal(srv.ListenAndServe())
}
