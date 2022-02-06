package endpoints

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/controller"
	"github.com/IamCathal/neo/services/datastore/datastructures"
	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/joho/godotenv"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/dtos"
	"github.com/neosteamfriendgraphing/common/util"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

var (
	testUser                        common.UserDocument
	testSaveUserDTO                 dtos.SaveUserDTO
	validFormatSteamID              = "76561197960287930"
	invalidFormatSteamID            = validFormatSteamID + "zzz"
	defaultPanicErrorMessageStarter = "Give the code monkeys this ID:"
)

func TestMain(m *testing.M) {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	initTestData()
	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"/dev/null"}
	logger, err := c.Build()
	if err != nil {
		log.Fatal(err)
	}
	configuration.Logger = logger
	createMockInfluxDBClient()

	code := m.Run()

	os.Exit(code)
}

func createMockInfluxDBClient() {
	os.Setenv("ENDPOINT_LATENCIES_BUCKET", "testDataBucket")

	configuration.InfluxDBClient = influxdb2.NewClient(
		os.Getenv("INFLUXDB_URL"),
		os.Getenv("SYSTEM_STATS_BUCKET_TOKEN"))
}

func initServerAndDependencies() (*controller.MockCntrInterface, int) {
	mockController := &controller.MockCntrInterface{}
	configuration.DBClient = &mongo.Client{}
	rand.Seed(time.Now().UnixNano())
	randomPort := rand.Intn(48150) + 1024
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	go runServer(mockController, ctx, randomPort)
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	time.Sleep(3 * time.Millisecond)
	return mockController, randomPort
}

func initTestData() {
	testUser = common.UserDocument{
		AccDetails: common.AccDetailsDocument{
			SteamID:        "76561197969081524",
			Personaname:    "persona name",
			Profileurl:     "profile url",
			Avatar:         "avatar url",
			Timecreated:    1223525546,
			Loccountrycode: "IE",
		},
		FriendIDs: []string{"1234", "5678"},
		GamesOwned: []common.GameOwnedDocument{
			{
				AppID:            102,
				Playtime_Forever: 1337,
			},
		},
	}
	testSaveUserDTO = dtos.SaveUserDTO{
		OriginalCrawlTarget: "76561197969081524",
		CurrentLevel:        2,
		MaxLevel:            3,
		User:                testUser,
	}
}

func runServer(cntr controller.CntrInterface, ctx context.Context, port int) {
	endpoints := &Endpoints{
		Cntr: cntr,
	}
	router := endpoints.SetupRouter()
	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%d", port),
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}

func TestGetAPIStatus(t *testing.T) {
	mockController := &controller.MockCntrInterface{}
	endpoints := Endpoints{
		mockController,
	}

	assert.HTTPStatusCode(t, endpoints.Status, "POST", "/status", nil, 200)
	assert.HTTPBodyContains(t, endpoints.Status, "POST", "/status", nil, "operational")
}

