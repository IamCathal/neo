package graphing

import (
	"sort"

	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/neosteamfriendgraphing/common"
)

func doesExistInMap(userMap map[string]bool, username string) bool {
	_, ok := userMap[username]
	if ok {
		return true
	}
	return false
}

func getAllSteamIDsFromJobsWithNoAssociatedUsernames(jobs []common.UsersGraphInformation) []string {
	steamIDs := []string{}
	for _, job := range jobs {
		if job.User.AccDetails.Personaname == "" {
			steamIDs = append(steamIDs, job.User.AccDetails.SteamID)
		}
	}
	return steamIDs
}

func getTopTenMostPopularGames(users []common.UsersGraphInformation) []int {
	type tempGame struct {
		appID      int
		occurances int
	}
	allGamesFrequenciesStruct := []tempGame{}
	// Find the occurancies of each game
	allGameFrequenciesMap := make(map[int]int)

	for _, user := range users {
		for _, usersGame := range user.User.GamesOwned {
			if _, ok := allGameFrequenciesMap[usersGame.AppID]; !ok {
				allGameFrequenciesMap[usersGame.AppID] = 1
			} else {
				allGameFrequenciesMap[usersGame.AppID] += 1
			}
		}
	}

	for key, val := range allGameFrequenciesMap {
		allGamesFrequenciesStruct = append(allGamesFrequenciesStruct, tempGame{appID: key, occurances: val})
	}
	sort.Slice(allGamesFrequenciesStruct, func(i, j int) bool {
		return allGamesFrequenciesStruct[i].occurances > allGamesFrequenciesStruct[j].occurances
	})

	if len(allGamesFrequenciesStruct) >= 10 {
		gameIDs := []int{}
		for _, gameID := range allGamesFrequenciesStruct[:10] {
			gameIDs = append(gameIDs, gameID.appID)
		}
		return gameIDs
	} else {
		gameIDs := []int{}
		for _, gameID := range allGamesFrequenciesStruct {
			gameIDs = append(gameIDs, gameID.appID)
		}
		return gameIDs
	}
}

func getTopTenOverallGameNames(cntr controller.CntrInterface, users []common.UsersGraphInformation) ([]common.BareGameInfo, error) {
	topTenGameIDs := getTopTenMostPopularGames(users)
	topTenGamesInfo, err := cntr.GetGameDetailsFromIDs(topTenGameIDs)
	if err != nil {
		return []common.BareGameInfo{}, err
	}
	return topTenGamesInfo, nil
}
