package controller

import (
	"encoding/json"
	"fmt"

	"github.com/iamcathal/neo/services/crawler/apikeymanager"
	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/iamcathal/neo/services/crawler/util"
)

type Cntr struct{}

type CntrInterface interface {
	CallGetFriends(steamID int64) (datastructures.Friendslist, error)
	// HasUserBeenCrawledBefore(steamID int64) (bool, error)
}

func (control Cntr) CallGetFriends(steamID int64) (datastructures.Friendslist, error) {
	friendsList := datastructures.Friendslist{}

	// // Check if DB has this user
	// friends, err := util.GetFriendsFromDatastore(steamID)
	// if err != nil {
	// 	return datastructures.Friendslist{}, err
	// }
	// if len(friendsList.Friends) > 0 {
	// 	return friends.FriendsList, nil
	// }

	// Call the steam web API
	friendsListObj := datastructures.UserDetails{}
	apiKey := apikeymanager.GetSteamAPIKey()
	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetFriendList/v0001/?key=%s&steamid=%d",
		apiKey, steamID)
	res, err := util.GetAndRead(targetURL)
	if err != nil {
		return friendsList, err
	}
	// if valid := IsValidAPIResponseForSteamId(string(res)); !valid {
	// 	return friendsListObj, MakeErr(fmt.Errorf("invalid steamID %s given", steamID))
	// }

	// if valid := IsValidResponseForAPIKey(string(res)); !valid {
	// 	return friendsListObj, MakeErr(fmt.Errorf("invalid api key: %s", apiKey))
	// }

	json.Unmarshal(res, &friendsListObj)
	// fmt.Printf("The object: %+v\n\n", friendsListObj)

	return friendsListObj.Friends, nil
}
