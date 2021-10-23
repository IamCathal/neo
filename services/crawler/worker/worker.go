package worker

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/streadway/amqp"
)

var (
	jobsChannel chan datastructures.Job
)

// Worker is a worker pool function that processes jobs asynchronously
func Worker(cntr controller.CntrInterface, job datastructures.Job) {
	// Get Friends
	friendsList, err := GetFriends(cntr, job.CurrentTargetSteamID)
	if err != nil {
		log.Fatal(err)
	}
	configuration.Logger.Info(fmt.Sprintf("Got %d friends at level %d", len(friendsList.Friends), job.CurrentLevel))
	for _, friend := range friendsList.Friends {
		nextLevel := job.CurrentLevel + 1
		if nextLevel <= job.MaxLevel {
			steamIDInt64, _ := strconv.ParseInt(friend.Steamid, 10, 64)
			newJob := datastructures.Job{
				JobType:               "crawl",
				OriginalTargetSteamID: job.OriginalTargetSteamID,
				CurrentTargetSteamID:  steamIDInt64,

				MaxLevel:     job.MaxLevel,
				CurrentLevel: nextLevel,
			}
			jsonObj, err := json.Marshal(newJob)
			if err != nil {
				log.Fatal(err)
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
			// configuration.Logger.Info(fmt.Sprintf("placed job %s:%d into queue", friend.Steamid, job.CurrentLevel))
		} else {
			// configuration.Logger.Info(fmt.Sprintf("job %d:%d was not published", job.CurrentTargetSteamID, job.CurrentLevel))
		}

	}
	// // Save to DB
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
func GetFriends(cntr controller.CntrInterface, steamID int64) (datastructures.Friendslist, error) {
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
	msgs, err := configuration.Channel.Consume(
		configuration.Queue.Name, // queue
		"",                       // consumer
		false,                    // auto-ack
		false,                    // exclusive
		false,                    // no-local
		false,                    // no-wait
		nil,                      // args
	)
	if err != nil {
		log.Fatal(err)
	}
	// wait for new job from queue
	for {
		for d := range msgs {
			newJob := datastructures.Job{}
			err := json.Unmarshal(d.Body, &newJob)
			if err != nil {
				log.Fatal(err)
			}
			Worker(cntr, newJob)
			d.Ack(false)
		}
	}
}

func CrawlUser(cntr controller.CntrInterface, steamID int64, level int) {
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

	err = configuration.Channel.Publish(
		"",                       // exchange
		configuration.Queue.Name, // routing key
		false,                    // mandatory
		false,                    // immediate
		amqp.Publishing{
			ContentType: "text/json",
			Body:        []byte(jsonObj),
		})
	configuration.Logger.Info(fmt.Sprintf("placed job %d:%d into queue", steamID, level))
}