func TestSaveUserWithExistingUser(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	mockController.On("UpdateCrawlingStatus",
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(true, nil)

	insertResult := mongo.InsertOneResult{}
	mockController.On("InsertOne",
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(&insertResult, nil)

	expectedResponse := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		"success",
		"very good",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}
	requestBodyJSON, err := json.Marshal(testSaveUserDTO)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/saveuser", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	assert.Equal(t, res.StatusCode, 200)
	assert.Equal(t, string(expectedJSONResponse)+"\n", string(body))
}

func TestSaveUserReturnsInvalidResponseWhenSaveCrawlingStatsReturnsAnError(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	mockController.On("UpdateCrawlingStatus",
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(false, errors.New("random error from UpdateCrawlingStatus"))

	requestBodyJSON, err := json.Marshal(testSaveUserDTO)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/saveuser", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Contains(t, string(body), defaultPanicErrorMessageStarter)
}

func TestSaveUserReturnsInvalidResponseWhenSaveUserToDBReturnsAnError(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	mockController.On("UpdateCrawlingStatus",
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(true, nil)

	insertResult := mongo.InsertOneResult{}
	mockController.On("InsertOne",
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(&insertResult, errors.New("random error from SaveUserToDB"))

	requestBodyJSON, err := json.Marshal(testSaveUserDTO)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/saveuser", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Contains(t, string(body), defaultPanicErrorMessageStarter)
}

func TestSaveUserOnlyCallsUpdateCrawlingStatusIfUserIsAtMaxLevel(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	maxLeveltestUserDTO := testSaveUserDTO
	maxLeveltestUserDTO.CurrentLevel = maxLeveltestUserDTO.MaxLevel

	mockController.On("UpdateCrawlingStatus",
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(true, nil)

	singleResult := mongo.InsertOneResult{}
	mockController.On("InsertOne",
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(&singleResult, nil)

	expectedResponse := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		"success",
		"very good",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}
	requestBodyJSON, err := json.Marshal(maxLeveltestUserDTO)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/saveuser", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	mockController.AssertNumberOfCalls(t, "UpdateCrawlingStatus", 1)
	mockController.AssertNumberOfCalls(t, "InsertOne", 1)

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, string(expectedJSONResponse)+"\n", string(body))
}

func TestGetUser(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	mockController.On("GetUser", mock.Anything, mock.AnythingOfType("string")).Return(testUser, nil)
	expectedResponse := struct {
		Status string              `json:"status"`
		User   common.UserDocument `json:"user"`
	}{
		"success",
		testUser,
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/api/getuser/%s", serverPort, testUser.AccDetails.SteamID))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse)+"\n", string(res))
}

func TestGetUserReturnsInvalidResponseWhenGetUseFromDBReturnsAnError(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	getUserError := errors.New("couldn't get user")
	mockController.On("GetUser", mock.Anything, mock.AnythingOfType("string")).Return(common.UserDocument{}, getUserError)

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/api/getuser/%s", serverPort, testUser.AccDetails.SteamID))
	if err != nil {
		log.Fatal(err)
	}

	assert.Contains(t, string(res), defaultPanicErrorMessageStarter)
	fmt.Print(string(res))
}

func TestGetUserReturnsInvalidResponseWhenGivenAnInvalidSteamID(t *testing.T) {
	mockController := &controller.MockCntrInterface{}
	endpoints := Endpoints{
		mockController,
	}
	assert.HTTPStatusCode(t, endpoints.GetUser, "GET", "/getuser/invalidsteamid", nil, 400)
}

func TestGetCrawlingStatsReturnsInvalidCrawlIDWhenGivenAnInvalidID(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	randomErr := errors.New("random error")
	mockController.On("GetCrawlingStatusFromDBFromCrawlID", mock.Anything, mock.AnythingOfType("string")).Return(common.UserDocument{}, randomErr)

	expectedResponse := struct {
		Error string `json:"error"`
	}{
		"invalid crawlid",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/api/getcrawlingstatus/gobbeldygook", serverPort))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse)+"\n", string(res))
}

func TestGetCrawlingStatsReturnsCorrectCrawlingStatusWhenGivenValidCrawlID(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	expectedCrawlingStatus := common.CrawlingStatus{
		TimeStarted:         time.Now().Unix(),
		CrawlID:             ksuid.New().String(),
		OriginalCrawlTarget: "someuser",
		MaxLevel:            3,
		TotalUsersToCrawl:   1337,
		UsersCrawled:        625,
	}

	mockController.On("GetCrawlingStatusFromDBFromCrawlID", mock.Anything, mock.Anything, mock.AnythingOfType("string")).Return(expectedCrawlingStatus, nil)

	expectedResponse := dtos.GetCrawlingStatusDTO{
		Status:         "success",
		CrawlingStatus: expectedCrawlingStatus,
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/api/getcrawlingstatus/%s", serverPort, expectedCrawlingStatus.CrawlID))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse)+"\n", string(res))
}

func TestGetCrawlingStatsReturnsCouldntGetCrawlingStatusWhenDBReturnsAnError(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	randomError := errors.New("hello world")
	mockController.On("GetCrawlingStatusFromDBFromCrawlID", mock.Anything, mock.Anything, mock.AnythingOfType("string")).Return(common.CrawlingStatus{}, randomError)

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/api/getcrawlingstatus/%s", serverPort, ksuid.New().String()))
	if err != nil {
		log.Fatal(err)
	}

	assert.Contains(t, string(res), defaultPanicErrorMessageStarter)
}

func TestGetUsernamesFromSteamIDsReturnsUsernamesForSteamID(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()
	expectedUsername := "Cathal"

	expectedSteamIDsToUsernames := make(map[string]string)
	expectedSteamIDsToUsernames[validFormatSteamID] = expectedUsername

	mockController.On("GetUsernames", mock.Anything, mock.Anything, mock.AnythingOfType("string")).Return(expectedSteamIDsToUsernames, nil)

	expectedResponse := dtos.GetUsernamesFromSteamIDsDTO{
		SteamIDAndUsername: []dtos.SteamIDAndUsername{
			{
				SteamID:  validFormatSteamID,
				Username: expectedUsername,
			},
		},
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}
	requestBody := dtos.GetUsernamesFromSteamIDsInputDTO{
		SteamIDs: []string{
			validFormatSteamID,
		},
	}
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/getusernamesfromsteamids", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse)+"\n", string(body))
}

func TestGetUsernamesReturnsInvalidInputForAnyInvalidFormatSteamIDsGiven(t *testing.T) {
	_, serverPort := initServerAndDependencies()

	requestBody := dtos.GetUsernamesFromSteamIDsInputDTO{
		SteamIDs: []string{
			invalidFormatSteamID,
		},
	}
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/getusernamesfromsteamids", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestGetUsernamesFromSteamIDsReturnsInvalidRequestWhenCallToDataStoreFails(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	randomError := errors.New("hello world")
	mockController.On("GetUsernames", mock.Anything, mock.Anything, mock.AnythingOfType("string")).Return(make(map[string]string), randomError)

	requestBody := dtos.GetUsernamesFromSteamIDsInputDTO{
		SteamIDs: []string{
			validFormatSteamID,
		},
	}
	requestBodyJSON, err := json.Marshal(requestBody)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/getusernamesfromsteamids", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
}

func TestGetGraphableDataReturnsGraphableDataForAValidUser(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	foundUser := common.UserDocument{
		AccDetails: common.AccDetailsDocument{
			Personaname: "Cathal",
			SteamID:     validFormatSteamID,
		},
		FriendIDs: []string{
			"1",
			"2",
			"3",
		},
	}
	expectedResponse := dtos.GetGraphableDataForUserDTO{
		Username:  foundUser.AccDetails.Personaname,
		SteamID:   foundUser.AccDetails.SteamID,
		FriendIDs: foundUser.FriendIDs,
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	mockController.On("GetUser", mock.Anything, foundUser.AccDetails.SteamID).Return(foundUser, nil)

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/api/getgraphabledata/%s", serverPort, foundUser.AccDetails.SteamID))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse)+"\n", string(res))
}

func TestGetGraphableDataReturnsCouldntGetUserWhenNoUserIsFound(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	expectedResponse := struct {
		Error string `json:"error"`
	}{
		"couldn't get user",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}
	randomError := errors.New("hello world")
	mockController.On("GetUser", mock.Anything, validFormatSteamID).Return(common.UserDocument{}, randomError)

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/api/getgraphabledata/%s", serverPort, validFormatSteamID))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse)+"\n", string(res))
}

func TestGetGraphableDataReturnsInvalidInputForInvalidFormatSteamIDs(t *testing.T) {
	_, serverPort := initServerAndDependencies()

	expectedResponse := struct {
		Error string `json:"error"`
	}{
		"Invalid input",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/api/getgraphabledata/%s", serverPort, invalidFormatSteamID))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse)+"\n", string(res))
}

func TestSaveCrawlingStatsToDB(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	crawlingStatsInput := dtos.SaveCrawlingStatsDTO{
		CurrentLevel: 2,
		CrawlingStatus: common.CrawlingStatus{
			TimeStarted:       time.Now().Unix(),
			MaxLevel:          3,
			UsersCrawled:      5,
			TotalUsersToCrawl: 20,
		},
	}
	requestBodyJSON, err := json.Marshal(crawlingStatsInput)
	if err != nil {
		log.Fatal(err)
	}
	mockController.On("UpdateCrawlingStatus", mock.Anything, mock.Anything, mock.Anything).Return(true, nil)

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/savecrawlingstats", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestInsertGame(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()
	bareGameInfo := common.BareGameInfo{
		AppID: 10,
		Name:  "Counter-Strike",
	}
	mockController.On("InsertGame", mock.Anything, bareGameInfo).Return(true, nil)

	requestBodyJSON, err := json.Marshal(bareGameInfo)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/insertgame", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)
	mockController.AssertNumberOfCalls(t, "InsertGame", 1)
}

func TestInsertGameReturnsCouldntInsertGameWhenAnErrorOccurs(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	bareGameInfo := common.BareGameInfo{
		AppID: 10,
		Name:  "Counter-Strike",
	}
	randomError := errors.New("Bobandy")
	mockController.On("InsertGame", mock.Anything, bareGameInfo).Return(false, randomError)
	requestBodyJSON, err := json.Marshal(bareGameInfo)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/insertgame", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	mockController.AssertNumberOfCalls(t, "InsertGame", 1)
}

func TestGetDetailsForGames(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	input := dtos.GetDetailsForGamesInputDTO{
		GameIDs: []int{90, 50},
	}
	expectedReturnedGameDetails := []common.BareGameInfo{
		{
			AppID: 90,
			Name:  "Half-Life Dedicated Server",
		},
		{
			AppID: 50,
			Name:  "Half-Life: Opposing Force",
		},
	}
	expectedResponse := struct {
		Status string                `json:"status"`
		Games  []common.BareGameInfo `json:"games"`
	}{
		"success",
		expectedReturnedGameDetails,
	}
	expectedResponseJSON, _ := json.Marshal(expectedResponse)
	mockController.On("GetDetailsForGames", mock.Anything, input.GameIDs).Return(expectedReturnedGameDetails, nil)

	requestBodyJSON, err := json.Marshal(input)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/getdetailsforgames", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, string(expectedResponseJSON)+"\n", string(body))
	mockController.AssertNumberOfCalls(t, "GetDetailsForGames", 1)
}

func TestGetDetailsForGamesReturnsErrorWhenNoneOrMoreThanTwentyGamesAreRequested(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	input := dtos.GetDetailsForGamesInputDTO{
		GameIDs: []int{},
	}
	expectedResponse := struct {
		Error string `json:"error"`
	}{
		"Can only request 1-20 games in a request",
	}
	expectedResponseJSON, _ := json.Marshal(expectedResponse)

	requestBodyJSON, err := json.Marshal(input)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/getdetailsforgames", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, string(expectedResponseJSON)+"\n", string(body))
	mockController.AssertNotCalled(t, "GetDetailsForGames")
}

func TestGetDetailsForGamesReturnsAnErrorWhenGetGameDetailsReturnsAnError(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	input := dtos.GetDetailsForGamesInputDTO{
		GameIDs: []int{90, 50},
	}

	randomError := errors.New("error")
	mockController.On("GetDetailsForGames", mock.Anything, input.GameIDs).Return([]common.BareGameInfo{}, randomError)

	requestBodyJSON, err := json.Marshal(input)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/getdetailsforgames", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Contains(t, string(body), defaultPanicErrorMessageStarter)
	mockController.AssertNumberOfCalls(t, "GetDetailsForGames", 1)
}

func TestGetDetailsForGamesReturnsAnEmptyGameDetailsResponseWhenNoGameDetailsAreFound(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	input := dtos.GetDetailsForGamesInputDTO{
		GameIDs: []int{90},
	}
	expectedReturnedGameDetails := []common.BareGameInfo{
		{Name: "Deep Rock Galactic", AppID: 90},
	}
	expectedResponse := struct {
		Status string                `json:"status"`
		Games  []common.BareGameInfo `json:"games"`
	}{
		"success",
		expectedReturnedGameDetails,
	}
	expectedResponseJSON, _ := json.Marshal(expectedResponse)
	mockController.On("GetDetailsForGames", mock.Anything, input.GameIDs).Return(expectedReturnedGameDetails, nil)

	requestBodyJSON, err := json.Marshal(input)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/getdetailsforgames", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, string(expectedResponseJSON)+"\n", string(body))
	mockController.AssertNumberOfCalls(t, "GetDetailsForGames", 1)
}

func TestSaveProcessedGraphData(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	crawlID := ksuid.New().String()
	input := common.UsersGraphData{
		UserDetails: common.UsersGraphInformation{
			User: common.UserDocument{
				AccDetails: common.AccDetailsDocument{
					Personaname: "cathal",
				},
			},
		},
		FriendDetails: []common.UsersGraphInformation{
			{
				User: common.UserDocument{
					AccDetails: common.AccDetailsDocument{
						Personaname: "joe",
					},
				},
			},
			{
				User: common.UserDocument{
					AccDetails: common.AccDetailsDocument{
						Personaname: "padraic",
					},
				},
			},
		},
	}
	expectedResponse := struct {
		Status string `json:"status"`
	}{
		"success",
	}
	expectedResponseJSON, _ := json.Marshal(expectedResponse)
	mockController.On("SaveProcessedGraphData", mock.Anything, input).Return(true, nil)

	requestBodyJSON, err := json.Marshal(input)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/saveprocessedgraphdata/%s", serverPort, crawlID), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, string(expectedResponseJSON)+"\n", string(body))
	mockController.AssertNumberOfCalls(t, "SaveProcessedGraphData", 1)
}

func TestSaveProcessedGraphDataReturnsInvalidInputForInvalidFormatCrawlID(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	crawlID := "invalid format crawlID"
	expectedResponse := struct {
		Error string `json:"error"`
	}{
		"invalid input",
	}
	expectedResponseJSON, _ := json.Marshal(expectedResponse)
	mockController.On("SaveProcessedGraphData", mock.Anything, mock.Anything).Return(true, nil)

	requestBodyJSON, err := json.Marshal(common.UsersGraphData{})
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/saveprocessedgraphdata/%s", serverPort, crawlID), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, string(expectedResponseJSON)+"\n", string(body))
	mockController.AssertNotCalled(t, "SaveProcessedGraphData")
}

func TestSaveProcessedGraphDataReturnsAnErrorWhenGraphDataCannotBeRetrieved(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	crawlID := ksuid.New().String()
	err := errors.New("random error")
	mockController.On("SaveProcessedGraphData", mock.Anything, mock.Anything).Return(false, err)

	requestBodyJSON, err := json.Marshal(common.UsersGraphData{})
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/saveprocessedgraphdata/%s", serverPort, crawlID), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Contains(t, string(body), defaultPanicErrorMessageStarter)
	mockController.AssertNotCalled(t, "SaveProcessedGraphData")
}

func TestGetProcessedGraphData(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	crawlID := ksuid.New().String()
	input := common.UsersGraphData{
		UserDetails: common.UsersGraphInformation{
			User: common.UserDocument{
				AccDetails: common.AccDetailsDocument{
					Personaname: "cathal",
				},
			},
		},
		FriendDetails: []common.UsersGraphInformation{
			{
				User: common.UserDocument{
					AccDetails: common.AccDetailsDocument{
						Personaname: "joe",
					},
				},
			},
			{
				User: common.UserDocument{
					AccDetails: common.AccDetailsDocument{
						Personaname: "padraic",
					},
				},
			},
		},
	}
	expectedResponse := datastructures.GetProcessedGraphDataDTO{
		Status:        "success",
		UserGraphData: input,
	}
	expectedResponseJSON, _ := json.Marshal(expectedResponse)
	mockController.On("GetProcessedGraphData", mock.Anything, mock.Anything).Return(input, nil)

	requestBodyJSON, err := json.Marshal(common.UsersGraphData{})
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/getprocessedgraphdata/%s?responsetype=json", serverPort, crawlID), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, string(expectedResponseJSON)+"\n", string(body))
	mockController.AssertNotCalled(t, "SaveProcessedGraphData")
}

func TestGetProcessedGraphDataReturnsInvalidInputForInvalidFormatCrawlID(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	crawlID := "invalid format"
	expectedResponse := struct {
		Error string `json:"error"`
	}{
		"invalid input",
	}
	expectedResponseJSON, _ := json.Marshal(expectedResponse)

	requestBodyJSON, err := json.Marshal(common.UsersGraphData{})
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/getprocessedgraphdata/%s?responsetype=json", serverPort, crawlID), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, string(expectedResponseJSON)+"\n", string(body))
	mockController.AssertNotCalled(t, "GetProcessedGraphData")
}

func TestSaveProcessedGraphDataReturnsAnErrorWhenRetrievingGraphDataReturnsAnError(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	crawlID := ksuid.New().String()
	err := errors.New("random error")
	mockController.On("GetProcessedGraphData", mock.Anything, mock.Anything).Return(common.UsersGraphData{}, err)

	requestBodyJSON, err := json.Marshal(common.UsersGraphData{})
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/getprocessedgraphdata/%s?responsetype=json", serverPort, crawlID), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Contains(t, string(body), defaultPanicErrorMessageStarter)
	mockController.AssertNotCalled(t, "SaveProcessedGraphData")
}

func TestHasBeenCrawledBefore(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()
	expectedCrawlID := ksuid.New().String()
	input := dtos.HasBeenCrawledBeforeInputDTO{
		Level:   2,
		SteamID: validFormatSteamID,
	}
	requestBodyJSON, err := json.Marshal(input)
	if err != nil {
		log.Fatal(err)
	}

	expectedResponse := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		"success",
		expectedCrawlID,
	}
	expectedResponseJSON, _ := json.Marshal(expectedResponse)
	mockController.On("HasUserBeenCrawledBeforeAtLevel", mock.Anything, input.Level, input.SteamID).Return(expectedCrawlID, nil)
	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/hasbeencrawledbefore", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, string(expectedResponseJSON)+"\n", string(body))
	mockController.AssertNumberOfCalls(t, "HasUserBeenCrawledBeforeAtLevel", 1)
}

func TestHasBeenCrawledBeforeWithInvalidFormatSteamIDReturnsInvalidInput(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	input := dtos.HasBeenCrawledBeforeInputDTO{
		Level:   2,
		SteamID: invalidFormatSteamID,
	}
	requestBodyJSON, err := json.Marshal(input)
	if err != nil {
		log.Fatal(err)
	}

	expectedResponse := struct {
		Error string `json:"error"`
	}{
		"Invalid input",
	}
	expectedResponseJSON, _ := json.Marshal(expectedResponse)

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/hasbeencrawledbefore", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, string(expectedResponseJSON)+"\n", string(body))
	mockController.AssertNotCalled(t, "HasUserBeenCrawledBeforeAtLevel")
}

func TestHasBeenCrawledBeforeReturnsNotFoundWhenNoCrawlingStatusExists(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	input := dtos.HasBeenCrawledBeforeInputDTO{
		Level:   2,
		SteamID: validFormatSteamID,
	}
	requestBodyJSON, err := json.Marshal(input)
	if err != nil {
		log.Fatal(err)
	}

	expectedResponse := struct {
		Status  string `json:"status"`
		Message string `json:"message"`
	}{
		"success",
		"",
	}
	expectedResponseJSON, _ := json.Marshal(expectedResponse)
	mockController.On("HasUserBeenCrawledBeforeAtLevel", mock.Anything, input.Level, input.SteamID).Return("", nil)

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/hasbeencrawledbefore", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, string(expectedResponseJSON)+"\n", string(body))
	mockController.AssertNumberOfCalls(t, "HasUserBeenCrawledBeforeAtLevel", 1)
}

func TestDoesProcessedGraphDataExist(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()
	crawlID := ksuid.New().String()

	expectedResponse := dtos.DoesProcessedGraphDataExistDTO{
		Status: "success",
		Exists: "no",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}
	mockController.On("DoesProcessedGraphDataExist", crawlID).Return(false, nil)

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/doesprocessedgraphdataexist/%s", serverPort, crawlID), "application/json", nil)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Equal(t, string(expectedJSONResponse)+"\n", string(body))
}

func TestDoesProcessedGraphDataExistReturnsInvalidWhenGivenAnInvalidCrawlID(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()
	crawlID := "invalid crawlID"

	expectedResponse := struct {
		Error string `json:"error"`
	}{
		"invalid input",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/doesprocessedgraphdataexist/%s", serverPort, crawlID), "application/json", nil)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
	assert.Equal(t, string(expectedJSONResponse)+"\n", string(body))
	mockController.AssertNotCalled(t, "DoesProcessedGraphDataExist")
}

func TestDoesProcessedGraphDataExistReturnsInvalidWhenItCannotRetrieveGraphData(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()
	crawlID := ksuid.New().String()

	randomError := errors.New("random error")
	mockController.On("DoesProcessedGraphDataExist", crawlID).Return(false, randomError)

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/api/doesprocessedgraphdataexist/%s", serverPort, crawlID), "application/json", nil)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Contains(t, string(body), defaultPanicErrorMessageStarter)
	mockController.AssertNumberOfCalls(t, "DoesProcessedGraphDataExist", 1)
}

func TestGetCrawlingUser(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()
	crawlID := ksuid.New().String()

	crawlingStatus := common.CrawlingStatus{
		OriginalCrawlTarget: testUser.AccDetails.SteamID,
		TotalUsersToCrawl:   140,
		UsersCrawled:        95,
	}

	mockController.On("GetCrawlingStatusFromDBFromCrawlID", mock.Anything, crawlID).Return(crawlingStatus, nil)
	mockController.On("GetUser", mock.Anything, testUser.AccDetails.SteamID).Return(testUser, nil)

	expectedResponse := struct {
		Status string              `json:"status"`
		User   common.UserDocument `json:"user"`
	}{
		"success",
		testUser,
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/api/getcrawlinguser/%s", serverPort, crawlID))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse)+"\n", string(res))
}

func TestGetCrawlingUserReturnsNotBeingCrawledWhenUserIsNotBeingCrawled(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()
	crawlID := ksuid.New().String()

	crawlingStatus := common.CrawlingStatus{
		OriginalCrawlTarget: testUser.AccDetails.SteamID,
		TotalUsersToCrawl:   140,
		UsersCrawled:        140,
	}
	randomError := errors.New("random error")
	mockController.On("GetCrawlingStatusFromDBFromCrawlID", mock.Anything, crawlID).Return(crawlingStatus, nil)
	mockController.On("GetUser", mock.Anything, testUser.AccDetails.SteamID).Return(common.UserDocument{}, randomError)

	expectedResponse := struct {
		Error string `json:"error"`
	}{
		"User is not currently being crawled",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/api/getcrawlinguser/%s", serverPort, crawlID))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse)+"\n", string(res))
}

func TestGetCrawlingUserReturnsUserDoesNotExistWhenUserIsNotFound(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()
	crawlID := ksuid.New().String()

	crawlingStatus := common.CrawlingStatus{
		OriginalCrawlTarget: testUser.AccDetails.SteamID,
		TotalUsersToCrawl:   140,
		UsersCrawled:        90,
	}

	mockController.On("GetCrawlingStatusFromDBFromCrawlID", mock.Anything, crawlID).Return(crawlingStatus, nil)
	mockController.On("GetUser", mock.Anything, testUser.AccDetails.SteamID).Return(common.UserDocument{}, mongo.ErrNoDocuments)

	expectedResponse := struct {
		Error string `json:"error"`
	}{
		"user does not exist",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/api/getcrawlinguser/%s", serverPort, crawlID))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse)+"\n", string(res))
}
