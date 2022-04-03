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
	"github.com/neosteamfriendgraphing/common/util"
	commonUtil "github.com/neosteamfriendgraphing/common/util"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
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
	maxRetryCount := 3
	successfulRequest := false

	targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetFriendList/v0001/?key=%s&steamid=%s",
		apiKey, steamID)
	res, err := MakeNetworkGETRequest(targetURL)
	isError := IsErrorResponse(string(res))
	invalidKeyResponse := IsInvalidKeyResponse(string(res))

	if err != nil || isError || invalidKeyResponse {
		logMsg := fmt.Sprintf("error from first call to GetFriendsList (%s), retrying now", targetURL)
		configuration.Logger.Info(logMsg,
			zap.String("errorMsg", fmt.Sprint(err)),
			zap.String("response", string(res)),
			zap.Bool("isError", isError),
			zap.Bool("invalidKeyResponse", invalidKeyResponse))

		for i := 0; i < maxRetryCount; i++ {
			noErrors := true
			// A fresh key must be used
			apiKey = apikeymanager.GetSteamAPIKey()
			targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetFriendList/v0001/?key=%s&steamid=%s",
				apiKey, steamID)

			res, err = MakeNetworkGETRequest(targetURL)
			if invalidKeyResponse := IsInvalidKeyResponse(string(res)); invalidKeyResponse {
				noErrors = false
				configuration.Logger.Sugar().Warnf("invalid key %s request to %s on retry %d caused response: %+v", apiKey, targetURL, i, string(res))
			}

			if isError := IsErrorResponse(string(res)); isError {
				noErrors = false
				configuration.Logger.Sugar().Warnf("got internal server error response: %s on retry %d  when requesting: %s", string(res), i, targetURL)
				time.Sleep(3000 * time.Millisecond)
			}

			if err == nil && noErrors {
				configuration.Logger.Sugar().Infof("success on the %d request to GetFriendsList (%s)", i, targetURL)
				successfulRequest = true
				break
			}

			// No sleep is needed since KEY_USAGE_TIMER limits the distribution of steam keys
			logMsg := fmt.Sprintf("failed to call get friends (%s) %d times", targetURL, i)
			configuration.Logger.Info(logMsg,
				zap.String("errorMsg", fmt.Sprint(err)),
				zap.String("response", string(res)),
				zap.Bool("isError", isError),
				zap.Bool("invalidKeyResponse", invalidKeyResponse))
		}
	} else {
		successfulRequest = true
	}

	if !successfulRequest {
		newErr := fmt.Errorf("failed %d retries to GetFriendList: %+v Most recent response: %+v", maxRetryCount, err, string(res))
		return []string{}, commonUtil.MakeErr(newErr)
	}

	if isError := IsErrorResponse(string(res)); isError {
		return []string{}, commonUtil.MakeErr(fmt.Errorf("invalid key %s caused response: %+v", apiKey, string(res)))
	}

	if invalidKeyResponse := IsInvalidKeyResponse(string(res)); invalidKeyResponse {
		return []string{}, commonUtil.MakeErr(fmt.Errorf("invalid key %s caused response: %+v", apiKey, string(res)))
	}

	err = json.Unmarshal(res, &friendsListObj)
	if err != nil {
		return []string{}, util.MakeErr(err, fmt.Sprintf("error unmarshaling friendsListObj object: %+v, got: %+v", friendsListObj, string(res)))
	}

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
	isError := IsErrorResponse(string(res))
	invalidKeyResponse := IsInvalidKeyResponse(string(res))

	if err != nil || isError || invalidKeyResponse {
		logMsg := fmt.Sprintf("error from first call to getPlayerSummaries (%s), retrying now", targetURL)
		configuration.Logger.Info(logMsg,
			zap.String("errorMsg", fmt.Sprint(err)),
			zap.String("response", string(res)),
			zap.Bool("isError", isError),
			zap.Bool("invalidKeyResponse", invalidKeyResponse))

		for i := 0; i < maxRetryCount; i++ {
			noErrors := true
			// A fresh key must be used
			apiKey = apikeymanager.GetSteamAPIKey()
			targetURL := fmt.Sprintf("http://api.steampowered.com/ISteamUser/GetPlayerSummaries/v0002/?key=%s&steamids=%s",
				apiKey, steamIDStringList)

			res, err = MakeNetworkGETRequest(targetURL)
			if invalidKeyResponse := IsInvalidKeyResponse(string(res)); invalidKeyResponse {
				noErrors = false
				configuration.Logger.Sugar().Warnf("invalid key %s request to %s on retry %d caused response: %+v", apiKey, targetURL, i, string(res))
			}

			if isError := IsErrorResponse(string(res)); isError {
				noErrors = false
				configuration.Logger.Sugar().Warnf("got internal server error response: %s on retry %d when requesting: %s", string(res), i, targetURL)
				time.Sleep(3000 * time.Millisecond)
			}

			if err == nil && noErrors {
				configuration.Logger.Sugar().Infof("success on the %d request to GetPlayerSummaries (%s)", i, targetURL)
				successfulRequest = true
				break
			}
			// No sleep is needed since KEY_USAGE_TIMER limits the distribution of steam keys
			logMsg := fmt.Sprintf("failed to call GetPlayerSummaries (%s) %d times", targetURL, i)
			configuration.Logger.Info(logMsg,
				zap.String("errorMsg", fmt.Sprint(err)),
				zap.String("response", string(res)),
				zap.Bool("isError", isError),
				zap.Bool("invalidKeyResponse", invalidKeyResponse))
		}
	} else {
		successfulRequest = true
	}

	if !successfulRequest {
		newErr := fmt.Errorf("failed %d retries to GetPlayerSummaries: %+v Most recent response: %+v", maxRetryCount, err, string(res))
		return []common.Player{}, commonUtil.MakeErr(newErr)
	}

	if isError := IsErrorResponse(string(res)); isError {
		return []common.Player{}, commonUtil.MakeErr(fmt.Errorf("got internal server error response: %s when requesting: %s", string(res), targetURL))
	}

	if invalidKeyResponse := IsInvalidKeyResponse(string(res)); invalidKeyResponse {
		return []common.Player{}, commonUtil.MakeErr(fmt.Errorf("invalid key %s caused response: %+v", apiKey, string(res)))
	}

	err = json.Unmarshal(res, &allPlayerSummaries)
	if err != nil {
		return []common.Player{}, util.MakeErr(err, fmt.Sprintf("error unmarshaling allPlayerSummaries object: %+v, got: %+v", allPlayerSummaries, string(res)))
	}

	return allPlayerSummaries.Response.Players, nil
}

