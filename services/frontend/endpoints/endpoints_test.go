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

	"github.com/IamCathal/neo/services/frontend/configuration"
	"github.com/IamCathal/neo/services/frontend/controller"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/dtos"
	"github.com/segmentio/ksuid"
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
	log, err := c.Build()
	if err != nil {
		panic(err)
	}
	configuration.Logger = log

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
	endpoints := Endpoints{
		&controller.MockCntrInterface{},
	}
	assert.HTTPStatusCode(t, endpoints.Status, "POST", "/status", nil, 200)
	assert.HTTPBodyContains(t, endpoints.Status, "POST", "/status", nil, "operational")
}

func TestCreateCrawlingStatusReturnsSuccessForValidResponseFromDataStore(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	mockController.On("SaveCrawlingStats", mock.Anything).Return(true, nil)
	expectedResponse := common.BasicAPIResponse{
		Status:  "success",
		Message: "very good",
	}
	expectedJSONResponse, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}

	crawlingStatsInput := dtos.SaveCrawlingStatsDTO{
		CurrentLevel: 0,
		CrawlingStatus: common.CrawlingStatus{
			CrawlID:             ksuid.New().String(),
			OriginalCrawlTarget: validFormatSteamID,
			MaxLevel:            3,
			TotalUsersToCrawl:   0,
			UsersCrawled:        0,
		},
	}
	requestBodyJSON, err := json.Marshal(crawlingStatsInput)
	if err != nil {
		log.Fatal(err)
	}

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/createcrawlingstatus", serverPort), "application/json", bytes.NewBuffer(requestBodyJSON))
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

func TestIsPrivateProfileReturnsCorrectProfilePrivacyType(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	expectedResponse := common.BasicAPIResponse{
		Status:  "success",
		Message: "public",
	}
	expectedResponseJSON, err := json.Marshal(expectedResponse)
	if err != nil {
		log.Fatal(err)
	}
	mockController.On("CallIsPrivateProfile", mock.AnythingOfType("string")).Return(expectedResponseJSON, nil)

	res, err := http.Get(fmt.Sprintf("http://localhost:%d/isprivateprofile/%s", serverPort, validFormatSteamID))
	if err != nil {
		log.Fatal(err)
	}
	realResponse, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	mockController.AssertNumberOfCalls(t, "CallIsPrivateProfile", 1)
	assert.Equal(t, string(expectedResponseJSON), string(realResponse))
	assert.Equal(t, 200, res.StatusCode)
}

func TestIsPrivateProfileReturnsInvalidWhenCrawlerReturnsAnError(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	emptyByteResponse := []byte{}
	randomError := errors.New("error")
	mockController.On("CallIsPrivateProfile", mock.AnythingOfType("string")).Return(emptyByteResponse, randomError)

	res, err := http.Get(fmt.Sprintf("http://localhost:%d/isprivateprofile/%s", serverPort, validFormatSteamID))
	if err != nil {
		log.Fatal(err)
	}

	mockController.AssertNumberOfCalls(t, "CallIsPrivateProfile", 1)
	assert.Equal(t, 400, res.StatusCode)
}

func TestIsPrivateProfileReturnsInvalidSteamIDWhenGivenInvalidSteamID(t *testing.T) {
	mockController, serverPort := initServerAndDependencies()

	res, err := http.Get(fmt.Sprintf("http://localhost:%d/isprivateprofile/%s", serverPort, "invalid steamID"))
	if err != nil {
		log.Fatal(err)
	}

	mockController.AssertNotCalled(t, "CallIsPrivateProfile")
	assert.Equal(t, 400, res.StatusCode)
}
