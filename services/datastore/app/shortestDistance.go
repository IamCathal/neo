package app

import (
	"fmt"
	"strconv"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/controller"
	"github.com/IamCathal/neo/services/datastore/datastructures"
	dijkstra "github.com/iamcathal/dijkstra2"
	"github.com/neosteamfriendgraphing/common"
)

func GetShortestDistanceInfo(cntr controller.CntrInterface, firstCrawlID, secondCrawlID string) (bool, datastructures.ShortestDistanceInfo, error) {
	firstUserGraphData, err := cntr.GetProcessedGraphData(firstCrawlID)
	if err != nil {
		return false, datastructures.ShortestDistanceInfo{}, err
	}
	secondUserGraphData, err := cntr.GetProcessedGraphData(secondCrawlID)
	if err != nil {
		return false, datastructures.ShortestDistanceInfo{}, err
	}

	exists, userDetailsForShortestPath, err := getUserDetailsForShortestDistancePath(firstUserGraphData, secondUserGraphData)
	if err != nil {
		return false, datastructures.ShortestDistanceInfo{}, err
	}
	if !exists {
		return false, datastructures.ShortestDistanceInfo{}, nil
	}

	shortestDistanceInfo := datastructures.ShortestDistanceInfo{
		FirstUser:        firstUserGraphData.UserDetails.User,
		SecondUser:       secondUserGraphData.UserDetails.User,
		ShortestDistance: userDetailsForShortestPath,
		TotalNetworkSpan: len(firstUserGraphData.FriendDetails) + len(secondUserGraphData.FriendDetails),
	}

	return true, shortestDistanceInfo, nil
}

func getUserDetailsForShortestDistancePath(userOne, userTwo common.UsersGraphData) (bool, []common.UserDocument, error) {
	exists, shortestPathIDs, err := getShortestDistance(userOne, userTwo)
	if err != nil {
		return false, []common.UserDocument{}, err
	}
	if !exists {
		return false, []common.UserDocument{}, nil
	}

	shortestPathUserDetails := []common.UserDocument{}
	for i := 0; i < len(shortestPathIDs); i++ {
		shortestPathUserDetails = append(shortestPathUserDetails, common.UserDocument{})
	}
	pathIDsMap := getMapOfShortestIDs(shortestPathIDs)

	shortestPathUserDetails[0] = userOne.UserDetails.User
	pathIDsMap[userOne.UserDetails.User.AccDetails.SteamID] = false

	for _, friend := range userOne.FriendDetails {
		user := friend.User
		if exists := ifIsTrue(user.AccDetails.SteamID, pathIDsMap); exists {
			pathIDsMap[user.AccDetails.SteamID] = false
			indexInShortestUserPath := indexOf(toInt64(user.AccDetails.SteamID), shortestPathIDs)
			shortestPathUserDetails[indexInShortestUserPath] = user
		}
	}
	for _, friend := range userTwo.FriendDetails {
		user := friend.User
		if exists := ifIsTrue(user.AccDetails.SteamID, pathIDsMap); exists {
			pathIDsMap[user.AccDetails.SteamID] = false
			indexInShortestUserPath := indexOf(toInt64(user.AccDetails.SteamID), shortestPathIDs)
			shortestPathUserDetails[indexInShortestUserPath] = user
		}
	}

	if isTrue := ifIsTrue(userTwo.UserDetails.User.AccDetails.SteamID, pathIDsMap); isTrue {
		shortestPathUserDetails[len(shortestPathUserDetails)-1] = userTwo.UserDetails.User
	}
	if len(shortestPathUserDetails) != len(shortestPathIDs) {
		configuration.Logger.Sugar().Panicf("could not find all user details for shortest path IDs: %+v", shortestPathIDs)
	}

	return true, shortestPathUserDetails, nil
}

func getShortestDistance(userOne, userTwo common.UsersGraphData) (bool, []int64, error) {
	graph := dijkstra.NewGraph()
	steamIDToGraphID := make(map[int64]int)
	graphIDToSteamID := make(map[int]int64)
	currGraphID := 0

	mainUserSteamID := toInt64(userOne.UserDetails.User.AccDetails.SteamID)
	targetUserSteamID := toInt64(userTwo.UserDetails.User.AccDetails.SteamID)

	steamIDToGraphID[mainUserSteamID] = currGraphID
	graphIDToSteamID[currGraphID] = mainUserSteamID
	graph.AddVertex(currGraphID)
	currGraphID++

	for _, friend := range userOne.FriendDetails {
		user := friend.User
		currUserID := toInt64(user.AccDetails.SteamID)
		steamIDToGraphID[currUserID] = currGraphID
		graphIDToSteamID[currGraphID] = currUserID
		graph.AddVertex(currGraphID)
		graph.AddArc(steamIDToGraphID[mainUserSteamID], steamIDToGraphID[currUserID], 1)
		currGraphID++
	}

	best, err := graph.Shortest(steamIDToGraphID[mainUserSteamID], steamIDToGraphID[targetUserSteamID])
	if err != nil {
		panic(err)
	}
	shortestPathSteamIDs := []int64{}
	if len(best.Path) == 0 {
		return false, []int64{}, nil
	}

	for _, graphID := range best.Path {
		shortestPathSteamIDs = append(shortestPathSteamIDs, graphIDToSteamID[graphID])
	}
	if len(shortestPathSteamIDs) != len(best.Path) {
		configuration.Logger.Sugar().Panicf("failed to get all steamIDs from best path")
	}
	return true, shortestPathSteamIDs, nil
}

func toInt64(steamID string) int64 {
	intVersion, err := strconv.ParseInt(steamID, 10, 64)
	if err != nil {
		panic(err)
	}
	return intVersion
}

func indexOf(steamID int64, IDs []int64) int {
	for i, val := range IDs {
		if val == steamID {
			return i
		}
	}
	configuration.Logger.Sugar().Panicf("failed to get index of %s in %+v", steamID, IDs)
	return -1
}

func ifIsTrue(key string, idMap map[string]bool) bool {
	if _, exists := idMap[key]; exists {
		return idMap[key]
	}
	return false
}

func getMapOfShortestIDs(steamIDs []int64) map[string]bool {
	pathIDsMap := make(map[string]bool)
	for _, ID := range steamIDs {
		pathIDsMap[fmt.Sprint(ID)] = true
	}
	return pathIDsMap
}