// CallGetOwnedGames calls the steam web api to retrieve all of a user's owned games
//		ownedGamesResponse, err := CallGetOwnedGames(steamID)
func (control Cntr) CallGetOwnedGames(steamID string) (common.GamesOwnedResponse, error) {
	apiResponse := common.GamesOwnedSteamResponse{}
	apiKey := apikeymanager.GetSteamAPIKey()
	maxRetryCount := 3
	successfulRequest := false

	targetURL := fmt.Sprintf("http://api.steampowered.com/IPlayerService/GetOwnedGames/v0001/?key=%s&steamid=%s&format=json&include_appinfo=true&include_played_free_games=true",
		apiKey, steamID)
	res, err := MakeNetworkGETRequest(targetURL)
	isError := IsErrorResponse(string(res))
	invalidKeyResponse := IsInvalidKeyResponse(string(res))

	if err != nil || isError || invalidKeyResponse {
		logMsg := fmt.Sprintf("error from first call to GetOwnedGames (%s), retrying now", targetURL)
		configuration.Logger.Info(logMsg,
			zap.String("errorMsg", fmt.Sprint(err)),
			zap.String("response", string(res)),
			zap.Bool("isError", isError),
			zap.Bool("invalidKeyResponse", invalidKeyResponse))

		for i := 0; i < maxRetryCount; i++ {
			noErrors := true
			// A fresh key must be used
			apiKey = apikeymanager.GetSteamAPIKey()
			targetURL := fmt.Sprintf("http://api.steampowered.com/IPlayerService/GetOwnedGames/v0001/?key=%s&steamid=%s&format=json&include_appinfo=true&include_played_free_games=true",
				apiKey, steamID)

			res, err = MakeNetworkGETRequest(targetURL)
			if invalidKeyResponse := IsInvalidKeyResponse(string(res)); invalidKeyResponse {
				noErrors = false
				configuration.Logger.Sugar().Warnf("invalid key %s request to %s on retry %d caused response: %+v", apiKey, targetURL, i, string(res))
			}

			if isError := IsErrorResponse(string(res)); isError {
				noErrors = false
				configuration.Logger.Sugar().Warnf("got internal server error response: %s  on retry %d when requesting: %s", string(res), i, targetURL)
				time.Sleep(3000 * time.Millisecond)
			}

			if err == nil && noErrors {
				configuration.Logger.Sugar().Infof("success on the %d request to GetFriendsList  (%s)", i, targetURL)
				successfulRequest = true
				break
			}
			// No sleep is needed since KEY_USAGE_TIMER limits the distribution of steam keys
			logMsg := fmt.Sprintf("failed to call get friends (%s) %d times", targetURL, i)
			configuration.Logger.Info(logMsg,
				zap.String("errorMsg", fmt.Sprint(err)),
				zap.String("response", string(res)),
				zap.Bool("isError", isError),
				zap.Bool("invalidKeyResponse", invalidKeyResponse))
		}
	} else {
		successfulRequest = true
	}

	if !successfulRequest {
		newErr := fmt.Errorf("failed %d retries to GetOwnedGames: %+v Most recent response: %+v", maxRetryCount, err, res)
		return common.GamesOwnedResponse{}, commonUtil.MakeErr(newErr)
	}

	if isError := IsErrorResponse(string(res)); isError {
		return common.GamesOwnedResponse{}, commonUtil.MakeErr(fmt.Errorf("got internal server error response: %s when requesting: %s", string(res), targetURL))
	}

	if invalidKeyResponse := IsInvalidKeyResponse(string(res)); invalidKeyResponse {
		return common.GamesOwnedResponse{}, commonUtil.MakeErr(fmt.Errorf("invalid key %s caused response: %+v", apiKey, string(res)))
	}

	err = json.Unmarshal(res, &apiResponse)
	if err != nil {
		return common.GamesOwnedResponse{}, util.MakeErr(err, fmt.Sprintf("error unmarshalling gamesOwnedResponed object: %+v. got response: %+v", apiResponse.Response, res))
	}

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
	targetURL := fmt.Sprintf("http://%s/api/saveuser", os.Getenv("DATASTORE_INSTANCE"))
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
	req.Header.Set("Authentication", os.Getenv("AUTH_KEY"))

	client := &http.Client{}
	maxRetryCount := 3
	successfulRequest := false

	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		logMsg := fmt.Sprintf("error from first call to saveuser (%s), retrying now", targetURL)
		configuration.Logger.Info(logMsg,
			zap.String("errorMsg", fmt.Sprint(err)),
			zap.String("request", fmt.Sprintf("%+v", res)),
			zap.String("response", fmt.Sprintf("%+v", req)))

		for i := 0; i < maxRetryCount; i++ {
			req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonObj))
			if err != nil {
				return false, err
			}
			req.Close = true
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authentication", os.Getenv("AUTH_KEY"))

			res, err = client.Do(req)
			if err == nil && res.StatusCode == http.StatusOK {
				successfulRequest = true
				defer res.Body.Close()
				break
			}

			exponentialBackOffSleepTime := math.Pow(2, float64(i)) * 16
			configuration.Logger.Sugar().Infof("failed to call %s for %s %d times. Sleeping for %d ms", targetURL, saveUser.User.AccDetails.SteamID, i, exponentialBackOffSleepTime)
			time.Sleep(time.Duration(exponentialBackOffSleepTime) * time.Millisecond)
		}
	} else {
		successfulRequest = true
	}
	// Failed after all retries
	if !successfulRequest {
		failedAllRetriesErr := fmt.Errorf("failed all retries to %s for steamID: %s", targetURL, saveUser.User.AccDetails.SteamID)
		return false, commonUtil.MakeErr(failedAllRetriesErr)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, commonUtil.MakeErr(err, "failed to readAll for saveUser body")
	}
	APIRes := dtos.GetUserDTO{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		return false, commonUtil.MakeErr(err, fmt.Sprintf("failed to unmarshal SaveUser object: %+v", string(body)))
	}

	return true, nil
}

