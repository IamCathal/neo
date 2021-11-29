package controller

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/iamcathal/neo/services/crawler/apikeymanager"
	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/util"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/dtos"
	"github.com/streadway/amqp"
)

type Cntr struct{}

type CntrInterface interface {
	// Steam web API related functions
	CallGetFriends(steamID string) ([]string, error)
	CallGetPlayerSummaries(steamIDList string) ([]common.Player, error)
	CallGetOwnedGames(steamID string) (common.GamesOwnedResponse, error)
	// RabbitMQ related functions
	PublishToJobsQueue(jobJSON []byte) error
	ConsumeFromJobsQueue() (<-chan amqp.Delivery, error)
	// Datastore related functions
	SaveUserToDataStore(dtos.SaveUserDTO) (bool, error)
	GetUserFromDataStore(steamID string) (common.UserDocument, error)
	SaveCrawlingStatsToDataStore(currentLevel int, crawlingStatus common.CrawlingStatus) (bool, error)
}

// CallGetFriends calls the steam web API to retrieve a list of
// friends (steam IDs) for a given user.
// 		friendIDs, err := CallGetFriends(steamID)
func (control Cntr) CallGetFriends(steamID string) ([]string, error) {
	friendsListObj := common.UserDetails{}
	apiKey := apikeymanager.GetSteamAPIKey()
	// fmt.Println("get friends list")
	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetFriendList/v0001/?key=%s&steamid=%s",
		apiKey, steamID)
	res, err := MakeNetworkGETRequest(targetURL)
	if err != nil {
		return []string{}, err
	}
	// if valid := IsValidAPIResponseForSteamId(string(res)); !valid {
	// 	return friendsListObj, MakeErr(fmt.Errorf("invalid steamID %s given", steamID))
	// }

	// if valid := IsValidResponseForAPIKey(string(res)); !valid {
	// 	return friendsListObj, MakeErr(fmt.Errorf("invalid api key: %s", apiKey))
	// }
	json.Unmarshal(res, &friendsListObj)

	friendIDs := []string{}
	for _, friend := range friendsListObj.Friends.Friends {
		friendIDs = append(friendIDs, friend.Steamid)
	}

	return friendIDs, nil
}

// CallGetPlayerSummaries calls the steam web API to retrieve player summaries for a list
// of steamIDs. A maximum of 100 steamIDs can be handled and must be encoded like this:
// steamID5641,steamID245,steamID43,steamID5747
//		playerSummaries, err := CallGetPlayerSummaries(steamIDList)
func (control Cntr) CallGetPlayerSummaries(steamIDStringList string) ([]common.Player, error) {
	allPlayerSummaries := common.SteamAPIResponse{}
	apiKey := apikeymanager.GetSteamAPIKey()
	// fmt.Println("get player summary")

	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s",
		apiKey, steamIDStringList)
	res, err := MakeNetworkGETRequest(targetURL)
	if err != nil {
		return []common.Player{}, err
	}
	json.Unmarshal(res, &allPlayerSummaries)

	// Check if empty
	if len(allPlayerSummaries.Response.Players) == 0 {
		configuration.Logger.Sugar().Infof("empty player summary for %s: %+v", targetURL, allPlayerSummaries.Response.Players)
	}
	return allPlayerSummaries.Response.Players, nil
}

// CallGetOwnedGames calls the steam web api to retrieve all of a user's owned games
//		ownedGamesResponse, err := CallGetOwnedGames(steamID)
func (control Cntr) CallGetOwnedGames(steamID string) (common.GamesOwnedResponse, error) {
	apiResponse := common.GamesOwnedSteamResponse{}
	apiKey := apikeymanager.GetSteamAPIKey()
	// fmt.Println("get owned games")

	targetURL := fmt.Sprintf("http://api.steampowered.com/IPlayerService/GetOwnedGames/v0001/?key=%s&steamid=%s&format=json&include_appinfo=true&include_played_free_games=true",
		apiKey, steamID)
	res, err := MakeNetworkGETRequest(targetURL)
	if err != nil {
		return common.GamesOwnedResponse{}, err
	}
	json.Unmarshal(res, &apiResponse)
	return apiResponse.Response, nil
}

// PublishToJobsQueue publishes a job to the rabbitMQ queue
//		err := PublishToJobsQueue(job)
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

// ConsumeFromJobsQueue consumns a job from the rabbitMQ queue
//		msgs, err := ConsumeFromJobsQueue()
//		...
//		for job := range msgs {
//			newJob := datastructures.Job{}
//			err := json.Unmarshal(d.Body, &newJob)
//		}
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

