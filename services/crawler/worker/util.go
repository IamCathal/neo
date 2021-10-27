package worker

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/streadway/amqp"
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

func VerifyFormatOfSteamIDs(input datastructures.CrawlUsersInput) ([]int64, error) {
	validSteamIDs := []int64{}
	match, err := regexp.MatchString("([0-9]){17}", strconv.FormatInt(input.FirstSteamID, 10))
	if err != nil {
		return validSteamIDs, err
	}
	if match {
		validSteamIDs = append(validSteamIDs, input.FirstSteamID)
	}

	match, err = regexp.MatchString("([0-9]){17}", strconv.FormatInt(input.SecondSteamID, 10))
	if err != nil {
		return validSteamIDs, err
	}
	if match {
		validSteamIDs = append(validSteamIDs, input.SecondSteamID)
	}
	return validSteamIDs, nil
}

func putFriendsIntoQueue(currentJob datastructures.Job, friends []datastructures.Friend) error {
	for _, friend := range friends {
		nextLevel := currentJob.CurrentLevel + 1
		if nextLevel <= currentJob.MaxLevel {
			steamIDInt64, _ := strconv.ParseInt(friend.Steamid, 10, 64)
			newJob := datastructures.Job{
				JobType:               "crawl",
				OriginalTargetSteamID: currentJob.OriginalTargetSteamID,
				CurrentTargetSteamID:  steamIDInt64,

				MaxLevel:     currentJob.MaxLevel,
				CurrentLevel: nextLevel,
			}
			jsonObj, err := json.Marshal(newJob)
			if err != nil {
				return err
			}
			err = configuration.Channel.Publish(
				"",                       // exchange
				configuration.Queue.Name, // routing key
				false,                    // mandatory
				false,                    // immediate
				amqp.Publishing{
					ContentType: "text/json",
					Body:        []byte(jsonObj),
				})
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

func getPlayerSummaries(cntr controller.CntrInterface, friends datastructures.Friendslist) ([]datastructures.Player, error) {
	// Only 100 steamIDs can be queried per call
	steamIDs := extractSteamIDsfromFriendsList(friends)
	stacksOfSteamIDs := breakIntoStacksOf100OrLessSteamIDs(steamIDs)

	allPlayerSummaries := []datastructures.Player{}
	for i := 0; i < len(stacksOfSteamIDs); i++ {
		batchOfPlayerSummaries, err := cntr.CallGetPlayerSummaries(stacksOfSteamIDs)
		if err != nil {
			return []datastructures.Player{}, err
		}
		allPlayerSummaries = append(allPlayerSummaries, batchOfPlayerSummaries...)
	}

	onlyPublicProfiles := getPublicProfiles(allPlayerSummaries)
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
		if users[i].Communityvisibilitystate == 1 {
			publicProfiles = append(publicProfiles, users[i])
		}
	}
	return publicProfiles
}

// func HasUserBeenCrawledBeforeAtThisLevel(steamID int64, level int) (bool, error) {

// }