// GetUserFromDataStore gets a user from the datastore service
// 		userFromDataStore, err := GetUserFromDataStore(steamID)
func (control Cntr) GetUserFromDataStore(steamID string) (common.UserDocument, error) {
	targetURL := fmt.Sprintf("http://%s/api/getuser/%s", os.Getenv("DATASTORE_INSTANCE"), steamID)
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return common.UserDocument{}, err
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authentication", os.Getenv("AUTH_KEY"))

	client := &http.Client{}
	maxRetryCount := 3
	successfulRequest := false

	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		logMsg := fmt.Sprintf("error from first call to getuser (%s), retrying now", targetURL)
		configuration.Logger.Info(logMsg,
			zap.String("errorMsg", fmt.Sprint(err)),
			zap.String("request", fmt.Sprintf("%+v", res)),
			zap.String("response", fmt.Sprintf("%+v", req)))

		for i := 0; i < maxRetryCount; i++ {
			res, err = client.Do(req)
			if err == nil && res.StatusCode == http.StatusOK {
				successfulRequest = true
				defer res.Body.Close()
				break
			}

			exponentialBackOffSleepTime := math.Pow(2, float64(i)) * 16
			configuration.Logger.Sugar().Infof("failed to call %s for %s %d times. Sleeping for %d ms", targetURL, steamID, i, exponentialBackOffSleepTime)
			time.Sleep(time.Duration(exponentialBackOffSleepTime) * time.Millisecond)
		}
	} else {
		successfulRequest = true
	}
	// Failed after all retries
	if !successfulRequest {
		failedAllRetriesErr := fmt.Errorf("failed all retries to %s for steamID: %s", targetURL, steamID)
		return common.UserDocument{}, commonUtil.MakeErr(failedAllRetriesErr)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return common.UserDocument{}, commonUtil.MakeErr(err, "failed to readall for getUser body")
	}

	userDoc := dtos.GetUserDTO{}
	err = json.Unmarshal(body, &userDoc)
	if err != nil {
		return common.UserDocument{}, commonUtil.MakeErr(err, fmt.Sprintf("failed to unmarshal getUser object: %+v", string(body)))
	}

	return userDoc.User, nil
}