// SaveUserToDataStore sends a user to the datastore service to be saved
// 		userWasSaved, err := SaveUserToDataStore(user)
func (control Cntr) SaveUserToDataStore(saveUser dtos.SaveUserDTO) (bool, error) {
	targetURL := fmt.Sprintf("%s/saveuser", os.Getenv("DATASTORE_URL"))
	jsonObj, err := json.Marshal(saveUser)
	if err != nil {
		return false, util.MakeErr(err)
	}
	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonObj))
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authentication", "something")

	client := &http.Client{}
	res := &http.Response{}
	var callErr error

	maxRetryCount := 3
	successfulRequest := false

	for i := 1; i <= maxRetryCount; i++ {
		res, callErr = client.Do(req)
		if callErr != nil {
			exponentialBackOffSleepTime := math.Pow(2, float64(i)) * 16
			configuration.Logger.Sugar().Infof("failed to call %s for %s %d times. Sleeping for %d ms", targetURL, saveUser.User.AccDetails.SteamID, i, exponentialBackOffSleepTime)
			time.Sleep(time.Duration(exponentialBackOffSleepTime) * time.Millisecond)
			res.Body.Close()
		} else {
			successfulRequest = true
			defer res.Body.Close()
			break
		}
	}
	// Failed after all retries
	if successfulRequest == false {
		return false, util.MakeErr(callErr)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, util.MakeErr(err)
	}
	APIRes := dtos.GetUserDTO{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		return false, util.MakeErr(err)
	}

	if res.StatusCode == 200 {
		return true, nil
	}

	return false, util.MakeErr(fmt.Errorf("error saving user: %+v", APIRes))
}

// GetUserFromDataStore gets a user from the datastore service
// 		userFromDataStore, err := GetUserFromDataStore(steamID)
func (control Cntr) GetUserFromDataStore(steamID string) (common.UserDocument, error) {
	targetURL := fmt.Sprintf("%s/getuser/%s", os.Getenv("DATASTORE_URL"), steamID)
	req, err := http.NewRequest("GET", targetURL, nil)
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authentication", "something")

	client := &http.Client{}
	res := &http.Response{}
	var callErr error
	maxRetryCount := 3
	successfulRequest := false

	for i := 1; i <= maxRetryCount; i++ {
		res, callErr = client.Do(req)
		if callErr != nil {
			configuration.Logger.Sugar().Infof("failed to call %s %d times", targetURL, i)
		} else {
			successfulRequest = true
			break
		}
	}
	// Failed after all retries
	if successfulRequest == false {
		return common.UserDocument{}, util.MakeErr(callErr)
	}

	// If no user exists in the DB (HTTP 404)
	if res.StatusCode == http.StatusNotFound {
		return common.UserDocument{}, nil
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return common.UserDocument{}, util.MakeErr(err)
	}

	userDoc := dtos.GetUserDTO{}
	err = json.Unmarshal(body, &userDoc)
	if err != nil {
		return common.UserDocument{}, util.MakeErr(err)
	}

	return userDoc.User, nil
}

func (control Cntr) SaveCrawlingStatsToDataStore(currentLevel int, crawlingStatus common.CrawlingStatus) (bool, error) {
	targetURL := fmt.Sprintf("%s/savecrawlingstats", os.Getenv("DATASTORE_URL"))
	crawlingStatsDTO := dtos.SaveCrawlingStatsDTO{
		CurrentLevel:   currentLevel,
		CrawlingStatus: crawlingStatus,
	}
	jsonObj, err := json.Marshal(crawlingStatsDTO)
	if err != nil {
		return false, util.MakeErr(err)
	}
	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonObj))
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authentication", "something")

	client := &http.Client{}
	res := &http.Response{}
	var callErr error

	maxRetryCount := 3
	successfulRequest := false

	for i := 1; i <= maxRetryCount; i++ {
		res, callErr = client.Do(req)
		if callErr != nil {
			configuration.Logger.Sugar().Infof("failed to call %s for %s %d times", targetURL, crawlingStatus.CrawlID, i)
		} else {
			successfulRequest = true
			break
		}
	}
	// Failed after all retries
	if successfulRequest == false {
		return false, util.MakeErr(callErr)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, util.MakeErr(err)
	}
	APIRes := common.BasicAPIResponse{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		return false, util.MakeErr(err)
	}

	if res.StatusCode == 200 {
		return true, nil
	}
	return false, util.MakeErr(fmt.Errorf("error saving crawling stats for existing user: %+v", APIRes))
}
