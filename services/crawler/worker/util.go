package worker

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
)

func InitWorkerConfig() datastructures.WorkerConfig {
	return datastructures.WorkerConfig{
		WorkerAmount: configuration.WorkerConfig.WorkerAmount,
	}
}

func StartUpWorkers(cntr controller.CntrInterface) {
	for i := 0; i < len(configuration.UsableAPIKeys.APIKeys); i++ {
		go ControlFunc(cntr)
	}
}

func VerifyFormatOfSteamIDs(input datastructures.CrawlUsersInput) ([]string, error) {
	validSteamIDs := []string{}
	match, err := regexp.MatchString("([0-9]){17}", input.FirstSteamID)
	if err != nil {
		return validSteamIDs, err
	}
	if match {
		validSteamIDs = append(validSteamIDs, input.FirstSteamID)
	}

	match, err = regexp.MatchString("([0-9]){17}", input.SecondSteamID)
	if err != nil {
		return validSteamIDs, err
	}
	if match {
		validSteamIDs = append(validSteamIDs, input.SecondSteamID)
	}
	return validSteamIDs, nil
}

func putFriendsIntoQueue(cntr controller.CntrInterface, currentJob datastructures.Job, friendIDs []string) error {
	for _, ID := range friendIDs {
		nextLevel := currentJob.CurrentLevel + 1
		if nextLevel <= currentJob.MaxLevel {
			newJob := datastructures.Job{
				JobType:               "crawl",
				OriginalTargetSteamID: currentJob.OriginalTargetSteamID,
				CurrentTargetSteamID:  ID,

				MaxLevel:     currentJob.MaxLevel,
				CurrentLevel: nextLevel,
			}
			jsonObj, err := json.Marshal(newJob)
			if err != nil {
				return err
			}
			configuration.Logger.Info(fmt.Sprintf("pushing job from: %+v", newJob))
			err = cntr.PublishToJobsQueue(jsonObj)
			if err != nil {
				return err
			}
			// configuration.Logger.Info(fmt.Sprintf("placed job %s:%d into queue", friend.Steamid, job.CurrentLevel))
		} else {
			// configuration.Logger.Info(fmt.Sprintf("job %d:%d was not published", job.CurrentTargetSteamID, job.CurrentLevel))
		}

	}
	return nil
}

func getPlayerSummaries(cntr controller.CntrInterface, job datastructures.Job, friends datastructures.Friendslist) ([]datastructures.Player, error) {
	// Only 100 steamIDs can be queried per call
	steamIDs := extractSteamIDsfromFriendsList(friends)
	stacksOfSteamIDs := breakIntoStacksOf100OrLessSteamIDs(steamIDs)
	configuration.Logger.Info(fmt.Sprintf("%s has %d private and public friends", job.CurrentTargetSteamID, len(steamIDs)))

	allPlayerSummaries := []datastructures.Player{}
	for i := 0; i < len(stacksOfSteamIDs); i++ {
		batchOfPlayerSummaries, err := cntr.CallGetPlayerSummaries(stacksOfSteamIDs[i])
		if err != nil {
			return []datastructures.Player{}, err
		}
		allPlayerSummaries = append(allPlayerSummaries, batchOfPlayerSummaries...)
	}

	onlyPublicProfiles := getPublicProfiles(allPlayerSummaries)
	configuration.Logger.Info(fmt.Sprintf("%d/%d public profiles from user %s", len(onlyPublicProfiles), len(allPlayerSummaries), job.CurrentTargetSteamID))
	return onlyPublicProfiles, nil
}

// breakIntoStacksOf100OrLessSteaMIDs divides a list of steam IDs into stacks of one
// hundred IDs or less. The GetPlayerSummary API only accepts up to 100 steam IDs per
// call
// Returns lists of formatted IDs that the steam API will take like
// {"1233,4324,5435,5677,2432","34689,1035,7847,4673,9384"}
func breakIntoStacksOf100OrLessSteamIDs(friendIDs []string) []string {
	totalFriendCount := len(friendIDs)
	steamIDList := []string{}

	fullOneHundredIDLists, remainder := divideAndGetRemainder(totalFriendCount, 100)
	// If less than one hundred total IDs
	if fullOneHundredIDLists == 0 {
		idList := []string{}
		for _, ID := range friendIDs {
			idList = append(idList, ID)
		}
		steamIDList = append(steamIDList, strings.Join(idList, ","))
		return steamIDList
	}

	// There are 100 or more total IDs
	for i := 0; i < fullOneHundredIDLists; i++ {
		// For each batch of 100 users
		idList := []string{}
		for j := 0; j < 100; j++ {
			idList = append(idList, friendIDs[j+(i*100)])
		}
		steamIDList = append(steamIDList, strings.Join(idList, ","))
	}
	if remainder == 0 {
		// There were a clean multiple of 100 steamIDs
		return steamIDList
	}

	// There was not a clean multiple of 100 steamIDs
	firstIndex := fullOneHundredIDLists * 100
	idList := []string{}
	for i := firstIndex; i < firstIndex+remainder; i++ {
		idList = append(idList, friendIDs[i])
	}
	steamIDList = append(steamIDList, strings.Join(idList, ","))

	return steamIDList
}

func divideAndGetRemainder(numerator, denominator int) (quotient, remainder int) {
	quotient = numerator / denominator
	remainder = numerator % denominator
	return
}

func extractSteamIDsFromPlayersList(friends []datastructures.Player) []string {
	steamIDs := []string{}
	for _, friend := range friends {
		steamIDs = append(steamIDs, friend.Steamid)
	}
	return steamIDs
}

func extractSteamIDsfromFriendsList(friends datastructures.Friendslist) []string {
	steamIDs := []string{}
	for _, friend := range friends.Friends {
		steamIDs = append(steamIDs, friend.Steamid)
	}
	return steamIDs
}

func getPublicProfiles(users []datastructures.Player) []datastructures.Player {
	publicProfiles := []datastructures.Player{}
	for i := 0; i < len(users); i++ {
		// Visibility state of 1 or 2 means some level of privacy
		if users[i].Communityvisibilitystate == 3 {
			publicProfiles = append(publicProfiles, users[i])
		}
	}
	return publicProfiles
}

func getSteamIDsFromPlayers(users []datastructures.Player) []string {
	steamIDs := []string{}
	for _, user := range users {
		steamIDs = append(steamIDs, user.Steamid)
	}
	return steamIDs
}

func getUsersProfileSummaryFromSlice(steamID string, playerSummaries []datastructures.Player) (bool, datastructures.Player) {
	for _, player := range playerSummaries {
		if player.Steamid == steamID {
			return true, player
		}
	}
	return false, datastructures.Player{}
}

// func HasUserBeenCrawledBeforeAtThisLevel(steamID int64, level int) (bool, error) {

// }
