package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/iamcathal/neo/services/crawler/amqpchannelmanager"
	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/iamcathal/neo/services/crawler/util"
	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/dtos"
	commonUtil "github.com/neosteamfriendgraphing/common/util"
)

// Worker crawls the steam API to get data from steam for a given user
// e.g account details and details of a user's friend
func Worker(cntr controller.CntrInterface, job datastructures.Job) {
	startTime := time.Now().UnixNano() / int64(time.Millisecond)

	userWasFoundInDB, friendsList, err := GetFriends(cntr, job.CurrentTargetSteamID)
	if err != nil {
		configuration.Logger.Sugar().Panicf("error getting friends initially in worker: %+v", err)
	}
	if userWasFoundInDB {
		crawlingStatus := common.CrawlingStatus{
			OriginalCrawlTarget: job.OriginalTargetSteamID,
			MaxLevel:            job.MaxLevel,
			CrawlID:             job.CrawlID,
			TotalUsersToCrawl:   len(friendsList),
		}

		success, err := cntr.SaveCrawlingStatsToDataStore(job.CurrentLevel, crawlingStatus)
		if err != nil {
			configuration.Logger.Sugar().Panicf("error saving crawling stats in worker: %+v", err)
		}
		if !success {
			configuration.Logger.Sugar().Panicf("failed to save crawling stats in worker: %+v", err)
		}

		friendsShoudlBeCrawled := util.JobIsNotLevelOneAndNotMax(job)
		// If the job is not at max level or has a max level of one, add
		// friends to the queue for crawling
		if friendsShoudlBeCrawled {
			err = putFriendsIntoQueue(cntr, job, friendsList)
			if err != nil {
				configuration.Logger.Sugar().Panicf("error publishing friends from steamID: %s to queue: %+v", job.CurrentTargetSteamID, err)
			}
		}

		point := influxdb2.NewPointWithMeasurement("crawlerMetrics").
			AddTag("service", "crawler").
			AddTag("fromdatastore", "yes").
			AddField("totalTime", commonUtil.GetCurrentTimeInMs()-startTime).
			AddField("totalfriends", len(friendsList)).
			SetTime(time.Now())
		configuration.EndpointWriteAPI.WritePoint(point)
		return
	}
	playerSummaryForCurrentUser := common.Player{}
	fiftyOrFewerGamesOwnedForCurrentUser := []common.GameOwnedDocument{}
	friendPlayerSummaries := []common.Player{}
	var waitG sync.WaitGroup

	durationForGetSummaryForMainUser := int64(0)
	durationForGetFiftyOrFewerGamesOwned := int64(0)
	durationForGetSummariesForFriends := int64(0)
	waitG.Add(1)
	go getSummaryForMainUserFunc(
		cntr,
		job.CurrentTargetSteamID,
		&playerSummaryForCurrentUser,
		&durationForGetSummaryForMainUser,
		&waitG)

	waitG.Add(1)
	go getFiftyOrFewerGamesOwnedFunc(
		cntr,
		job.CurrentTargetSteamID,
		&fiftyOrFewerGamesOwnedForCurrentUser,
		&durationForGetFiftyOrFewerGamesOwned,
		&waitG)

	waitG.Add(1)
	go getSummariesForFriendsFunc(
		cntr,
		friendsList,
		&friendPlayerSummaries,
		&durationForGetSummariesForFriends,
		&waitG)

	waitG.Wait()
	emptyPlayer := common.Player{}
	if playerSummaryForCurrentUser == emptyPlayer {
		configuration.Logger.Sugar().Infof("caught ultra secure private user, ignoring this job: %+v", job)
		return
	}

	// ASYNC BLOCK TWO

	privateFriendCount := len(friendsList) - len(friendPlayerSummaries)
	publicFriendCount := len(friendsList) - privateFriendCount

	logMsg := fmt.Sprintf("Got data for [%s][%s][%s][%d public %d private friends][%d games]",
		playerSummaryForCurrentUser.Steamid, playerSummaryForCurrentUser.Personaname, playerSummaryForCurrentUser.Loccountrycode,
		publicFriendCount, privateFriendCount, len(fiftyOrFewerGamesOwnedForCurrentUser))
	configuration.Logger.Info(logMsg)

	// PUT FRIENDS INTO QUEUE
	friendPlayerSummarySteamIDs := getSteamIDsFromPlayers(friendPlayerSummaries)
	friendsShoudlBeCrawled := util.JobIsNotLevelOneAndNotMax(job)
	publishFriendsToQueueDuration := int64(0)
	if friendsShoudlBeCrawled {
		waitG.Add(1)
		go publishFriendsToQueueFunc(cntr, job, friendPlayerSummarySteamIDs, &publishFriendsToQueueDuration, &waitG)
	}

	// // Save user to DB
	saveUser := dtos.SaveUserDTO{
		OriginalCrawlTarget: job.OriginalTargetSteamID,
		CurrentLevel:        job.CurrentLevel,
		MaxLevel:            job.MaxLevel,
		CrawlID:             job.CrawlID,
		User: common.UserDocument{
			AccDetails: common.AccDetailsDocument{
				SteamID:        playerSummaryForCurrentUser.Steamid,
				Personaname:    playerSummaryForCurrentUser.Personaname,
				Profileurl:     playerSummaryForCurrentUser.Profileurl,
				Avatar:         playerSummaryForCurrentUser.Avatar,
				Timecreated:    playerSummaryForCurrentUser.Timecreated,
				Loccountrycode: playerSummaryForCurrentUser.Loccountrycode,
			},
			FriendIDs:  friendPlayerSummarySteamIDs,
			GamesOwned: fiftyOrFewerGamesOwnedForCurrentUser,
		},
	}

	waitG.Add(1)
	saveUserDuration := int64(0)
	go saveUserFunc(cntr, saveUser, &saveUserDuration, &waitG)

	waitG.Wait()

	writeAPI := configuration.InfluxDBClient.WriteAPI(os.Getenv("ORG"), "crawlerMetrics")
	point := influxdb2.NewPointWithMeasurement("crawlerMetrics").
		AddTag("fromdatastore", "no").
		AddField("totalfriends", publicFriendCount).
		AddField("totalTime", commonUtil.GetCurrentTimeInMs()-startTime).
		AddField("getplayersummaryduration", durationForGetSummaryForMainUser).
		AddField("getgamesownedduration", durationForGetFiftyOrFewerGamesOwned).
		AddField("getfriendsplayersummariesduration", durationForGetSummariesForFriends).
		AddField("publishfriendstoqueueduration", publishFriendsToQueueDuration).
		AddField("saveuserduration", saveUserDuration).
		AddField("gamesowned", len(fiftyOrFewerGamesOwnedForCurrentUser)).
		SetTime(time.Now())
	writeAPI.WritePoint(point)
}

