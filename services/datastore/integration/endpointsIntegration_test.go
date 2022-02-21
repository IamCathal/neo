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
	if err := os.Mkdir("logs", os.ModePerm); err != nil {
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

		os.RemoveAll("logs/")
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

	expectedUserAccDetails := common.AccDetailsDocument{
		SteamID:        validSteamID,
		Personaname:    "和平戰士",
		Profileurl:     "https://steamcommunity.com/id/vanek4d/",
		Avatar:         "https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/12/124898e6c4a01af630d1045ce4eb272315372115.jpg",
		Timecreated:    1369245328,
		Loccountrycode: "UA",
	}

	returnedUser := dtos.GetUserDTO{}
	authHeader := http.Header{}
	authHeader.Set("Authentication", os.Getenv("AUTH_KEY"))

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%s/api/getuser/%s", os.Getenv("API_PORT"), validSteamID), []http.Header{authHeader})
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

	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%s/api/getdetailsforgames", os.Getenv("API_PORT")), bytes.NewBuffer(requestBodyJSON))
	req.Header.Set("Authentication", os.Getenv("AUTH_KEY"))

	res, err := client.Do(req)
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
		Level:   3,
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

	client := &http.Client{}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:%s/api/hasbeencrawledbefore", os.Getenv("API_PORT")), bytes.NewBuffer(requestBodyJSON))
	req.Header.Set("Authentication", os.Getenv("AUTH_KEY"))

	res, err := client.Do(req)
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

	authHeader := http.Header{}
	authHeader.Set("Authentication", os.Getenv("AUTH_KEY"))
	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%s/api/getcrawlingstatus/%s", os.Getenv("API_PORT"), targetCrawlID), []http.Header{authHeader})
	if err != nil {
		log.Fatal(err)
	}

	assert.Nil(t, err)
	assert.Equal(t, string(expectedResponseJSON)+"\n", string(res))
}

func TestGetGraphableData(t *testing.T) {
	expectedResponse := dtos.GetGraphableDataForUserDTO{
		Username: "和平戰士",
		SteamID:  validSteamID,
		FriendIDs: []string{
			"76561198191107486",
			"76561198136510967",
			"76561198131449116",
			"76561198190732702",
			"76561198088474430",
			"76561198083409284",
			"76561198045282617",
			"76561198317155994",
			"76561198126587585",
			"76561198142760569",
			"76561198086269353",
			"76561198159247789",
			"76561198143843680",
			"76561198133307284",
			"76561198101073166",
			"76561198024839524",
			"76561198090235873",
			"76561198069919140",
			"76561198085289737",
			"76561198123717365",
			"76561198188277111",
			"76561198083190558",
			"76561198126591801",
			"76561198213247165",
			"76561198088194219",
			"76561198034583226",
			"76561198216495733",
			"76561198186273698",
			"76561198011819964",
			"76561198082232131",
			"76561198016413581",
			"76561198028922601",
			"76561198140370577",
			"76561198144084014",
			"76561198071341378",
			"76561198078629620",
			"76561197963431679",
			"76561198093076042",
			"76561198122809510",
			"76561198120025773",
			"76561198047139630",
		},
	}
	expectedResponseJSON, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	authHeader := http.Header{}
	authHeader.Set("Authentication", os.Getenv("AUTH_KEY"))
	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%s/api/getgraphabledata/%s", os.Getenv("API_PORT"), validSteamID), []http.Header{authHeader})
	if err != nil {
		log.Fatal(err)
	}

	assert.Nil(t, err)
	assert.Equal(t, string(expectedResponseJSON)+"\n", string(res))
}
