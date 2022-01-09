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

	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/iamcathal/neo/services/crawler/util"
	"github.com/neosteamfriendgraphing/common"
	"github.com/segmentio/ksuid"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

var (
	validFormatSteamID = "76561197960287930"
)

func TestMain(m *testing.M) {
	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"/dev/null"}
	logger, err := c.Build()
	if err != nil {
		log.Fatal(err)
	}
	configuration.Logger = logger
	configuration.AmqpChannels = append(configuration.AmqpChannels, amqp.Channel{})

	code := m.Run()

	os.Exit(code)
}

func initServerAndDependencies() (*controller.MockCntrInterface, int) {
	mockController := &controller.MockCntrInterface{}
	rand.Seed(time.Now().UnixNano())
	randomPort := rand.Intn(48150) + 1024

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	go runServer(mockController, ctx, randomPort)
	go func() {
		time.Sleep(15 * time.Millisecond)
		cancel()
	}()
	time.Sleep(1 * time.Millisecond)
	return mockController, randomPort
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

func TestIsPrivateProfileWithPublicProfileReturnsPublic(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	testFriendsList := []string{"1234", "5467"}
	mockController.On("CallGetFriends", mock.AnythingOfType("string")).Return(testFriendsList, nil)
	expectedResponse := common.BasicAPIResponse{
		Status:  "success",
		Message: "public",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Get(fmt.Sprintf("http://localhost:%d/isprivateprofile/%s", serverPort, validFormatSteamID))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, res.StatusCode, 200)
	assert.Equal(t, string(expectedJSONResponse), string(body))
}

func TestIsPrivateProfileReturnsInvalidResponseWithInvalidFormatSteamID(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	res, err := http.Get(fmt.Sprintf("http://localhost:%d/isprivateprofile/%s", serverPort, "invalid format steamID"))
	if err != nil {
		log.Fatal(err)
	}

	mockController.AssertNotCalled(t, "CallGetFriends")
	assert.Equal(t, 400, res.StatusCode)
}

func TestIsPrivateProfileReturnsInvalidResponseWhenCallGetFriendsReturnsAnError(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	randomError := errors.New("hello world")
	mockController.On("CallGetFriends", mock.AnythingOfType("string")).Return([]string{}, randomError)

	res, err := http.Get(fmt.Sprintf("http://localhost:%d/isprivateprofile/%s", serverPort, validFormatSteamID))
	if err != nil {
		log.Fatal(err)
	}

	mockController.AssertCalled(t, "CallGetFriends", validFormatSteamID)
	assert.Equal(t, 400, res.StatusCode)
}

func TestCrawlOneValidUser(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	userCrawlInput := datastructures.CrawlUsersInput{
		FirstSteamID: validFormatSteamID,
		Level:        3,
	}
	requestBodyJSON, err := json.Marshal(userCrawlInput)
	if err != nil {
		log.Fatal(err)
	}

	mockController.On("PublishToJobsQueue", mock.Anything, mock.Anything).Return(nil)

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/crawl", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	apiResponse := common.BasicAPIResponse{}
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "success", apiResponse.Status)
}

func TestCrawlUserReturnsInvalidLevelGivenWhenItGetsInvalidInput(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	invalidUserCrawlInput := common.BasicAPIResponse{
		Status:  "error",
		Message: "Ribena",
	}
	requestBodyJSON, err := json.Marshal(invalidUserCrawlInput)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(fmt.Sprintf("http://localhost:%d/crawl", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	expectedResponse := struct {
		Error string `json:"error"`
	}{
		"Invalid level given",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	mockController.AssertNotCalled(t, "CallGetFriends")
	assert.Equal(t, 400, res.StatusCode)
	assert.Equal(t, string(expectedJSONResponse)+"\n", string(body))
}

func TestCrawlUserReturnsInvalidFormatSteamIDsForInvalidSteamIDs(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	invalidUserCrawlInput := datastructures.CrawlUsersInput{
		FirstSteamID: "uachtar reoite",
		Level:        3,
	}
	requestBodyJSON, err := json.Marshal(invalidUserCrawlInput)
	if err != nil {
		log.Fatal(err)
	}
	res, err := http.Post(fmt.Sprintf("http://localhost:%d/crawl", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	expectedResponse := struct {
		Error string `json:"error"`
	}{
		"No valid format steamIDs given",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	mockController.AssertNotCalled(t, "CallGetFriends")
	assert.Equal(t, 400, res.StatusCode)
	assert.Equal(t, string(expectedJSONResponse)+"\n", string(body))
}

func TestCreateGraph(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	returnedCrawlingStatus := datastructures.CrawlingStatus{
		TimeStarted: time.Now().Unix(),
		CrawlID:     ksuid.New().String(),
	}
	mockController.On("GetCrawlingStatsFromDataStore", returnedCrawlingStatus.CrawlID).Return(returnedCrawlingStatus, nil)

	expectedResponse := common.BasicAPIResponse{
		Status:  "success",
		Message: "graph creation has been initiated",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(util.MakeErr(err))
	}

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/creategraph/%s", serverPort, returnedCrawlingStatus.CrawlID), "application/json", nil)
	if err != nil {
		log.Fatal(err)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse), string(body))
}
