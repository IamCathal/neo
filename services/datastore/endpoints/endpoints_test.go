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
	testUser        common.UserDocument
	testSaveUserDTO dtos.SaveUserDTO
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
	os.Setenv("DATASTORE_LATENCIES_BUCKET", "testDataBucket")
	configuration.InfluxDBClient = influxdb2.NewClient(
		os.Getenv("INFLUXDB_URL"),
		os.Getenv("BUCKET_TOKEN"))
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
		time.Sleep(2 * time.Millisecond)
		cancel()
	}()
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

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/saveuser", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
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

	expectedResponse := struct {
		Error string `json:"error"`
	}{
		"cannot save crawling stats",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}
	requestBodyJSON, err := json.Marshal(testSaveUserDTO)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/saveuser", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, res.StatusCode, 400)
	assert.Equal(t, string(expectedJSONResponse)+"\n", string(body))
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

	expectedResponse := struct {
		Error string `json:"error"`
	}{
		"cannot save user",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}
	requestBodyJSON, err := json.Marshal(testSaveUserDTO)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/saveuser", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, res.StatusCode, 400)
	assert.Equal(t, string(expectedJSONResponse)+"\n", string(body))
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

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/saveuser", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
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

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/getuser/%s", serverPort, testUser.AccDetails.SteamID))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse)+"\n", string(res))
}

func TestGetUserReturnsInvalidResponseWhenGetUseFromDBReturnsAnError(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	expectedError := errors.New("couldn't get user")
	mockController.On("GetUser", mock.Anything, mock.AnythingOfType("string")).Return(common.UserDocument{}, expectedError)
	expectedResponse := struct {
		Error string `json:"error"`
	}{
		expectedError.Error(),
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/getuser/%s", serverPort, testUser.AccDetails.SteamID))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse)+"\n", string(res))
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
	mockController.On("GetCrawlingStatusFromDB", mock.Anything, mock.AnythingOfType("string")).Return(common.UserDocument{}, randomErr)

	expectedResponse := struct {
		Error string `json:"error"`
	}{
		"invalid crawlid",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/getcrawlingstatus/gobbeldygook", serverPort))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse)+"\n", string(res))
}

func TestGetCrawlingStatsReturnsCorrectCrawlingStatusWhenGivenValidCrawlID(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	expectedCrawlingStatus := common.CrawlingStatus{
		TimeStarted:         time.Now(),
		CrawlID:             ksuid.New().String(),
		OriginalCrawlTarget: "someuser",
		MaxLevel:            3,
		TotalUsersToCrawl:   1337,
		UsersCrawled:        625,
	}

	mockController.On("GetCrawlingStatusFromDB", mock.Anything, mock.Anything, mock.AnythingOfType("string")).Return(expectedCrawlingStatus, nil)

	expectedResponse := dtos.GetCrawlingStatusDTO{
		Status:         "success",
		CrawlingStatus: expectedCrawlingStatus,
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/getcrawlingstatus/%s", serverPort, expectedCrawlingStatus.CrawlID))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse)+"\n", string(res))
}

func TestGetCrawlingStatsReturnsCouldntGetCrawlingStatusWhenDBReturnsAnError(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	randomError := errors.New("hello world")
	mockController.On("GetCrawlingStatusFromDB", mock.Anything, mock.Anything, mock.AnythingOfType("string")).Return(common.CrawlingStatus{}, randomError)

	expectedResponse := struct {
		Message string `json:"error"`
	}{
		"couldn't get crawling status",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/getcrawlingstatus/%s", serverPort, ksuid.New().String()))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse)+"\n", string(res))
}