// GetFriends gets the friendslist for a given user through either datastore
// or the steam web API.
// 		userWasFoundInDB, friendIDs, err := GetFriends(cntr, steamID)
func GetFriends(cntr controller.CntrInterface, steamID string) (bool, []string, error) {
	userFromDB, err := cntr.GetUserFromDataStore(steamID)
	if err != nil {
		configuration.Logger.Sugar().Infof("error getting user in DB: %+v", err)
	}
	if userFromDB.AccDetails.SteamID != "" {
		configuration.Logger.Sugar().Infof("returning user retrieved from DB: %+v", userFromDB.AccDetails.SteamID)
		return true, userFromDB.FriendIDs, nil
	}

	configuration.Logger.Sugar().Infof("user %s was not found in DB", steamID)
	// User was not found in DB, call the API
	friendsList, err := cntr.CallGetFriends(steamID)
	if err != nil {
		return false, []string{}, err
	}
	return false, friendsList, nil
}

// ControlFunc manages workers
func ControlFunc(cntr controller.CntrInterface) {
	msgs, err := cntr.ConsumeFromJobsQueue()
	if err != nil {
		configuration.Logger.Fatal(fmt.Sprintf("failed to consume from jobs queue on ControlFunc init: %v", err))
		log.Fatal(err)
	}

	for {
		for d := range msgs {
			newJob := datastructures.Job{}
			err := json.Unmarshal(d.Body, &newJob)
			if err != nil {
				configuration.Logger.Sugar().Panicf(fmt.Sprintf("failed unmarshal job from queue: %v", err))
			}

			configuration.Logger.Sugar().Infof("control func received job: %+v", newJob)
			Worker(cntr, newJob)
			d.Ack(false)
		}
	}
}

