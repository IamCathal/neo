package integration

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/controller"
	"github.com/IamCathal/neo/services/datastore/endpoints"
	"github.com/joho/godotenv"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/dtos"
	"github.com/neosteamfriendgraphing/common/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	err := configuration.InitConfig()
	if err != nil {
		log.Fatal(err)
	}
	// Override with silent logger
	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"/dev/null"}
	logger, err := c.Build()
	if err != nil {
		log.Fatal(err)
	}
	configuration.Logger = logger

	serverIsReadyChan := make(chan bool, 0)
	go initLocalServer(serverIsReadyChan)

	isReady := <-serverIsReadyChan
	if isReady {
		code := m.Run()

		os.Exit(code)
	}
}

func initLocalServer(serverIsReady chan bool) {
	controller := controller.Cntr{}
	endpoints := &endpoints.Endpoints{
		Cntr: controller,
	}
	router := endpoints.SetupRouter()
	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%s", os.Getenv("API_PORT")),
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	configuration.Logger.Info(fmt.Sprintf("datastore start up and serving requests on %s:%s", util.GetLocalIPAddress(), os.Getenv("API_PORT")))
	serverIsReady <- true
	log.Fatal(srv.ListenAndServe())
}

func TestGetKnownUser(t *testing.T) {
	targetSteamID := "76561197989571618"

	expectedUserAccDetails := common.AccDetailsDocument{
		SteamID:        targetSteamID,
		Personaname:    "madeforpvp",
		Profileurl:     "https://steamcommunity.com/profiles/76561197989571618/",
		Avatar:         "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/35/35783d9d4a03ccf95c48f22f0af350614f2330d7.jpg",
		Timecreated:    1177342433,
		Loccountrycode: "SV",
	}

	returnedUser := dtos.GetUserDTO{}
	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%s/api/getuser/%s", os.Getenv("API_PORT"), targetSteamID))
	if err != nil {
		log.Fatal(err)
	}

	json.Unmarshal(res, &returnedUser)

	assert.Nil(t, err)
	assert.Equal(t, expectedUserAccDetails, returnedUser.User.AccDetails)
}
