package app

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/controller"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/dtos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

var (
	testSaveUserDTO dtos.SaveUserDTO
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
	testSaveUserDTO = dtos.SaveUserDTO{
		OriginalCrawlTarget: "testID",
		CurrentLevel:        1,
		MaxLevel:            3,
		User: common.UserDocument{
			SteamID: "testID",
			AccDetails: common.Player{
				Steamid:                  "testID",
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
		},
	}
}
func TestSaveUserToDBCallsMongoDB(t *testing.T) {
	mockController := &controller.MockCntrInterface{}
	mockController.On("InsertOne",
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(nil, nil)
	bson := make([]byte, 1)

	_, err := mockController.InsertOne(context.TODO(), nil, bson)

	assert.Nil(t, err)
}

func TestSaveUserToDBCallsReturnsErrorWhenMongoDoes(t *testing.T) {
	mockController := &controller.MockCntrInterface{}
	expectedError := errors.New("expected error response")
	mockController.On("InsertOne",
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(nil, expectedError)
	bson := make([]byte, 1)

	_, err := mockController.InsertOne(context.TODO(), nil, bson)

	assert.EqualError(t, err, expectedError.Error())
}

func TestSaveCrawlingStatsToDBForExistingUserAtMaxLevelOnlyCallsUpdate(t *testing.T) {
	mockController := &controller.MockCntrInterface{}
	maxLevelTestSaveUserDTO := testSaveUserDTO
	maxLevelTestSaveUserDTO.CurrentLevel = maxLevelTestSaveUserDTO.MaxLevel

	mockController.On("UpdateCrawlingStatus",
		mock.Anything,
		mock.Anything,
		maxLevelTestSaveUserDTO,
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return(true, nil)
	configuration.DBClient = &mongo.Client{}

	err := SaveCrawlingStatsToDB(mockController, maxLevelTestSaveUserDTO)

	assert.Nil(t, err)
	mockController.AssertNumberOfCalls(t, "UpdateCrawlingStatus", 1)
	mockController.AssertNotCalled(t, "InsertOne")
}

// func TestSaveCrawlingStatsToDBForNewUserCallsUpdateAndThenInsert(t *testing.T) {
// 	mockController := &controller.MockCntrInterface{}

// }