func CrawlUser(cntr controller.CntrInterface, steamID, crawlID string, level int) error {
	newJob := datastructures.Job{
		JobType:               "crawl",
		OriginalTargetSteamID: steamID,
		CurrentTargetSteamID:  steamID,
		CrawlID:               crawlID,
		MaxLevel:              level,
		CurrentLevel:          1,
	}
	jsonObj, err := json.Marshal(newJob)
	if err != nil {
		return commonUtil.MakeErr(err, fmt.Sprintf("failed to marshal initial crawl job: %+v", newJob))
	}

	crawlingStatus := common.CrawlingStatus{
		TimeStarted:         time.Now().Unix(),
		OriginalCrawlTarget: newJob.OriginalTargetSteamID,
		MaxLevel:            newJob.MaxLevel,
		CrawlID:             newJob.CrawlID,
		UsersCrawled:        0,
		TotalUsersToCrawl:   1,
	}
	success, err := cntr.SaveCrawlingStatsToDataStore(newJob.CurrentLevel, crawlingStatus)
	if err != nil {
		return err
	}
	if !success {
		configuration.Logger.Sugar().Errorf("failed to save crawling stats to DB for existing user: %+v", err)
		return err
	}
	configuration.Logger.Sugar().Infof("created crawling %+v", crawlingStatus)

	err = amqpchannelmanager.PublishToJobsQueue(cntr, jsonObj)
	if err != nil {
		configuration.Logger.Sugar().Errorf("failed to publish new crawl user job with steamID: %s level: %d to queue: %+v",
			steamID, level, err)
	} else {
		configuration.Logger.Info(fmt.Sprintf("placed job steamID: %s level: %d into queue", steamID, level))
	}
	return err
}

func getSummaryForMainUserFunc(cntr controller.CntrInterface, steamID string, mainUser *common.Player, durationForGetPlayerSummary *int64, waitG *sync.WaitGroup) {
	defer waitG.Done()
	startTime := time.Now().UnixNano() / int64(time.Millisecond)
	playerSummaries, err := cntr.CallGetPlayerSummaries(steamID)
	if err != nil {
		configuration.Logger.Fatal(fmt.Sprintf("failed to get player summary for target user %s: %v", steamID, err.Error()))
		log.Fatal(err)
	}

	// Sometimes occurs with accounts that have complex combinatioons of data privacy settings
	if len(playerSummaries) == 0 {
		playerSummaries, err = cntr.CallGetPlayerSummaries(steamID)
		if err != nil {
			configuration.Logger.Sugar().Panicf(fmt.Sprintf("failed AGAIN to get player summary for target user: %+v", err))
		}
		if len(playerSummaries) == 0 {
			// This is a very odd occurance and I do not know how a user can
			// have their privacy settings this strict. It occurs very rarely
			// like for 76561198043146238. Instead this filler user will
			// be returned
			randomFourDigitNumber := rand.Intn(9000) + 999
			playerSummaries = append(playerSummaries, common.Player{
				Steamid:        fmt.Sprintf("7656119804695%d", randomFourDigitNumber),
				Personaname:    "unknown",
				Profileurl:     "https://steamcommunity.com/id/gabelogannewell",
				Avatar:         "https://avatars.cloudflare.steamstatic.com/c5d56249ee5d28a07db4ac9f7f60af961fab5426.jpg",
				Timecreated:    1648665076,
				Loccountrycode: "US",
			})
			configuration.Logger.Sugar().Infof(
				"target user %s has ultra secure privacy settings: %+v", steamID, err)
		}
	}
	*mainUser = playerSummaries[0]
	*durationForGetPlayerSummary = commonUtil.GetCurrentTimeInMs() - startTime
}