func (control Cntr) SaveCrawlingStatsToDataStore(currentLevel int, crawlingStatus common.CrawlingStatus) (bool, error) {
	targetURL := fmt.Sprintf("http://%s/api/savecrawlingstats", os.Getenv("DATASTORE_INSTANCE"))
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
	req.Header.Set("Authentication", os.Getenv("AUTH_KEY"))

	client := &http.Client{}
	maxRetryCount := 3
	successfulRequest := false

	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		logMsg := fmt.Sprintf("error from first call to savecrawlingstats (%s), retrying now", targetURL)
		configuration.Logger.Info(logMsg,
			zap.String("errorMsg", fmt.Sprint(err)),
			zap.String("request", fmt.Sprintf("%+v", res)),
			zap.String("response", fmt.Sprintf("%+v", req)))

		for i := 0; i < maxRetryCount; i++ {
			req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonObj))
			if err != nil {
				return false, err
			}
			req.Close = true
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authentication", os.Getenv("AUTH_KEY"))

			res, err = client.Do(req)
			if err == nil && res.StatusCode == http.StatusOK {
				successfulRequest = true
				defer res.Body.Close()
				break
			}

			exponentialBackOffSleepTime := math.Pow(2, float64(i)) * 16
			configuration.Logger.Sugar().Infof("failed to call %s for %+v %d times. Sleeping for %d ms", targetURL, crawlingStatus, i, exponentialBackOffSleepTime)
			time.Sleep(time.Duration(exponentialBackOffSleepTime) * time.Millisecond)
		}
	} else {
		successfulRequest = true
	}
	// Failed after all retries
	if !successfulRequest {
		failedAllRetriesErr := fmt.Errorf("failed all retries to %s for crawlingStatus: %+v", targetURL, crawlingStatus)
		return false, commonUtil.MakeErr(failedAllRetriesErr)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, commonUtil.MakeErr(err, "failed to readAll for savecrawlingstats body")
	}
	APIRes := common.BasicAPIResponse{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		return false, commonUtil.MakeErr(err, fmt.Sprintf("failed to unmarshal savecrawlingstats object: %+v", string(body)))
	}

	return true, nil
}

