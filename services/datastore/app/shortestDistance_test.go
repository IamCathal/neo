package app

import (
	"fmt"
	"testing"

	"github.com/IamCathal/neo/services/datastore/controller"
	"github.com/IamCathal/neo/services/datastore/datastructures"
	"github.com/neosteamfriendgraphing/common"
	"github.com/segmentio/ksuid"
	"github.com/stretchr/testify/assert"
)

func TestGetShortestDistanceWithTwoUsersWhoAreDirectFriendsWithEachother(t *testing.T) {
	mockController := &controller.MockCntrInterface{}

	firstUserCrawlID := ksuid.New().String()
	secondUserCrawlID := ksuid.New().String()

	expectedShortestPathInfo := datastructures.ShortestDistanceInfo{
		CrawlIDs:   []string{firstUserCrawlID, secondUserCrawlID},
		FirstUser:  userOneGraphData.UserDetails.User,
		SecondUser: userTwoGraphData.UserDetails.User,
		ShortestDistance: []common.UserDocument{
			userOneGraphData.UserDetails.User,
			userTwoGraphData.UserDetails.User,
		},
		TotalNetworkSpan: 2,
	}

	mockController.On("GetProcessedGraphData", firstUserCrawlID).Return(userOneGraphData, nil)
	mockController.On("GetProcessedGraphData", secondUserCrawlID).Return(userTwoGraphData, nil)

	exists, actualShortestPathInfo, err := CalulateShortestDistanceInfo(
		mockController,
		firstUserCrawlID,
		secondUserCrawlID)

	assert.True(t, exists)
	assert.Equal(t, expectedShortestPathInfo.CrawlIDs, actualShortestPathInfo.CrawlIDs)
	assert.Equal(t, expectedShortestPathInfo.FirstUser, actualShortestPathInfo.FirstUser)
	assert.Equal(t, expectedShortestPathInfo.SecondUser, actualShortestPathInfo.SecondUser)
	assert.Equal(t, expectedShortestPathInfo.ShortestDistance, actualShortestPathInfo.ShortestDistance)
	assert.Equal(t, expectedShortestPathInfo.TotalNetworkSpan, actualShortestPathInfo.TotalNetworkSpan)
	assert.Nil(t, err)
}

func TestGetShortestDistanceWithTwoUsersWhoShareOneCommonFriend(t *testing.T) {
	mockController := &controller.MockCntrInterface{}

	firstUserCrawlID := ksuid.New().String()
	secondUserCrawlID := ksuid.New().String()

	expectedShortestPathInfo := datastructures.ShortestDistanceInfo{
		CrawlIDs:   []string{firstUserCrawlID, secondUserCrawlID},
		FirstUser:  userOneWithOneSharedCommonFriendGraphData.UserDetails.User,
		SecondUser: userTwoWithOneSharedCommonFriendGraphData.UserDetails.User,
		ShortestDistance: []common.UserDocument{
			userOneWithOneSharedCommonFriendGraphData.UserDetails.User,
			commonFriendGraphData.UserDetails.User,
			userTwoWithOneSharedCommonFriendGraphData.UserDetails.User,
		},
		TotalNetworkSpan: 3,
	}

	mockController.On("GetProcessedGraphData", firstUserCrawlID).Return(userOneWithOneSharedCommonFriendGraphData, nil)
	mockController.On("GetProcessedGraphData", secondUserCrawlID).Return(userTwoWithOneSharedCommonFriendGraphData, nil)

	exists, actualShortestPathInfo, err := CalulateShortestDistanceInfo(
		mockController,
		firstUserCrawlID,
		secondUserCrawlID)

	assert.True(t, exists)
	assert.Equal(t, expectedShortestPathInfo.CrawlIDs, actualShortestPathInfo.CrawlIDs)
	assert.Equal(t, expectedShortestPathInfo.FirstUser, actualShortestPathInfo.FirstUser)
	assert.Equal(t, expectedShortestPathInfo.SecondUser, actualShortestPathInfo.SecondUser)
	assert.Equal(t, expectedShortestPathInfo.ShortestDistance, actualShortestPathInfo.ShortestDistance)
	assert.Equal(t, expectedShortestPathInfo.TotalNetworkSpan, actualShortestPathInfo.TotalNetworkSpan)
	assert.Nil(t, err)
}

func TestGetUniqueFriends(t *testing.T) {
	expectedUniqueFriends := []common.UserDocument{commonFriendGraphData.UserDetails.User}

	uniqueFriends := getUniqueFriends(
		userOneWithOneSharedCommonFriendGraphData,
		userTwoWithOneSharedCommonFriendGraphData)

	assert.Equal(t, expectedUniqueFriends, uniqueFriends)
}

func TestIfSteamIDSeenBefore(t *testing.T) {
	steamID := int64(1234325425345)
	steamIDToGraphIDMap := make(map[int64]int)
	steamIDToGraphIDMap[steamID] = 0

	actualSeenBefore := ifSteamIDSeenBefore(steamID, steamIDToGraphIDMap)

	assert.True(t, actualSeenBefore)
}

func TestGetMapOfShortestIDs(t *testing.T) {
	IDsSlice := []int64{1, 15, 78}
	expectedMap := make(map[string]bool)
	for _, val := range IDsSlice {
		expectedMap[fmt.Sprint(val)] = true
	}

	actualMap := getMapOfShortestIDs(IDsSlice)

	assert.Equal(t, expectedMap, actualMap)
}

func TestIfIsTrue(t *testing.T) {
	steamID := 12314234535354
	myMap := make(map[string]bool)
	myMap[fmt.Sprint(steamID)] = true

	assert.True(t, ifIsTrue(fmt.Sprint(steamID), myMap))
}

func TestIndexOf(t *testing.T) {
	mySlice := []int64{56, 23, 1337, 263}
	expectedIndex := 2

	assert.Equal(t, expectedIndex, indexOf(1337, mySlice))
}

func TestGetIndexOfNonExistantElement(t *testing.T) {
	defer func() {
		if err := recover(); err == nil {
			assert.Fail(t, "Did not fail to get index of non existant index")
		}
	}()

	steamIDs := []int64{1, 2, 3, 4}
	findIndexOf := 1337

	_ = indexOf(int64(findIndexOf), steamIDs)

}

func TestToInt64(t *testing.T) {
	steamID := "123423454"
	expected := int64(123423454)

	assert.Equal(t, expected, toInt64(steamID))
}