func getFiftyOrFewerGamesOwnedFunc(cntr controller.CntrInterface, steamID string, gamesOwned *[]common.GameOwnedDocument, durationForGetFiftyOrFewerGamesOwned *int64, waitG *sync.WaitGroup) {
	defer waitG.Done()
	startTime := time.Now().UnixNano() / int64(time.Millisecond)
	allGamesOwnedForCurrentUser, err := getGamesOwned(cntr, steamID)
	if err != nil {
		configuration.Logger.Sugar().Panicf("failed to get player summaries for friends: %+v", err)
	}
	topFiftyOrFewerTopPlayedGames := getTopFiftyOrFewerGames(allGamesOwnedForCurrentUser)
	topFiftyOrFewerGamesOwnedSlimmedDown := GetSlimmedDownOwnedGames(topFiftyOrFewerTopPlayedGames)

	*gamesOwned = topFiftyOrFewerGamesOwnedSlimmedDown
	*durationForGetFiftyOrFewerGamesOwned = commonUtil.GetCurrentTimeInMs() - startTime
}

func getSummariesForFriendsFunc(cntr controller.CntrInterface, friendIDs []string, friends *[]common.Player, durationForGetSummariesForFriends *int64, waitG *sync.WaitGroup) {
	defer waitG.Done()
	startTime := time.Now().UnixNano() / int64(time.Millisecond)
	if len(friendIDs) == 0 {
		*durationForGetSummariesForFriends = commonUtil.GetCurrentTimeInMs() - startTime
		return
	}

	friendPlayerSummaries, err := getPlayerSummaries(cntr, friendIDs)
	if err != nil {
		configuration.Logger.Sugar().Panicf("failed to get player summaries for friends: %+v", err)
		log.Fatal(err)
	}

	*friends = friendPlayerSummaries
	*durationForGetSummariesForFriends = commonUtil.GetCurrentTimeInMs() - startTime
}

func publishFriendsToQueueFunc(cntr controller.CntrInterface, job datastructures.Job, friendIDs []string, publishFriendsToQueueDuration *int64, waitG *sync.WaitGroup) {
	defer waitG.Done()
	startTime := time.Now().UnixNano() / int64(time.Millisecond)
	err := putFriendsIntoQueue(cntr, job, friendIDs)
	if err != nil {
		configuration.Logger.Sugar().Panicf("failed publish friends from steamID: %s to queue: %+v", job.CurrentTargetSteamID, err)
	}
	*publishFriendsToQueueDuration = commonUtil.GetCurrentTimeInMs() - startTime
}

func saveUserFunc(cntr controller.CntrInterface, saveUser dtos.SaveUserDTO, saveUserDuration *int64, waitG *sync.WaitGroup) {
	defer waitG.Done()
	startTime := time.Now().UnixNano() / int64(time.Millisecond)
	success, err := cntr.SaveUserToDataStore(saveUser)
	if err != nil {
		configuration.Logger.Sugar().Panicf("error when saving user user %s to DB: %+v", saveUser.User.AccDetails.SteamID, err)
	}
	if !success {
		configuration.Logger.Sugar().Panicf("failed to save user %s to DB: %+v", saveUser.User.AccDetails.SteamID, err)
	}
	*saveUserDuration = commonUtil.GetCurrentTimeInMs() - startTime
}