func (control Cntr) GetCrawlingStatsFromDataStore(crawlID string) (common.CrawlingStatus, error) {
	targetURL := fmt.Sprintf("http://%s/api/getcrawlingstatus/%s", os.Getenv("DATASTORE_INSTANCE"), crawlID)
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return common.CrawlingStatus{}, err
	}
	req.Close = true
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authentication", os.Getenv("AUTH_KEY"))

	client := &http.Client{}
	maxRetryCount := 3
	successfulRequest := false

	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		logMsg := fmt.Sprintf("error from first call to getcrawlingstatus (%s), retrying now", targetURL)
		configuration.Logger.Info(logMsg,
			zap.String("errorMsg", fmt.Sprint(err)),
			zap.String("request", fmt.Sprintf("%+v", res)),
			zap.String("response", fmt.Sprintf("%+v", req)))

		for i := 0; i < maxRetryCount; i++ {
			res, err = client.Do(req)
			if err == nil && res.StatusCode == http.StatusOK {
				successfulRequest = true
				defer res.Body.Close()
				break
			}

			exponentialBackOffSleepTime := math.Pow(2, float64(i)) * 16
			configuration.Logger.Sugar().Infof("failed to call %s for %s %d times. Sleeping for %d ms", targetURL, crawlID, i, exponentialBackOffSleepTime)
			time.Sleep(time.Duration(exponentialBackOffSleepTime) * time.Millisecond)
		}
	} else {
		successfulRequest = true
	}
	// Failed after all retries
	if !successfulRequest {
		failedAllRetriesErr := fmt.Errorf("failed all retries to %s for crawlID: %s", targetURL, crawlID)
		return common.CrawlingStatus{}, commonUtil.MakeErr(failedAllRetriesErr)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return common.CrawlingStatus{}, commonUtil.MakeErr(err, "failed to readAll for getcrawlingstatus body")
	}
	APIRes := dtos.GetCrawlingStatusDTO{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		return common.CrawlingStatus{}, commonUtil.MakeErr(err, fmt.Sprintf("failed to unmarshal getcrawlingstatus object: %+v", string(body)))
	}

	return APIRes.CrawlingStatus, nil
}

func (control Cntr) GetGraphableDataFromDataStore(steamID string) (dtos.GetGraphableDataForUserDTO, error) {
	targetURL := fmt.Sprintf("http://%s/api/getgraphabledata/%s", os.Getenv("DATASTORE_INSTANCE"), steamID)
	req, err := http.NewRequest("GET", targetURL, nil)
	if err != nil {
		return dtos.GetGraphableDataForUserDTO{}, err
	}
	req.Close = true
	req.Header.Set("Authentication", os.Getenv("AUTH_KEY"))

	client := &http.Client{}
	maxRetryCount := 3
	successfulRequest := false

	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		logMsg := fmt.Sprintf("error from first call to getgraphabledata (%s), retrying now", targetURL)
		configuration.Logger.Info(logMsg,
			zap.String("errorMsg", fmt.Sprint(err)),
			zap.String("request", fmt.Sprintf("%+v", res)),
			zap.String("response", fmt.Sprintf("%+v", req)))

		for i := 0; i < maxRetryCount; i++ {
			res, err = client.Do(req)
			if err == nil && res.StatusCode == http.StatusOK {
				successfulRequest = true
				defer res.Body.Close()
				break
			}

			exponentialBackOffSleepTime := math.Pow(2, float64(i)) * 16
			configuration.Logger.Sugar().Infof("failed to call %s for %s %d times. Sleeping for %d ms", targetURL, steamID, i, exponentialBackOffSleepTime)
			time.Sleep(time.Duration(exponentialBackOffSleepTime) * time.Millisecond)
		}
	} else {
		successfulRequest = true
	}
	// Failed after all retries
	if !successfulRequest {
		failedAllRetriesErr := fmt.Errorf("failed all retries to %s for steamID: %s", targetURL, steamID)
		return dtos.GetGraphableDataForUserDTO{}, commonUtil.MakeErr(failedAllRetriesErr)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return dtos.GetGraphableDataForUserDTO{}, commonUtil.MakeErr(err, "failed to readAll for getgraphabledata body")
	}

	APIRes := dtos.GetGraphableDataForUserDTO{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		return dtos.GetGraphableDataForUserDTO{}, commonUtil.MakeErr(err, "failed to unmarshal getgraphabledata")
	}

	if res.StatusCode == 200 {
		return APIRes, nil
	}
	return dtos.GetGraphableDataForUserDTO{}, commonUtil.MakeErr(fmt.Errorf("error getting crawling status: %+v", APIRes))
}

