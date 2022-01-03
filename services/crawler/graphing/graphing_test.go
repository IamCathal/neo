package graphing

import (
	"log"
	"os"
	"testing"

	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/neosteamfriendgraphing/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	c := zap.NewProductionConfig()
	// c.OutputPaths = []string{"/dev/null"}
	c.OutputPaths = []string{"stdout"}
	logger, err := c.Build()
	if err != nil {
		log.Fatal(err)
	}
	configuration.Logger = logger

	code := m.Run()

	os.Exit(code)
}

func TestCrawlerDoesFinish(t *testing.T) {
	mockController := &controller.MockCntrInterface{}

	firstUser := common.UserDocument{
		AccDetails: common.AccDetailsDocument{
			Personaname: "cathal",
			SteamID:     "12345",
		},
		FriendIDs: []string{
			"123456",
			"1234567",
		},
	}
	secondUser := common.UserDocument{
		AccDetails: common.AccDetailsDocument{
			Personaname: "joe",
			SteamID:     "123456",
		},
		FriendIDs: []string{
			"12345654435",
			"12112811690",
		},
	}
	thirdUser := common.UserDocument{
		AccDetails: common.AccDetailsDocument{
			Personaname: "padraic",
			SteamID:     "1234567",
		},
		FriendIDs: []string{
			"473657567587",
			"123456778910",
		},
	}

	returnedSteamIDToUsernameMap := make(map[string]string)
	returnedSteamIDToUsernameMap[secondUser.AccDetails.SteamID] = secondUser.AccDetails.Personaname
	returnedSteamIDToUsernameMap[thirdUser.AccDetails.SteamID] = thirdUser.AccDetails.Personaname

	mockController.On("GetUserFromDataStore", firstUser.AccDetails.SteamID).Return(firstUser, nil)
	mockController.On("GetUserFromDataStore", secondUser.AccDetails.SteamID).Return(secondUser, nil)
	mockController.On("GetUserFromDataStore", thirdUser.AccDetails.SteamID).Return(thirdUser, nil)
	mockController.On("GetUsernamesForSteamIDs", mock.Anything).Return(returnedSteamIDToUsernameMap, nil)
	graphWorkerConfig := GraphWorkerConfig{
		TotalUsersToCrawl: 3,
		UsersCrawled:      0,
		MaxLevel:          2,
	}

	allUsersGraphableData, err := Control2Func(mockController, firstUser.AccDetails.SteamID, graphWorkerConfig)

	assert.Nil(t, err)
	assert.Len(t, allUsersGraphableData, 3)
}
