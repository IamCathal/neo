package app

import (
	"errors"
	"fmt"
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
func TestSaveUserToDBCallsMongoDBOnce(t *testing.T) {
	mockController := &controller.MockCntrInterface{}
	mockController.On("InsertOne",
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(nil, nil)
	configuration.DBClient = &mongo.Client{}

	err := SaveUserToDB(mockController, testSaveUserDTO.User)

	assert.Nil(t, err)
	mockController.AssertNumberOfCalls(t, "InsertOne", 1)
}

func TestSaveUserToDBCallsReturnsErrorWhenMongoDoes(t *testing.T) {
	mockController := &controller.MockCntrInterface{}
	expectedError := errors.New("expected error response")
	mockController.On("InsertOne",
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(nil, expectedError)
	configuration.DBClient = &mongo.Client{}

	err := SaveUserToDB(mockController, testSaveUserDTO.User)

	assert.EqualError(t, err, expectedError.Error())
	mockController.AssertNumberOfCalls(t, "InsertOne", 1)
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

func TestSaveCrawlingStatsToDBCallsUpdateAndThenInsertForNewUser(t *testing.T) {
	mockController := &controller.MockCntrInterface{}
	configuration.DBClient = &mongo.Client{}

	// Return document does not exist when trying to update it
	mockController.On("UpdateCrawlingStatus",
		mock.Anything,
		mock.Anything,
		testSaveUserDTO,
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return(false, nil)

	// Return valid for insertion of new record
	mockController.On("InsertOne",
		mock.Anything,
		mock.Anything,
		mock.Anything).Return(nil, nil)

	err := SaveCrawlingStatsToDB(mockController, testSaveUserDTO)

	assert.Nil(t, err)
	mockController.AssertNumberOfCalls(t, "UpdateCrawlingStatus", 1)
	mockController.AssertNumberOfCalls(t, "InsertOne", 1)
}

func TestSaveCrawlingStatsToDBReturnsAnErrorWhenFailsToIncrementUsersCrawledForUserOnMaxLevel(t *testing.T) {
	mockController := &controller.MockCntrInterface{}
	configuration.DBClient = &mongo.Client{}
	maxLevelTestSaveUserDTO := testSaveUserDTO
	maxLevelTestSaveUserDTO.CurrentLevel = maxLevelTestSaveUserDTO.MaxLevel
	expectedError := fmt.Errorf("failed to increment userscrawled on last level for DTO: '%+v'", maxLevelTestSaveUserDTO)

	// Return document does not exist when trying to update it
	mockController.On("UpdateCrawlingStatus",
		mock.Anything,
		mock.Anything,
		maxLevelTestSaveUserDTO,
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return(false, nil).Once()

	// Return an error when this max level user cannot be updated
	mockController.On("UpdateCrawlingStatus",
		mock.Anything,
		mock.Anything,
		maxLevelTestSaveUserDTO,
		mock.AnythingOfType("int"),
		mock.AnythingOfType("int")).Return(false, nil).Once()

	err := SaveCrawlingStatsToDB(mockController, maxLevelTestSaveUserDTO)

	assert.EqualError(t, err, expectedError.Error())
	mockController.AssertNumberOfCalls(t, "UpdateCrawlingStatus", 1)
	mockController.AssertNotCalled(t, "InsertOne")
}
