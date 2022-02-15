package controller

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"time"

	"github.com/iamcathal/neo/services/crawler/apikeymanager"
	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/dtos"
	commonUtil "github.com/neosteamfriendgraphing/common/util"
	"github.com/streadway/amqp"
)

type Cntr struct{}

type CntrInterface interface {
	// Steam web API related functions
	CallGetFriends(steamID string) ([]string, error)
	CallGetPlayerSummaries(steamIDList string) ([]common.Player, error)
	CallGetOwnedGames(steamID string) (common.GamesOwnedResponse, error)
	// RabbitMQ related functions
	PublishToJobsQueue(channel amqp.Channel, jobJSON []byte) error
	ConsumeFromJobsQueue() (<-chan amqp.Delivery, error)
	// Datastore related functions
	SaveUserToDataStore(dtos.SaveUserDTO) (bool, error)
	GetUserFromDataStore(steamID string) (common.UserDocument, error)
	SaveCrawlingStatsToDataStore(currentLevel int, crawlingStatus common.CrawlingStatus) (bool, error)
	GetCrawlingStatsFromDataStore(crawlID string) (common.CrawlingStatus, error)
	GetGraphableDataFromDataStore(steamID string) (dtos.GetGraphableDataForUserDTO, error)
	GetUsernamesForSteamIDs(steamIDs []string) (map[string]string, error)
	SaveProcessedGraphDataToDataStore(crawlID string, graphData common.UsersGraphData) (bool, error)
	GetGameDetailsFromIDs(gameIDs []int) ([]common.BareGameInfo, error)

	Sleep(duration time.Duration)
}

// CallGetFriends calls the steam web API to retrieve a list of
// friends (steam IDs) for a given user.
// 		friendIDs, err := CallGetFriends(steamID)
func (control Cntr) CallGetFriends(steamID string) ([]string, error) {
	friendsListObj := common.UserDetails{}
	apiKey := apikeymanager.GetSteamAPIKey()
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
	maxRetryCount := 3
	successfulRequest := false

	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s",
		apiKey, steamIDStringList)
	res, err := MakeNetworkGETRequest(targetURL)
	if err != nil {
		configuration.Logger.Info("error from first call to getPlayerSummaries, retrying")
		for i := 0; i < maxRetryCount; i++ {
			res, err = MakeNetworkGETRequest(targetURL)
			if err == nil {
				configuration.Logger.Sugar().Infof("success on the %d request", i)
				successfulRequest = true
				break
			}
			exponentialBackOffSleepTime := math.Pow(2, float64(i)) * 12
			configuration.Logger.Sugar().Infof("failed to call get player summaries (%s) %d times. Sleeping for %d ms", steamIDStringList, i, exponentialBackOffSleepTime)
			time.Sleep(time.Duration(exponentialBackOffSleepTime) * time.Millisecond)
		}
	} else {
		successfulRequest = true
	}

	if successfulRequest == false {
		return []common.Player{}, err
	}

	json.Unmarshal(res, &allPlayerSummaries)
	// Check if empty
	if len(allPlayerSummaries.Response.Players) == 0 {
		configuration.Logger.Sugar().Panicf("empty player summary for %s: %+v", targetURL, allPlayerSummaries.Response.Players)
	}
	return allPlayerSummaries.Response.Players, nil
}

// CallGetOwnedGames calls the steam web api to retrieve all of a user's owned games
//		ownedGamesResponse, err := CallGetOwnedGames(steamID)
func (control Cntr) CallGetOwnedGames(steamID string) (common.GamesOwnedResponse, error) {
	apiResponse := common.GamesOwnedSteamResponse{}
	apiKey := apikeymanager.GetSteamAPIKey()

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
func (control Cntr) PublishToJobsQueue(channel amqp.Channel, jobJSON []byte) error {
	return channel.Publish(
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
	return configuration.ConsumeChannel.Consume(
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
	targetURL := fmt.Sprintf("%s/api/saveuser", os.Getenv("DATASTORE_URL"))
	jsonObj, err := json.Marshal(saveUser)
	if err != nil {
		return false, commonUtil.MakeErr(err)
	}
	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonObj))
	if err != nil {
		return false, err
	}
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
	if !successfulRequest {
		return false, commonUtil.MakeErr(callErr)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, commonUtil.MakeErr(err)
	}
	APIRes := dtos.GetUserDTO{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		return false, commonUtil.MakeErr(err)
	}

	if res.StatusCode == 200 {
		return true, nil
	}

	return false, commonUtil.MakeErr(fmt.Errorf("error saving user: %+v", APIRes))
}

// GetUserFromDataStore gets a user from the datastore service
// 		userFromDataStore, err := GetUserFromDataStore(steamID)
func (control Cntr) GetUserFromDataStore(steamID string) (common.UserDocument, error) {
	targetURL := fmt.Sprintf("%s/api/getuser/%s", os.Getenv("DATASTORE_URL"), steamID)
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return common.UserDocument{}, err
	}
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
			res.Body.Close()
		} else {
			successfulRequest = true
			defer res.Body.Close()
			break
		}
	}
	// Failed after all retries
	if !successfulRequest {
		return common.UserDocument{}, commonUtil.MakeErr(callErr)
	}

	// If no user exists in the DB (HTTP 404)
	if res.StatusCode == http.StatusNotFound {
		return common.UserDocument{}, nil
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return common.UserDocument{}, commonUtil.MakeErr(err)
	}

	userDoc := dtos.GetUserDTO{}
	err = json.Unmarshal(body, &userDoc)
	if err != nil {
		return common.UserDocument{}, commonUtil.MakeErr(err)
	}

	return userDoc.User, nil
}

