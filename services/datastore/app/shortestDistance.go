package app

import (
	"fmt"
	"strconv"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/controller"
	"github.com/IamCathal/neo/services/datastore/datastructures"
	"github.com/IamCathal/neo/services/datastore/graphing"
	"github.com/neosteamfriendgraphing/common"
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

	exists, userDetailsForShortestPath, err := getUserDetailsForShortestDistancePath(cntr, firstUserGraphData, secondUserGraphData)
	if err != nil {
		return false, datastructures.ShortestDistanceInfo{}, err
	}
	if !exists {
		return false, datastructures.ShortestDistanceInfo{}, nil
	}

	shortestDistanceInfo := datastructures.ShortestDistanceInfo{
		CrawlIDs:         []string{firstCrawlID, secondCrawlID},
		FirstUser:        firstUserGraphData.UserDetails.User,
		SecondUser:       secondUserGraphData.UserDetails.User,
		ShortestDistance: userDetailsForShortestPath,
		TotalNetworkSpan: len(firstUserGraphData.FriendDetails) + len(secondUserGraphData.FriendDetails),
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
		fmt.Println(shortestPathIDs[i])
		shortestPathUserDetails = append(shortestPathUserDetails, idToUserMap[fmt.Sprint(shortestPathIDs[i])])
	}

	return true, shortestPathUserDetails, nil
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
