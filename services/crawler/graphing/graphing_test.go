package graphing

import (
	"log"
	"os"
	"testing"

	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/neosteamfriendgraphing/common/dtos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"/dev/null"}
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

	firstUserGraphableData := dtos.GetGraphableDataForUserDTO{
		Username: "cathal",
		SteamID:  "12345",
		FriendIDs: []string{
			"123456",
			"1234567",
		},
	}
	secondUserGraphableData := dtos.GetGraphableDataForUserDTO{
		Username: "joe",
		SteamID:  "123456",
		FriendIDs: []string{
			"12345654435",
			"12345673453",
		},
	}
	thirdUserGraphableData := dtos.GetGraphableDataForUserDTO{
		Username: "padraic",
		SteamID:  "1234567",
		FriendIDs: []string{
			"123456789",
			"123456778910",
		},
	}
	returnedSteamIDToUsernameMap := make(map[string]string)
	returnedSteamIDToUsernameMap[secondUserGraphableData.SteamID] = secondUserGraphableData.Username
	returnedSteamIDToUsernameMap[thirdUserGraphableData.SteamID] = thirdUserGraphableData.Username

	mockController.On("GetGraphableDataFromDataStore", firstUserGraphableData.SteamID).Return(firstUserGraphableData, nil)
	mockController.On("GetGraphableDataFromDataStore", secondUserGraphableData.SteamID).Return(secondUserGraphableData, nil)
	mockController.On("GetGraphableDataFromDataStore", thirdUserGraphableData.SteamID).Return(thirdUserGraphableData, nil)
	mockController.On("GetUsernamesForSteamIDs", mock.Anything).Return(returnedSteamIDToUsernameMap, nil)
	graphWorkerConfig := GraphWorkerConfig{
		TotalUsersToCrawl: 3,
		UsersCrawled:      0,
		MaxLevel:          2,
	}

	allUsersGraphableData, err := ControlFunc(mockController, firstUserGraphableData.SteamID, graphWorkerConfig)

	assert.Nil(t, err)
	assert.Len(t, allUsersGraphableData, 3)
}
