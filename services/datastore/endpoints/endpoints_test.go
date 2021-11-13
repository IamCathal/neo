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
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/dtos"
	"github.com/neosteamfriendgraphing/common/util"
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
	initTestData()
	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"/dev/null"}
	logger, err := c.Build()
	if err != nil {
		log.Fatal(err)
	}
	configuration.Logger = logger

	rand.Seed(time.Now().UnixNano())

	code := m.Run()
	os.Exit(code)
}

func initTestData() {
	testUser = common.UserDocument{
		SteamID: "76561197969081524",
		AccDetails: common.Player{
			Steamid:                  "76561197969081524",
			Communityvisibilitystate: 3,
			Profilestate:             2,
			Personaname:              "persona name",
			Commentpermission:        0,
			Profileurl:               "profile url",
			Avatar:                   "avatar url",
			Avatarmedium:             "medium avatar",
			Avatarfull:               "full avatar",
			Avatarhash:               "avatar hash",
			Personastate:             3,
			Realname:                 "real name",
			Primaryclanid:            "clan ID",
			Timecreated:              1223525546,
			Personastateflags:        124,
			Loccountrycode:           "IE",
		},
		FriendIDs: []string{"1234", "5678"},
		GamesOwned: []common.GameInfo{
			{
				Name:            "CS:GO",
				PlaytimeForever: 1337,
				Playtime2Weeks:  50,
				ImgIconURL:      "example url",
				ImgLogoURL:      "example url",
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
	mockController := &controller.MockCntrInterface{}
	configuration.DBClient = &mongo.Client{}
	randomPort := rand.Intn(48150) + 1024

	// Start a server with this test's mock controller
	// and shutdown after 2ms
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	go runServer(mockController, ctx, randomPort)
	time.Sleep(2 * time.Millisecond)
	cancel()

	mockController.On("UpdateCrawlingStatus",
		mock.Anything,
		mock.Anything,
		testSaveUserDTO,
		len(testSaveUserDTO.User.FriendIDs),
		1).Return(true, nil)

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

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/saveuser", randomPort), "application/json", bytes.NewBuffer(requestBodyJSON))
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
	mockController := &controller.MockCntrInterface{}
	configuration.DBClient = &mongo.Client{}
	randomPort := rand.Intn(48150) + 1024

	// Start a server with this test's mock controller
	// and shutdown after 2ms
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	go runServer(mockController, ctx, randomPort)
	time.Sleep(2 * time.Millisecond)
	cancel()

	mockController.On("UpdateCrawlingStatus",
		mock.Anything,
		mock.Anything,
		testSaveUserDTO,
		len(testSaveUserDTO.User.FriendIDs),
		1).Return(false, errors.New("random error from UpdateCrawlingStatus"))

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

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/saveuser", randomPort), "application/json", bytes.NewBuffer(requestBodyJSON))
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
	mockController := &controller.MockCntrInterface{}
	configuration.DBClient = &mongo.Client{}
	randomPort := rand.Intn(48150) + 1024

	// Start a server with this test's mock controller
	// and shutdown after 2ms
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	go runServer(mockController, ctx, randomPort)
	time.Sleep(2 * time.Millisecond)
	cancel()

	mockController.On("UpdateCrawlingStatus",
		mock.Anything,
		mock.Anything,
		testSaveUserDTO,
		len(testSaveUserDTO.User.FriendIDs),
		1).Return(true, nil)

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

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/saveuser", randomPort), "application/json", bytes.NewBuffer(requestBodyJSON))
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
	mockController := &controller.MockCntrInterface{}
	configuration.DBClient = &mongo.Client{}
	randomPort := rand.Intn(48150) + 1024

	maxLeveltestUserDTO := testSaveUserDTO
	maxLeveltestUserDTO.CurrentLevel = maxLeveltestUserDTO.MaxLevel

	// Start a server with this test's mock controller
	// and shutdown after 2ms
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	go runServer(mockController, ctx, randomPort)
	time.Sleep(2 * time.Millisecond)
	cancel()

	mockController.On("UpdateCrawlingStatus",
		mock.Anything,
		mock.Anything,
		maxLeveltestUserDTO,
		0,
		1).Return(true, nil)

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

	res, err := http.Post(fmt.Sprintf("http://localhost:%d/saveuser", randomPort), "application/json", bytes.NewBuffer(requestBodyJSON))
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
	mockController := &controller.MockCntrInterface{}
	randomPort := rand.Intn(48150) + 1024

	// Start a server with this test's mock controller
	// and shutdown after 2ms
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	go runServer(mockController, ctx, randomPort)
	time.Sleep(2 * time.Millisecond)
	cancel()

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

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/getuser/%s", randomPort, testUser.SteamID))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse)+"\n", string(res))
}

func TestGetUserReturnsInvalidResponseWhenGetUseFromDBReturnsAnError(t *testing.T) {
	mockController := &controller.MockCntrInterface{}
	randomPort := rand.Intn(48150) + 1024

	// Start a server with this test's mock controller
	// and shutdown after 2ms
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	go runServer(mockController, ctx, randomPort)
	time.Sleep(2 * time.Millisecond)
	cancel()

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

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:%d/getuser/%s", randomPort, testUser.SteamID))
	if err != nil {
		log.Fatal(err)
	}

	assert.Equal(t, string(expectedJSONResponse)+"\n", string(res))

	time.Sleep(100 * time.Millisecond)
}

func TestGetUserReturnsInvalidResponseWhenGivenAnInvalidSteamID(t *testing.T) {
	mockController := &controller.MockCntrInterface{}
	endpoints := Endpoints{
		mockController,
	}
	assert.HTTPStatusCode(t, endpoints.GetUser, "GET", "/getuser/invalidsteamid", nil, 400)
}
