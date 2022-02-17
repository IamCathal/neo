//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

var (
	validSteamID = "76561198092048556"
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

func TestGetDetailsForGames(t *testing.T) {
	gameIDs := []int{730, 271590}

	expectedResponse := dtos.GetDetailsForGamesDTO{
		Status: "success",
		Games: []common.BareGameInfo{
			{
				AppID: 730,
				Name:  "Counter-Strike: Global Offensive",
			},
			{
				AppID: 271590,
				Name:  "Grand Theft Auto V",
			},
		},
	}
	expectedResponseJSON, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	requestBody := dtos.GetDetailsForGamesInputDTO{
		GameIDs: gameIDs,
	}
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(fmt.Sprintf("http://localhost:%s/api/getdetailsforgames", os.Getenv("API_PORT")), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, string(expectedResponseJSON)+"\n", string(body))
}

func TestHasUserBeenCrawledBefore(t *testing.T) {

	requestBody := dtos.HasBeenCrawledBeforeInputDTO{
		SteamID: validSteamID,
		Level:   2,
	}
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(fmt.Sprintf("http://localhost:%s/api/hasbeencrawledbefore", os.Getenv("API_PORT")), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestHasUserBeenCrawledBeforeForUserThatDoesHasNotBeenCrawled(t *testing.T) {

	requestBody := dtos.HasBeenCrawledBeforeInputDTO{
		SteamID: validSteamID,
		Level:   999,
	}
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal(err)
	}
	expectedResponse := common.BasicAPIResponse{
		Status:  "success",
		Message: "",
	}
	expectedResponseJSON, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.Post(fmt.Sprintf("http://localhost:%s/api/hasbeencrawledbefore", os.Getenv("API_PORT")), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, string(expectedResponseJSON)+"\n", string(body))
}

func TestGetCrawlingStatus(t *testing.T) {
	targetCrawlID := "253v0czhdyyYWfce4LhfN1x1Nhv"

	expectedResponse := dtos.GetCrawlingStatusDTO{
		Status: "success",
		CrawlingStatus: common.CrawlingStatus{
			TimeStarted:         1644768362,
			CrawlID:             targetCrawlID,
			OriginalCrawlTarget: "76561198079437417",
			MaxLevel:            2,
			TotalUsersToCrawl:   13,
			UsersCrawled:        13,
		},
	}
	expectedResponseJSON, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%s/api/getcrawlingstatus/%s", os.Getenv("API_PORT"), targetCrawlID))
	if err != nil {
		log.Fatal(err)
	}

	assert.Nil(t, err)
	assert.Equal(t, string(expectedResponseJSON)+"\n", string(res))
}