func (control Cntr) SaveCrawlingStatsToDataStore(currentLevel int, crawlingStatus common.CrawlingStatus) (bool, error) {
	targetURL := fmt.Sprintf("%s/api/savecrawlingstats", os.Getenv("DATASTORE_URL"))
	crawlingStatsDTO := dtos.SaveCrawlingStatsDTO{
		CurrentLevel:   currentLevel,
		CrawlingStatus: crawlingStatus,
	}
	jsonObj, err := json.Marshal(crawlingStatsDTO)
	if err != nil {
		return false, commonUtil.MakeErr(err)
	}
	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonObj))
	if err != nil {
		return false, err
	}
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
			res.Body.Close()
		} else {
			successfulRequest = true
			defer res.Body.Close()
			break
		}
	}
	// Failed after all retries
	if !successfulRequest {
		return false, commonUtil.MakeErr(callErr)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, commonUtil.MakeErr(err)
	}
	APIRes := common.BasicAPIResponse{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		return false, commonUtil.MakeErr(err)
	}

	if res.StatusCode == 200 {
		return true, nil
	}
	return false, commonUtil.MakeErr(fmt.Errorf("error saving crawling stats for existing user: %+v", APIRes))
}

func (control Cntr) GetCrawlingStatsFromDataStore(crawlID string) (common.CrawlingStatus, error) {
	targetURL := fmt.Sprintf("%s/api/getcrawlingstatus/%s", os.Getenv("DATASTORE_URL"), crawlID)
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return common.CrawlingStatus{}, err
	}
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
			configuration.Logger.Sugar().Infof("failed to call %s for %s %d times", targetURL, crawlID, i)
			res.Body.Close()
		} else {
			successfulRequest = true
			defer res.Body.Close()
			break
		}
	}
	// Failed after all retries
	if !successfulRequest {
		return common.CrawlingStatus{}, commonUtil.MakeErr(callErr)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return common.CrawlingStatus{}, commonUtil.MakeErr(err)
	}
	APIRes := dtos.GetCrawlingStatusDTO{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		return common.CrawlingStatus{}, commonUtil.MakeErr(err)
	}

	if res.StatusCode == 200 {
		return APIRes.CrawlingStatus, nil
	}
	return common.CrawlingStatus{}, commonUtil.MakeErr(fmt.Errorf("error getting crawling status: %+v", APIRes))
}

func (control Cntr) GetGraphableDataFromDataStore(steamID string) (dtos.GetGraphableDataForUserDTO, error) {
	targetURL := fmt.Sprintf("%s/api/getgraphabledata/%s", os.Getenv("DATASTORE_URL"), steamID)
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return dtos.GetGraphableDataForUserDTO{}, err
	}
	req.Close = true
	req.Header.Set("Authentication", "something")

	client := &http.Client{}
	res := &http.Response{}
	var callErr error

	maxRetryCount := 3
	successfulRequest := false

	for i := 1; i <= maxRetryCount; i++ {
		res, callErr = client.Do(req)
		if callErr != nil {
			configuration.Logger.Sugar().Infof("failed to call %s for %s %d times", targetURL, steamID, i)
			res.Body.Close()
		} else {
			successfulRequest = true
			defer res.Body.Close()
			break
		}
	}
	// Failed after all retries
	if !successfulRequest {
		return dtos.GetGraphableDataForUserDTO{}, commonUtil.MakeErr(callErr)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return dtos.GetGraphableDataForUserDTO{}, commonUtil.MakeErr(err)
	}
	APIRes := dtos.GetGraphableDataForUserDTO{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		return dtos.GetGraphableDataForUserDTO{}, commonUtil.MakeErr(err)
	}

	if res.StatusCode == 200 {
		return APIRes, nil
	}
	return dtos.GetGraphableDataForUserDTO{}, commonUtil.MakeErr(fmt.Errorf("error getting crawling status: %+v", APIRes))
}

