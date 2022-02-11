package graphing

import (
	"strconv"

	"github.com/neosteamfriendgraphing/common"
)

func GetIDToUserMap(userOne, userTwo common.UsersGraphData) map[string]common.UserDocument {
	idToFriendMap := make(map[string]common.UserDocument)

	idToFriendMap[userOne.UserDetails.User.AccDetails.SteamID] = userOne.UserDetails.User
	idToFriendMap[userTwo.UserDetails.User.AccDetails.SteamID] = userTwo.UserDetails.User

	for _, user := range userOne.FriendDetails {
		friend := user.User
		if exists := ifKeyExists(friend.AccDetails.SteamID, idToFriendMap); !exists {
			idToFriendMap[friend.AccDetails.SteamID] = friend
		}
	}
	for _, user := range userTwo.FriendDetails {
		friend := user.User
		if exists := ifKeyExists(friend.AccDetails.SteamID, idToFriendMap); !exists {
			idToFriendMap[friend.AccDetails.SteamID] = friend
		}
	}
	return idToFriendMap
}

func ifKeyExists(key string, idMap map[string]common.UserDocument) bool {
	if _, exists := idMap[key]; exists {
		return true
	}
	return false
}

func toInt64(steamID string) int64 {
	intVersion, err := strconv.ParseInt(steamID, 10, 64)
	if err != nil {
		panic(err)
	}
	return intVersion
}

func ifSteamIDSeenBefore(steamID int64, steamToGraph map[int64]int) bool {
	if _, exists := steamToGraph[steamID]; exists {
		return true
	}
	return false
}
