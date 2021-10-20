package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/iamcathal/neo/services/crawler/apikeymanager"
	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/endpoints"
)

func main() {
	err := configuration.InitConfig()
	if err != nil {
		log.Fatalf("failure initialising config: %v", err)
	}

	endpoints := &endpoints.Endpoints{
		Cntr: controller.Cntr{},
	}
	apikeymanager.InitApiKeys()

	router := endpoints.SetupRouter()

	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%s", os.Getenv("API_PORT")),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	configuration.Logger.Info(fmt.Sprintf("crawler start up and serving requsts on %s:%s", configuration.GetLocalIPAddress(), os.Getenv("API_PORT")))
	log.Fatal(srv.ListenAndServe())
}
