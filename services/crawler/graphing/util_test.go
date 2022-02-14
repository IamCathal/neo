package graphing

import (
	"testing"

	"github.com/iamcathal/neo/services/crawler/controller"
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
	jobs := []common.UsersGraphInformation{
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
	users := []common.UsersGraphInformation{
		{
			User: common.UserDocument{
				GamesOwned: []common.GameOwnedDocument{
					{AppID: 8, Playtime_Forever: 6},
					{AppID: 80, Playtime_Forever: 160},
				},
			},
		},
		{
			User: common.UserDocument{
				GamesOwned: []common.GameOwnedDocument{
					{AppID: 90, Playtime_Forever: 20},
					{AppID: 200, Playtime_Forever: 500000},
				},
			},
		},
		{
			User: common.UserDocument{
				GamesOwned: []common.GameOwnedDocument{
					{AppID: 170, Playtime_Forever: 78555},
					{AppID: 80, Playtime_Forever: 10},
				},
			},
		},
		{
			User: common.UserDocument{
				GamesOwned: []common.GameOwnedDocument{
					{AppID: 90, Playtime_Forever: 5000},
				},
			},
		},
	}
	expected := []int{200, 170, 90, 80, 8}

	actual := getTopTenMostPopularGames(users)

	assert.Equal(t, expected, actual)
}

func TestGetTopTenMostPopularGamesOnlyReturns10GamesIfMoreThan10AreGiven(t *testing.T) {
	users := []common.UsersGraphInformation{
		{
			User: common.UserDocument{
				GamesOwned: []common.GameOwnedDocument{
					{AppID: 8, Playtime_Forever: 6},
					{AppID: 80, Playtime_Forever: 160},
					{AppID: 12, Playtime_Forever: 123},
					{AppID: 164, Playtime_Forever: 14323},
					{AppID: 546, Playtime_Forever: 133},
					{AppID: 45765, Playtime_Forever: 14523},
					{AppID: 54, Playtime_Forever: 167723},
					{AppID: 895, Playtime_Forever: 12883},
				},
			},
		},
		{
			User: common.UserDocument{
				GamesOwned: []common.GameOwnedDocument{
					{AppID: 90, Playtime_Forever: 20},
					{AppID: 200, Playtime_Forever: 500000},
					{AppID: 256, Playtime_Forever: 508880},
					{AppID: 567, Playtime_Forever: 508260},
				},
			},
		},
	}

	actual := getTopTenMostPopularGames(users)

	assert.Len(t, actual, 10)
}

func TestGetTopTenOverallGameNames(t *testing.T) {
	mockController := &controller.MockCntrInterface{}
	users := []common.UsersGraphInformation{
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
	expected := []common.BareGameInfo{
		{AppID: 200, Name: "CS:GO"},
		{AppID: 80, Name: "Sunset Overdrive"},
		{AppID: 90, Name: "Deep Rock Galactic"},
	}
	mockController.On("GetGameDetailsFromIDs", mock.Anything).Return(expected, nil)

	topTenGames, err := getTopTenOverallGameNames(mockController, users)

	assert.Nil(t, err)
	assert.Equal(t, expected, topTenGames)
}
