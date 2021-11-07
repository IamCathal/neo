package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/iamcathal/neo/services/crawler/apikeymanager"
	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/neosteamfriendgraphing/common"
	commonUtil "github.com/neosteamfriendgraphing/common/util"
	"github.com/streadway/amqp"
)

type Cntr struct{}

type CntrInterface interface {
	// Steam web API related functions
	CallGetFriends(steamID string) (common.Friendslist, error)
	CallGetPlayerSummaries(steamIDList string) ([]common.Player, error)
	CallGetOwnedGames(steamID string) (common.GamesOwnedResponse, error)
	// RabbitMQ related functions
	PublishToJobsQueue(jobJSON []byte) error
	ConsumeFromJobsQueue() (<-chan amqp.Delivery, error)
	// Datastore related functions
	SaveFriendsListToDataStore(common.UserDetails) (bool, error)
	// HasUserBeenCrawledBefore(steamID int64) (bool, error)
}

func (control Cntr) CallGetFriends(steamID string) (common.Friendslist, error) {
	friendsList := common.Friendslist{}
	if len(steamID) > 25 {
		panic("GREATER THAN 25 wtf")
	}
	// // Check if DB has this user
	// friends, err := util.GetFriendsFromDatastore(steamID)
	// if err != nil {
	// 	return datastructures.Friendslist{}, err
	// }
	// if len(friendsList.Friends) > 0 {
	// 	return friends.FriendsList, nil
	// }

	// Call the steam web API
	friendsListObj := common.UserDetails{}
	apiKey := apikeymanager.GetSteamAPIKey()
	fmt.Println("get friends list")
	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetFriendList/v0001/?key=%s&steamid=%s",
		apiKey, steamID)
	res, err := commonUtil.GetAndRead(targetURL)
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

func (control Cntr) CallGetPlayerSummaries(steamIDStringList string) ([]common.Player, error) {
	allPlayerSummaries := common.SteamAPIResponse{}
	apiKey := apikeymanager.GetSteamAPIKey()
	fmt.Println("get player summary")

	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s",
		apiKey, steamIDStringList)
	res, err := commonUtil.GetAndRead(targetURL)
	if err != nil {
		return []common.Player{}, err
	}
	json.Unmarshal(res, &allPlayerSummaries)

	return allPlayerSummaries.Response.Players, nil
}

func (control Cntr) CallGetOwnedGames(steamID string) (common.GamesOwnedResponse, error) {
	apiResponse := common.GamesOwnedSteamResponse{}
	apiKey := apikeymanager.GetSteamAPIKey()
	fmt.Println("get owned games")

	targetURL := fmt.Sprintf("http://api.steampowered.com/IPlayerService/GetOwnedGames/v0001/?key=%s&steamid=%s&format=json&include_appinfo=true&include_played_free_games=true",
		apiKey, steamID)
	res, err := commonUtil.GetAndRead(targetURL)
	if err != nil {
		return common.GamesOwnedResponse{}, err
	}
	json.Unmarshal(res, &apiResponse)
	return apiResponse.Response, nil
}

func (control Cntr) PublishToJobsQueue(jobJSON []byte) error {
	return configuration.Channel.Publish(
		"",                       // exchange
		configuration.Queue.Name, // routing key
		false,                    // mandatory
		false,                    // immediate
		amqp.Publishing{
			ContentType: "text/json",
			Body:        jobJSON,
		})
}

func (control Cntr) ConsumeFromJobsQueue() (<-chan amqp.Delivery, error) {
	return configuration.Channel.Consume(
		configuration.Queue.Name, // queue
		"",                       // consumer
		false,                    // auto-ack
		false,                    // exclusive
		false,                    // no-local
		false,                    // no-wait
		nil,                      // args
	)
}

func (control Cntr) SaveFriendsListToDataStore(userDetails common.UserDetails) (bool, error) {
	targetURL := fmt.Sprintf("%s/saveUser", os.Getenv("DATASTORE_URL"))
	jsonObj, err := json.Marshal(userDetails)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonObj))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authentication", "something")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	APIRes := datastructures.APIResponse{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		return false, err
	}

	if res.StatusCode == 200 {
		return true, nil
	}

	return false, fmt.Errorf("error saving user: %s", APIRes.Message)
}
