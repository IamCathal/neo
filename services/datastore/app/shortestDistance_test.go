package app

import (
	"fmt"
	"testing"

	"github.com/neosteamfriendgraphing/common"
	"github.com/stretchr/testify/assert"
)

var (
	userOne common.UsersGraphData
	userTwo common.UsersGraphData
)

func TestGetShortestDistance(t *testing.T) {
	expectedShortestPath := []int64{
		toInt64(userOne.UserDetails.User.AccDetails.SteamID),
		toInt64(userOne.FriendDetails[0].User.AccDetails.SteamID),
		toInt64(userTwo.UserDetails.User.AccDetails.SteamID),
	}

	exists, actualShortestPath, err := getShortestDistance(userOne, userTwo)

	assert.True(t, exists)
	assert.Equal(t, expectedShortestPath, actualShortestPath)
	assert.Nil(t, err)
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

func TestToInt64(t *testing.T) {
	steamID := "123423454"
	expected := int64(123423454)

	assert.Equal(t, expected, toInt64(steamID))
}