func (control Cntr) GetUsernamesForSteamIDs(steamIDs []string) (map[string]string, error) {
	targetURL := fmt.Sprintf("http://%s/api/getusernamesfromsteamids", os.Getenv("DATASTORE_INSTANCE"))
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
	req.Header.Set("Authentication", os.Getenv("AUTH_KEY"))

	client := &http.Client{}
	maxRetryCount := 3
	successfulRequest := false
	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		logMsg := fmt.Sprintf("error from first call to getusernamesfromsteamids (%s), retrying now", targetURL)
		configuration.Logger.Info(logMsg,
			zap.String("errorMsg", fmt.Sprint(err)),
			zap.String("request", fmt.Sprintf("%+v", res)),
			zap.String("response", fmt.Sprintf("%+v", req)))

		for i := 0; i < maxRetryCount; i++ {
			req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonObj))
			if err != nil {
				return make(map[string]string), err
			}
			req.Close = true
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authentication", os.Getenv("AUTH_KEY"))

			res, err = client.Do(req)
			if err == nil && res.StatusCode == http.StatusOK {
				successfulRequest = true
				defer res.Body.Close()
				break
			}

			exponentialBackOffSleepTime := math.Pow(2, float64(i)) * 16
			configuration.Logger.Sugar().Infof("failed to call %s for %+v %d times. Sleeping for %d ms", targetURL, steamIDs, i, exponentialBackOffSleepTime)
			time.Sleep(time.Duration(exponentialBackOffSleepTime) * time.Millisecond)
		}
	} else {
		successfulRequest = true
	}
	// Failed after all retries
	if !successfulRequest {
		failedAllRetriesErr := fmt.Errorf("failed all retries to %s for steamIDs: %+v", targetURL, steamIDs)
		return make(map[string]string), commonUtil.MakeErr(failedAllRetriesErr)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return make(map[string]string), commonUtil.MakeErr(err, "failed to readAll for getusernamesfromsteamids body")
	}

	APIRes := dtos.GetUsernamesFromSteamIDsDTO{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		return make(map[string]string), commonUtil.MakeErr(err, fmt.Sprintf("failed to unmarshal getusernamesfromsteamids object: %+v", string(body)))
	}

	steamIDToUserMap := make(map[string]string)
	for _, user := range APIRes.SteamIDAndUsername {
		steamIDToUserMap[user.SteamID] = user.Username
	}
	return steamIDToUserMap, nil
}

func (control Cntr) SaveProcessedGraphDataToDataStore(crawlID string, graphData common.UsersGraphData) (bool, error) {
	targetURL := fmt.Sprintf("http://%s/api/saveprocessedgraphdata/%s", os.Getenv("DATASTORE_INSTANCE"), crawlID)

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
	req.Header.Set("Authentication", os.Getenv("AUTH_KEY"))

	client := &http.Client{}
	maxRetryCount := 3
	successfulRequest := false

	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		logMsg := fmt.Sprintf("error from first call to saveprocessedgraphdata (%s), retrying now", targetURL)
		configuration.Logger.Info(logMsg,
			zap.String("errorMsg", fmt.Sprint(err)),
			zap.String("request", fmt.Sprintf("%+v", res)),
			zap.String("response", fmt.Sprintf("%+v", req)))

		for i := 0; i < maxRetryCount; i++ {
			req, err := http.NewRequest("POST", targetURL, &gzippedData)
			if err != nil {
				return false, err
			}
			req.Close = true
			req.Header.Set("Content-Encoding", "gzip")
			req.Header.Set("Content-Type", "application/javascript")
			req.Header.Set("Authentication", os.Getenv("AUTH_KEY"))

			res, err = client.Do(req)
			if err == nil {
				defer res.Body.Close()
			}
			if err == nil && res.StatusCode == http.StatusOK {
				successfulRequest = true
				break
			}

			exponentialBackOffSleepTime := math.Pow(2, float64(i)) * 16
			logMsg := fmt.Sprintf("failed to call %s for %+v %d times. Sleeping for %f ms", targetURL, crawlID, i, exponentialBackOffSleepTime)
			configuration.Logger.Info(logMsg,
				zap.String("errorMsg", fmt.Sprint(err)),
				zap.String("request", fmt.Sprintf("%+v", res)),
				zap.String("response", fmt.Sprintf("%+v", req)))
			time.Sleep(time.Duration(exponentialBackOffSleepTime) * time.Millisecond)
		}
	} else {
		successfulRequest = true
	}
	// Failed after all retries
	if !successfulRequest {
		failedAllRetriesErr := fmt.Errorf("failed all retries to %s for crawlID: %s", targetURL, crawlID)
		return false, commonUtil.MakeErr(failedAllRetriesErr)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, commonUtil.MakeErr(err, "failed to readAll for saveprocessedgraphdata body")
	}

	APIRes := dtos.GetUsernamesFromSteamIDsDTO{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		return false, commonUtil.MakeErr(err, "failed to unmarshal saveprocessedgraphdata object")
	}

	return true, nil
}

