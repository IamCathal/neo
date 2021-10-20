package worker

import (
	"fmt"
	"log"

	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
)

// // Worker is a worker pool function that processes jobs asynchronously
// func Worker(job datastructures.Job) {
// 	friendsObj, err := GetFriends(job.CurrentTargetSteamID)
// }

// GetFriends gets the friendslist for a given user through either the steam web API
// or cache
func GetFriends(cntr controller.CntrInterface, steamID int64) datastructures.Friendslist {
	// First call the db

	userObj, err := cntr.CallGetFriends(steamID)
	if err != nil {
		log.Fatal(err)
	}
	// Save to DB

	fmt.Println("Returned obj:")
	fmt.Printf("%+v\n\n", userObj)
	return userObj
}

// // ControlFunc manages workers
// func ControlFunc(jobsChan <-chan datastructures.Job) {
// 	// wait for new job frmo queue
// 	for {
// 		newJob := <-jobsChan
// 		go Worker(newJob)
// 	}
// }

func CrawlUser(cntr controller.CntrInterface, steamID int64, level int) datastructures.Friendslist {
	// // push new crawl job to the queue
	// jobs := make(chan datastructures.Job, 20)
	// for {
	// 	// find a job from rabbitmQ and push it to the jobs queue
	// 	newJobFromQueue := datastructures.Job{}

	// 	jobs <- newJobFromQueue
	// }

	return GetFriends(cntr, steamID)
}
