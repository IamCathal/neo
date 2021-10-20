package worker

import (
	"fmt"
	"log"

	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
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

	fmt.Println(friendsList)
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
func ControlFunc(cntr controller.CntrInterface, jobsChan <-chan datastructures.Job) {
	// wait for new job frmo queue
	for {
		newJob := <-jobsChan
		go Worker(cntr, newJob)
	}
}

func CrawlUser(cntr controller.CntrInterface, steamID int64, level int) {
	if jobsChannel == nil {
		jobsChannel = make(chan datastructures.Job, 10)
		newJob := datastructures.Job{
			JobType:               "crawl",
			OriginalTargetSteamID: steamID,
			CurrentTargetSteamID:  steamID,
			MaxLevel:              3,
			CurrentLevel:          1,
		}
		jobsChannel <- newJob
		go ControlFunc(cntr, jobsChannel)
	} else {
		go ControlFunc(cntr, jobsChannel)
	}
	// // push new crawl job to the queue
	// jobs := make(chan datastructures.Job, 20)
	// for {
	// 	// find a job from rabbitmQ and push it to the jobs queue
	// 	newJobFromQueue := datastructures.Job{}

	// 	jobs <- newJobFromQueue
	// }
}
