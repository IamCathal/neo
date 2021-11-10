package endpoints

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
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

var (
	testUser common.UserDocument
)

func TestMain(m *testing.M) {
	initTestData()
	c := zap.NewProductionConfig()
	logger, err := c.Build()
	if err != nil {
		log.Fatal(err)
	}
	configuration.Logger = logger

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
}

func runServer(cntr controller.CntrInterface) {
	endpoints := &Endpoints{
		Cntr: cntr,
	}
	router := endpoints.SetupRouter()
	srv := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":5150"),
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

func TestGetUser(t *testing.T) {
	mockController := &controller.MockCntrInterface{}

	go runServer(mockController)
	time.Sleep(5 * time.Millisecond)

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

	res, err := util.GetAndRead(fmt.Sprintf("http://localhost:5150/getUser/%s", testUser.SteamID))
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
