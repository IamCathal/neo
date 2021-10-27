package worker

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
)

var (
	jobsChannel chan datastructures.Job
)

// Worker crawls the steam API to get data from steam for a given user
// e.g account details and details of a user's friend
func Worker(cntr controller.CntrInterface, job datastructures.Job) {
	friendsList, err := GetFriends(cntr, job.CurrentTargetSteamID)
	if err != nil {
		log.Fatal(err)
	}

	playerSummaryForCurrentUser, err := cntr.CallGetPlayerSummaries(job.CurrentTargetSteamID)
	if err != nil {
		configuration.Logger.Fatal(fmt.Sprintf("failed to get player summary for target user: %v", err.Error()))
		log.Fatal(err)
	}

	friendPlayerSummaries, err := getPlayerSummaries(cntr, job, friendsList)
	if err != nil {
		configuration.Logger.Fatal(fmt.Sprintf("failed to get player summaries for friends: %v", err.Error()))
		log.Fatal(err)
	}

	friendPlayerSummarySteamIDs := getSteamIDsFromPlayers(friendPlayerSummaries)
	err = putFriendsIntoQueue(job, friendPlayerSummarySteamIDs)
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

	// Print locally just for convenience
	// userDocument := datastructures.UserDocument{
	// 	SteamID:    job.CurrentTargetSteamID,
	// 	AccDetails: playerSummary[0],
	// 	FriendIDs:  extractSteamIDsFromPlayersList(friendPlayerSummaries),
	// }
	// yuppa, err := json.Marshal(userDocument)
	// if err != nil {
	// 	configuration.Logger.Fatal(err.Error())
	// 	panic(err)
	// }
	// fmt.Printf("\n\n\nThe data: \n %s\n\n", yuppa)
	logMsg := fmt.Sprintf("Got data for [%s][%s][%s] %d friends",
		playerSummaryForCurrentUser[0].Steamid, playerSummaryForCurrentUser[0].Personaname, playerSummaryForCurrentUser[0].Loccountrycode,
		len(friendPlayerSummaries))
	configuration.Logger.Info(logMsg)

	// Get game details for target user

	// Save game details to DB

	// // Save friendslist to DB
	// userIDWithFriendsList := datastructures.UserDetails{
	// 	SteamID: job.CurrentTargetSteamID,
	// 	Friends: friendsList,
	// }
	// success, err := cntr.SaveFriendsListToDataStore(userIDWithFriendsList)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// if !success {

	// }
}

// GetFriends gets the friendslist for a given user through either the steam web API
// or cache
func GetFriends(cntr controller.CntrInterface, steamID string) (datastructures.Friendslist, error) {
	// First call the db

	friendsList, err := cntr.CallGetFriends(steamID)
	if err != nil {
		return datastructures.Friendslist{}, err
	}
	// fmt.Println("Returned obj:")
	// fmt.Printf("%+v\n\n", friendsList)
	return friendsList, nil
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