func (control Cntr) GetGameDetailsFromIDs(gameIDs []int) ([]common.BareGameInfo, error) {
	targetURL := fmt.Sprintf("http://%s/api/getdetailsforgames", os.Getenv("DATASTORE_INSTANCE"))

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
	req.Header.Set("Authentication", os.Getenv("AUTH_KEY"))

	client := &http.Client{}
	maxRetryCount := 3
	successfulRequest := false

	res, err := client.Do(req)
	if err != nil || res.StatusCode != http.StatusOK {
		logMsg := fmt.Sprintf("error from first call to getdetailsforgames (%s), retrying now", targetURL)
		configuration.Logger.Info(logMsg,
			zap.String("errorMsg", fmt.Sprint(err)),
			zap.String("request", fmt.Sprintf("%+v", res)),
			zap.String("response", fmt.Sprintf("%+v", req)))

		for i := 0; i < maxRetryCount; i++ {
			req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(jsonObj))
			if err != nil {
				return []common.BareGameInfo{}, err
			}
			req.Close = true
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authentication", os.Getenv("AUTH_KEY"))

			res, err = client.Do(req)
			if err == nil && res.StatusCode == http.StatusOK {
				defer res.Body.Close()
				successfulRequest = true
				break
			}

			exponentialBackOffSleepTime := math.Pow(2, float64(i)) * 16
			configuration.Logger.Sugar().Infof("failed to call %s for %+v %d times. Sleeping for %d ms", targetURL, gameIDs, i, exponentialBackOffSleepTime)
			time.Sleep(time.Duration(exponentialBackOffSleepTime) * time.Millisecond)
		}
	} else {
		successfulRequest = true
	}
	// Failed after all retries
	if !successfulRequest {
		failedAllRetriesErr := fmt.Errorf("failed all retries to %s for gameIDs: %+v", targetURL, gameIDs)
		return []common.BareGameInfo{}, commonUtil.MakeErr(failedAllRetriesErr)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return []common.BareGameInfo{}, commonUtil.MakeErr(err, "failed to readAll for getdetailsforgames body")
	}

	APIRes := dtos.GetDetailsForGamesDTO{}
	err = json.Unmarshal(body, &APIRes)
	if err != nil {
		return []common.BareGameInfo{}, commonUtil.MakeErr(err, fmt.Sprintf("failed to unmarshal getdetailsforgames object: %+v", string(body)))
	}

	return APIRes.Games, nil
}

func (control Cntr) Sleep(duration time.Duration) {
	time.Sleep(duration)
}

func IsErrorResponse(response string) bool {
	return response == "<html><head><title>Internal Server Error</title></head><body><h1>Internal Server Error</h1>Failed to forward request message to internal server</body></html>"
}

func IsInvalidKeyResponse(response string) bool {
	return response == "<html><head><title>Forbidden</title></head><body><h1>Forbidden</h1>Access is denied. Retrying will not help. Please verify your <pre>key=</pre> parameter.</body></html>"
}