func (control Cntr) GetUsernamesForSteamIDs(steamIDs []string) (map[string]string, error) {
	targetURL := fmt.Sprintf("%s/api/getusernamesfromsteamids", os.Getenv("DATASTORE_URL"))
	steamIDsInput := dtos.GetUsernamesFromSteamIDsInputDTO{
		SteamIDs: steamIDs,
	}
	jsonObj, err := json.Marshal(steamIDsInput)
	if err != nil {
		return make(map[string]string), commonUtil.MakeErr(err)
	}

	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonObj))
	if err != nil {
		return make(map[string]string), err
	}
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
			configuration.Logger.Sugar().Infof("failed to call %s for %s %d times", targetURL, steamIDs, i)
			res.Body.Close()
		} else {
			successfulRequest = true
			defer res.Body.Close()
			break
		}
	}
	// Failed after all retries
	if !successfulRequest {
		return make(map[string]string), commonUtil.MakeErr(callErr)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return make(map[string]string), commonUtil.MakeErr(err)
	}
	APIRes := dtos.GetUsernamesFromSteamIDsDTO{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		return make(map[string]string), commonUtil.MakeErr(err)
	}

	if res.StatusCode == 200 {
		steamIDToUserMap := make(map[string]string)
		for _, user := range APIRes.SteamIDAndUsername {
			steamIDToUserMap[user.SteamID] = user.Username
		}
		return steamIDToUserMap, nil
	}

	return make(map[string]string), commonUtil.MakeErr(fmt.Errorf("error getting usernames for steamIDs: %+v", APIRes))
}

func (control Cntr) SaveProcessedGraphDataToDataStore(crawlID string, graphData common.UsersGraphData) (bool, error) {
	targetURL := fmt.Sprintf("%s/api/saveprocessedgraphdata/%s", os.Getenv("DATASTORE_URL"), crawlID)

	jsonObj, err := json.Marshal(graphData)
	if err != nil {
		return false, commonUtil.MakeErr(err)
	}

	gzippedData := bytes.Buffer{}
	gz := gzip.NewWriter(&gzippedData)
	if _, err = gz.Write(jsonObj); err != nil {
		return false, err
	}
	if err = gz.Close(); err != nil {
		return false, err
	}

	req, err := http.NewRequest("POST", targetURL, &gzippedData)
	if err != nil {
		return false, err
	}
	req.Close = true
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/javascript")

	client := &http.Client{}
	res := &http.Response{}
	var callErr error

	maxRetryCount := 3
	successfulRequest := false

	for i := 1; i <= maxRetryCount; i++ {
		res, callErr = client.Do(req)
		if callErr != nil {
			configuration.Logger.Sugar().Infof("failed to call %s for %s %d times", targetURL, crawlID, i)
			res.Body.Close()
		} else {
			successfulRequest = true
			defer res.Body.Close()
			break
		}
	}
	// Failed after all retries
	if !successfulRequest {
		return false, commonUtil.MakeErr(callErr)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, commonUtil.MakeErr(err)
	}
	APIRes := dtos.GetUsernamesFromSteamIDsDTO{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		return false, commonUtil.MakeErr(err)
	}

	if res.StatusCode == 200 {
		return true, nil
	}

	return false, commonUtil.MakeErr(fmt.Errorf("error saving processed graphdata: %+v", APIRes))
}

func (control Cntr) GetGameDetailsFromIDs(gameIDs []int) ([]common.BareGameInfo, error) {
	targetURL := fmt.Sprintf("%s/api/getdetailsforgames", os.Getenv("DATASTORE_URL"))

	detailsForGamesInput := dtos.GetDetailsForGamesInputDTO{
		GameIDs: gameIDs,
	}
	jsonObj, err := json.Marshal(detailsForGamesInput)
	if err != nil {
		return []common.BareGameInfo{}, commonUtil.MakeErr(err)
	}

	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonObj))
	if err != nil {
		return []common.BareGameInfo{}, err
	}
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
			configuration.Logger.Sugar().Infof("failed to call %s for %+v %d times", targetURL, gameIDs, i)
			res.Body.Close()
		} else {
			successfulRequest = true
			defer res.Body.Close()
			break
		}
	}
	// Failed after all retries
	if !successfulRequest {
		return []common.BareGameInfo{}, commonUtil.MakeErr(callErr)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []common.BareGameInfo{}, commonUtil.MakeErr(err)
	}
	APIRes := dtos.GetDetailsForGamesDTO{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		return []common.BareGameInfo{}, commonUtil.MakeErr(err)
	}

	if res.StatusCode == 200 {
		return APIRes.Games, nil
	}

	return []common.BareGameInfo{}, commonUtil.MakeErr(fmt.Errorf("error when retrieving details for games: %+v", APIRes))
}

func (control Cntr) Sleep(duration time.Duration) {
	time.Sleep(duration)
}
