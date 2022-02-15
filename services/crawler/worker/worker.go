package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/iamcathal/neo/services/crawler/amqpchannelmanager"
	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/iamcathal/neo/services/crawler/util"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/dtos"
	commonUtil "github.com/neosteamfriendgraphing/common/util"
)

// Worker crawls the steam API to get data from steam for a given user
// e.g account details and details of a user's friend
func Worker(cntr controller.CntrInterface, job datastructures.Job) {
	userWasFoundInDB, friendsList, err := GetFriends(cntr, job.CurrentTargetSteamID)
	if err != nil {
		log.Fatal(err)
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
			log.Fatal(err)
		}
		if !success {
			configuration.Logger.Sugar().Fatalf("failed to save crawling stats to DB for existing user: %+v", err)
			log.Fatal(err)
		}

		friendsShoudlBeCrawled := util.JobIsNotLevelOneAndNotMax(job)
		// If the job is not at max level or has a max level of one, add
		// friends to the queue for crawling
		if friendsShoudlBeCrawled {
			err = putFriendsIntoQueue(cntr, job, friendsList)
			if err != nil {
				configuration.Logger.Fatal(fmt.Sprintf("failed publish friends from steamID: %s to queue: %v", job.CurrentTargetSteamID, err.Error()))
				log.Fatal(err)
			}
		}
		return
	}
	// User was never crawled before
	playerSummaries, err := cntr.CallGetPlayerSummaries(job.CurrentTargetSteamID)
	if err != nil {
		configuration.Logger.Fatal(fmt.Sprintf("failed to get player summary for target user: %v", err.Error()))
		log.Fatal(err)
	}

	// Sometimes occurs with accounts that have complex combinatioons of data privacy settings
	if len(playerSummaries) == 0 {
		playerSummaries, err = cntr.CallGetPlayerSummaries(job.CurrentTargetSteamID)
		if err != nil {
			configuration.Logger.Sugar().Panicf(fmt.Sprintf("failed AGAIN to get player summary for target user: %+v", err))
		}
		if len(playerSummaries) == 0 {
			configuration.Logger.Sugar().Panicf("failed to get a non empty player summary for target user for a second time: %+v", err)
		}
	}
	playerSummaryForCurrentUser := playerSummaries[0]

	allGamesOwnedForCurrentUser, err := getGamesOwned(cntr, job.CurrentTargetSteamID)
	if err != nil {
		configuration.Logger.Fatal(fmt.Sprintf("failed to get player summaries for friends: %v", err.Error()))
		log.Fatal(err)
	}
	topFiftyOrFewerTopPlayedGames := getTopFiftyOrFewerGames(allGamesOwnedForCurrentUser)
	topFiftyOrFewerGamesOwnedSlimmedDown := GetSlimmedDownOwnedGames(topFiftyOrFewerTopPlayedGames)

	friendPlayerSummaries := []common.Player{}

	if len(friendsList) != 0 {
		friendPlayerSummaries, err = getPlayerSummaries(cntr, job, friendsList)
		if err != nil {
			configuration.Logger.Fatal(fmt.Sprintf("failed to get player summaries for friends: %v", err.Error()))
			log.Fatal(err)
		}
	}

	friendPlayerSummarySteamIDs := getSteamIDsFromPlayers(friendPlayerSummaries)
	friendsShoudlBeCrawled := util.JobIsNotLevelOneAndNotMax(job)
	if friendsShoudlBeCrawled {
		err = putFriendsIntoQueue(cntr, job, friendPlayerSummarySteamIDs)
		if err != nil {
			configuration.Logger.Fatal(fmt.Sprintf("failed publish friends from steamID: %s to queue: %v", job.CurrentTargetSteamID, err.Error()))
			log.Fatal(err)
		}
	}

	privateFriendCount := len(friendsList) - len(friendPlayerSummaries)
	publicFriendCount := len(friendsList) - privateFriendCount
	logMsg := fmt.Sprintf("Got data for [%s][%s][%s][%d public %d private friends][%d games]",
		playerSummaryForCurrentUser.Steamid, playerSummaryForCurrentUser.Personaname, playerSummaryForCurrentUser.Loccountrycode,
		publicFriendCount, privateFriendCount, len(topFiftyOrFewerGamesOwnedSlimmedDown))
	configuration.Logger.Info(logMsg)

	// Save game details to DB

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
			GamesOwned: topFiftyOrFewerGamesOwnedSlimmedDown,
		},
		// GamesOwnedFull: topTwentyOrFewerGamesSlimmedDown,
	}

	success, err := cntr.SaveUserToDataStore(saveUser)
	if err != nil {
		log.Fatal(err)
	}
	if !success {
		configuration.Logger.Sugar().Errorf("failed to save user %s to DB: %+v", saveUser.User.AccDetails.SteamID, err)
		panic(fmt.Sprintf("failed to save user %s to DB: %+v", saveUser.User.AccDetails.SteamID, err))
	}
}

// GetFriends gets the friendslist for a given user through either datastore
// or the steam web API.
// 		userWasFoundInDB, friendIDs, err := GetFriends(cntr, steamID)
func GetFriends(cntr controller.CntrInterface, steamID string) (bool, []string, error) {
	userWasFoundInDB := false
	userFromDB, err := cntr.GetUserFromDataStore(steamID)
	if err != nil {
		configuration.Logger.Sugar().Infof("error getting user in DB: %+v", err)
	}
	if userFromDB.AccDetails.SteamID == "" {
		configuration.Logger.Sugar().Infof("user %s was not found in DB", steamID)
	} else {
		userWasFoundInDB = true
		// configuration.Logger.Sugar().Infof("returning user retrieved from DB: %+v", userFromDB.AccDetails.SteamID)
		return userWasFoundInDB, userFromDB.FriendIDs, nil
	}

	// User was not found in DB, call the API
	friendsList, err := cntr.CallGetFriends(steamID)
	if err != nil {
		return userWasFoundInDB, []string{}, err
	}
	return userWasFoundInDB, friendsList, nil
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
