package worker

import (
	"encoding/json"
	"regexp"
	"strconv"

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

// func HasUserBeenCrawledBeforeAtThisLevel(steamID int64, level int) (bool, error) {

// }
