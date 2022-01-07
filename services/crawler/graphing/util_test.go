package graphing

import (
	"testing"

	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/neosteamfriendgraphing/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDoesExistInMap(t *testing.T) {
	testMap := make(map[string]bool)
	testMap["Aldi"] = true
	exists := doesExistInMap(testMap, "Aldi")

	assert.True(t, exists)
}

func TestDoesExistInMapReturnsFalseForNonExistantElement(t *testing.T) {
	testMap := make(map[string]bool)
	exists := doesExistInMap(testMap, "Aldi")

	assert.False(t, exists)
}

func TestGetAllSteamIDsFromJobsWithNoAssociatedUsernames(t *testing.T) {
	jobs := []datastructures.ResStruct{
		{
			User: common.UserDocument{
				AccDetails: common.AccDetailsDocument{
					Personaname: "",
					SteamID:     "1234567",
				},
			},
		},
		{
			User: common.UserDocument{
				AccDetails: common.AccDetailsDocument{
					Personaname: "padraic",
					SteamID:     "1444567",
				},
			},
		},
	}
	expected := []string{jobs[0].User.AccDetails.SteamID}
	actual := getAllSteamIDsFromJobsWithNoAssociatedUsernames(jobs)

	assert.Equal(t, expected, actual)
}

func TestGetTopTenMostPopularGames(t *testing.T) {
	users := []datastructures.ResStruct{
		{
			User: common.UserDocument{
				GamesOwned: []common.GameOwnedDocument{
					{AppID: 90},
					{AppID: 80},
				},
			},
		},
		{
			User: common.UserDocument{
				GamesOwned: []common.GameOwnedDocument{
					{AppID: 90},
					{AppID: 200},
				},
			},
		},
		{
			User: common.UserDocument{
				GamesOwned: []common.GameOwnedDocument{
					{AppID: 90},
					{AppID: 80},
				},
			},
		},
		{
			User: common.UserDocument{
				GamesOwned: []common.GameOwnedDocument{
					{AppID: 90},
				},
			},
		},
	}
	expected := []int{90, 80, 200}

	actual := getTopTenMostPopularGames(users)

	assert.Equal(t, expected, actual)
}

func TestGetTopTenOverallGameNames(t *testing.T) {
	mockController := &controller.MockCntrInterface{}
	users := []datastructures.ResStruct{
		{
			User: common.UserDocument{
				GamesOwned: []common.GameOwnedDocument{
					{AppID: 90},
					{AppID: 80},
				},
			},
		},
		{
			User: common.UserDocument{
				GamesOwned: []common.GameOwnedDocument{
					{AppID: 90},
					{AppID: 200},
				},
			},
		},
		{
			User: common.UserDocument{
				GamesOwned: []common.GameOwnedDocument{
					{AppID: 90},
					{AppID: 80},
				},
			},
		},
		{
			User: common.UserDocument{
				GamesOwned: []common.GameOwnedDocument{
					{AppID: 90},
				},
			},
		},
	}
	expected := []datastructures.BareGameInfo{
		{AppID: 200, Name: "CS:GO"},
		{AppID: 80, Name: "Sunset Overdrive"},
		{AppID: 90, Name: "Deep Rock Galactic"},
	}
	mockController.On("GetGameDetailsFromIDs", mock.Anything).Return(expected, nil)

	topTenGames, err := getTopTenOverallGameNames(mockController, users)

	assert.Nil(t, err)
	assert.Equal(t, expected, topTenGames)
}
