package app

import (
	"fmt"
	"strconv"
	"time"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/controller"
	"github.com/IamCathal/neo/services/datastore/datastructures"
	"github.com/IamCathal/neo/services/datastore/graphing"
	"github.com/neosteamfriendgraphing/common"
	commonUtil "github.com/neosteamfriendgraphing/common/util"
)

func CalulateShortestDistanceInfo(cntr controller.CntrInterface, firstCrawlID, secondCrawlID string) (bool, datastructures.ShortestDistanceInfo, error) {
	firstUserGraphData, err := cntr.GetProcessedGraphData(firstCrawlID)
	if err != nil {
		return false, datastructures.ShortestDistanceInfo{}, err
	}
	secondUserGraphData, err := cntr.GetProcessedGraphData(secondCrawlID)
	if err != nil {
		return false, datastructures.ShortestDistanceInfo{}, err
	}
	if firstUserGraphData.UserDetails.User.AccDetails.SteamID == "" ||
		secondUserGraphData.UserDetails.User.AccDetails.SteamID == "" {
		// Users have not been graphed yet
		return false, datastructures.ShortestDistanceInfo{}, nil
	}

	_, userDetailsForShortestPath, err := getUserDetailsForShortestDistancePath(cntr, firstUserGraphData, secondUserGraphData)
	if err != nil {
		return false, datastructures.ShortestDistanceInfo{}, err
	}
	uniqueFriends := getUniqueFriends(firstUserGraphData, secondUserGraphData)
	shortestDistanceInfo := datastructures.ShortestDistanceInfo{
		CrawlIDs:         []string{firstCrawlID, secondCrawlID},
		FirstUser:        firstUserGraphData.UserDetails.User,
		SecondUser:       secondUserGraphData.UserDetails.User,
		ShortestDistance: userDetailsForShortestPath,
		TotalNetworkSpan: len(uniqueFriends) + 2,
		TimeStarted:      time.Now().Unix(),
	}

	return true, shortestDistanceInfo, nil
}

func getUserDetailsForShortestDistancePath(cntr controller.CntrInterface, userOne, userTwo common.UsersGraphData) (bool, []common.UserDocument, error) {
	exists, shortestPathIDs, err := graphing.GetShortestPathIDs(cntr, userOne, userTwo)
	if err != nil {
		return false, []common.UserDocument{}, err
	}
	if !exists {
		return false, []common.UserDocument{}, nil
	}

	idToUserMap := graphing.GetIDToUserMap(userOne, userTwo)
	shortestPathUserDetails := []common.UserDocument{}
	for i := 0; i < len(shortestPathIDs); i++ {
		shortestPathUserDetails = append(shortestPathUserDetails, idToUserMap[fmt.Sprint(shortestPathIDs[i])])
	}

	return true, shortestPathUserDetails, nil
}

func getUniqueFriends(firstUserGraphData, secondUserGraphData common.UsersGraphData) []common.UserDocument {
	uniqueFriendsGraphInformation := []common.UserDocument{}
	seenIDs := make(map[string]bool)

	// Original users do not belong in the friend list
	seenIDs[firstUserGraphData.UserDetails.User.AccDetails.SteamID] = true
	seenIDs[secondUserGraphData.UserDetails.User.AccDetails.SteamID] = true

	for _, graphData := range firstUserGraphData.FriendDetails {
		if isTrue := ifIsTrue(graphData.User.AccDetails.SteamID, seenIDs); !isTrue {
			seenIDs[graphData.User.AccDetails.SteamID] = true
			uniqueFriendsGraphInformation = append(uniqueFriendsGraphInformation, graphData.User)
		}
	}
	for _, graphData := range secondUserGraphData.FriendDetails {
		if isTrue := ifIsTrue(graphData.User.AccDetails.SteamID, seenIDs); !isTrue {
			seenIDs[graphData.User.AccDetails.SteamID] = true
			uniqueFriendsGraphInformation = append(uniqueFriendsGraphInformation, graphData.User)
		}
	}
	return uniqueFriendsGraphInformation
}

func toInt64(steamID string) int64 {
	intVersion, err := strconv.ParseInt(steamID, 10, 64)
	if err != nil {
		panic(commonUtil.MakeErr(err))
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

func ifSteamIDSeenBefore(steamID int64, steamToGraph map[int64]int) bool {
	if _, exists := steamToGraph[steamID]; exists {
		return true
	}
	return false
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
