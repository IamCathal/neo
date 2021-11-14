package worker

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/iamcathal/neo/services/crawler/util"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/dtos"
)

var (
	jobsChannel chan datastructures.Job
)

// Worker crawls the steam API to get data from steam for a given user
// e.g account details and details of a user's friend
func Worker(cntr controller.CntrInterface, job datastructures.Job) {
	userWasFoundInDB, friendsList, err := GetFriends(cntr, job.CurrentTargetSteamID)
	if err != nil {
		log.Fatal(err)
	}
	if userWasFoundInDB {
		friendsNeedCrawling := util.JobIsNotLevelOneAndNotMax(job)
		// If the job is not at max level or has a max level of one, add
		// friends to the queue for crawling
		if friendsNeedCrawling {
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
	playerSummaryForCurrentUser := playerSummaries[0]

	gamesOwnedForCurrentUser, err := getGamesOwned(cntr, job.CurrentTargetSteamID)
	if err != nil {
		configuration.Logger.Fatal(fmt.Sprintf("failed to get player summaries for friends: %v", err.Error()))
		log.Fatal(err)
	}

	friendPlayerSummaries, err := getPlayerSummaries(cntr, job, friendsList)
	if err != nil {
		configuration.Logger.Fatal(fmt.Sprintf("failed to get player summaries for friends: %v", err.Error()))
		log.Fatal(err)
	}

	friendPlayerSummarySteamIDs := getSteamIDsFromPlayers(friendPlayerSummaries)
	err = putFriendsIntoQueue(cntr, job, friendPlayerSummarySteamIDs)
	if err != nil {
		configuration.Logger.Fatal(fmt.Sprintf("failed publish friends from steamID: %s to queue: %v", job.CurrentTargetSteamID, err.Error()))
		log.Fatal(err)
	}

	// TODO Implement when target user profile summary is included in the main call
	// Will also need to slice off this user because a user cannot be in their own friendslist
	// found, targetUsersProfileSummary := getUsersProfileSummaryFromSlice(job.CurrentTargetSteamID, playerSummaries)
	// if !found {
	// 	log.Fatal("players own summary not found in lookup")
	// }
	logMsg := fmt.Sprintf("Got data for [%s][%s][%s][%d friends][%d games]",
		playerSummaryForCurrentUser.Steamid, playerSummaryForCurrentUser.Personaname, playerSummaryForCurrentUser.Loccountrycode,
		len(friendPlayerSummaries), len(gamesOwnedForCurrentUser))
	configuration.Logger.Info(logMsg)

	// Save game details to DB

	// // Save user to DB
	saveUser := dtos.SaveUserDTO{
		OriginalCrawlTarget: job.OriginalTargetSteamID,
		CurrentLevel:        job.CurrentLevel,
		MaxLevel:            job.MaxLevel,
		User: common.UserDocument{
			SteamID:    job.CurrentTargetSteamID,
			AccDetails: playerSummaryForCurrentUser,
			FriendIDs:  friendsList,
			GamesOwned: gamesOwnedForCurrentUser,
		},
	}
	success, err := cntr.SaveFriendsListToDataStore(saveUser)
	if err != nil {
		log.Fatal(err)
	}
	if !success {
		configuration.Logger.Sugar().Fatalf("failed to save user to DB: %+v", err)
		log.Fatal(err)
	}
	configuration.Logger.Info("saved user to DB")
}

// GetFriends gets the friendslist for a given user through either the steam web API
// or cache
func GetFriends(cntr controller.CntrInterface, steamID string) (bool, []string, error) {
	// First call the db
	userWasFoundInDB := false
	userFromDB, err := cntr.GetUserFromDataStore(steamID)
	if err != nil {
		configuration.Logger.Sugar().Infof("error getting user in DB: %+v", err)
	}
	// TODO implement proper account exist check
	if userFromDB.SteamID == "" {
		configuration.Logger.Sugar().Infof("user %s was not found in DB", steamID)
	} else {
		userWasFoundInDB = true
		configuration.Logger.Sugar().Infof("returning user retrieved from DB: %+v", userFromDB.SteamID)
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
				configuration.Logger.Fatal(fmt.Sprintf("failed unmarshal job from queue: %v", err))
				log.Fatal(err)
			}

			logMsg := fmt.Sprintf("Received job: original: %s, current: %s, max level: %d, currlevel: %d", newJob.OriginalTargetSteamID, newJob.CurrentTargetSteamID, newJob.MaxLevel, newJob.CurrentLevel)
			configuration.Logger.Info(logMsg)

			Worker(cntr, newJob)
			d.Ack(false)
		}
	}
}

func CrawlUser(cntr controller.CntrInterface, steamID string, level int) {
	newJob := datastructures.Job{
		JobType:               "crawl",
		OriginalTargetSteamID: steamID,
		CurrentTargetSteamID:  steamID,
		MaxLevel:              level,
		CurrentLevel:          1,
	}
	jsonObj, err := json.Marshal(newJob)
	if err != nil {
		log.Fatal(err)
	}

	err = cntr.PublishToJobsQueue(jsonObj)
	if err != nil {
		logMsg := fmt.Sprintf("failed to publish new crawl user job with steamID: %s level: %d to queue: %+v",
			steamID, level, err)
		configuration.Logger.Error(logMsg)
	} else {
		configuration.Logger.Info(fmt.Sprintf("placed job steamID: %s level: %d into queue", steamID, level))
	}
}
