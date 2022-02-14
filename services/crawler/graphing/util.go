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
		appID         int
		totalPlaytime int
	}
	allGamesTotalPlaytimesStruct := []tempGame{}
	allGamesTotalPlaytimesMap := make(map[int]int)

	for _, user := range users {
		for _, usersGame := range user.User.GamesOwned {
			if _, ok := allGamesTotalPlaytimesMap[usersGame.AppID]; !ok {
				allGamesTotalPlaytimesMap[usersGame.AppID] = usersGame.Playtime_Forever
			} else {
				allGamesTotalPlaytimesMap[usersGame.AppID] += usersGame.Playtime_Forever
			}
		}
	}

	for key, val := range allGamesTotalPlaytimesMap {
		allGamesTotalPlaytimesStruct = append(allGamesTotalPlaytimesStruct, tempGame{appID: key, totalPlaytime: val})
	}
	sort.Slice(allGamesTotalPlaytimesStruct, func(i, j int) bool {
		return allGamesTotalPlaytimesStruct[i].totalPlaytime > allGamesTotalPlaytimesStruct[j].totalPlaytime
	})

	if len(allGamesTotalPlaytimesStruct) >= 10 {
		gameIDs := []int{}
		for _, gameID := range allGamesTotalPlaytimesStruct[:10] {
			gameIDs = append(gameIDs, gameID.appID)
		}
		return gameIDs
	} else {
		gameIDs := []int{}
		for _, gameID := range allGamesTotalPlaytimesStruct {
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
