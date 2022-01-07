package worker

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/iamcathal/neo/services/crawler/amqpchannelmanager"
	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
)

var (
	jobPublishLock    sync.Mutex
	lastPublishedTime time.Time
)

func init() {
	lastPublishedTime = time.Now()
}

// GetSteamAPIKey gets a steam API key. It picks any steam API key
// stored that has not been used in the last $KEY_SLEEP_TIME ms,
// If none are found then the function waits a short period
// and tries again until one is returned.
func publishJobWithThrottling(cntr controller.CntrInterface, job datastructures.Job) error {
	startTime := time.Now()
	jobPublishLock.Lock()
	for {
		// if time.Since(lastPublishedTime) > time.Duration(25*time.Millisecond) {
		if time.Since(lastPublishedTime) > time.Duration(2*time.Nanosecond) {
			lastPublishedTime = time.Now()

			// go func() {
			// 	// Ensure the unlock call is not waiting for
			// 	// the response from the publish request
			// 	time.Sleep(2 * time.Millisecond)
			// 	jobPublishLock.Unlock()
			// }()

			fmt.Println(time.Since(startTime))
			err := publishJob(cntr, job)
			jobPublishLock.Unlock()
			if err != nil {
				return err
			}
			return nil
		}
	}
}

// func publishJobWithThrottling(cntr controller.CntrInterface, job datastructures.Job) error {
// 	startTime := time.Now()
// 	jobPublishLock.Lock()
// 	for {
// 		if time.Since(lastPublishedTime) > time.Duration(25*time.Millisecond) {
// 			lastPublishedTime = time.Now()

// 			go func() {
// 				// Ensure the unlock call is not waiting for
// 				// the response from the publish request
// 				time.Sleep(2 * time.Millisecond)
// 				jobPublishLock.Unlock()
// 			}()

// 			fmt.Println(time.Since(startTime))
// 			err := publishJob(cntr, job)
// 			if err != nil {
// 				return err
// 			}
// 			return nil
// 		}
// 	}
// }

func publishJob(cntr controller.CntrInterface, job datastructures.Job) error {
	startTime := time.Now()

	jobJSON, err := json.Marshal(job)
	if err != nil {
		return err
	}

	err = amqpchannelmanager.PublishToJobsQueue(cntr, jobJSON)
	if err != nil {
		configuration.Logger.Sugar().Infof("failed to publish job: %+v retrying now", string(jobJSON))
		maxRetries := 3
		successfulRequest := false
		sleepTimers := []int{80, 18500, 8500}

		for i := 0; i < maxRetries; i++ {
			time.Sleep(time.Duration(sleepTimers[i]) * time.Millisecond)
			err = amqpchannelmanager.PublishToJobsQueue(cntr, jobJSON)
			if err == nil {
				configuration.Logger.Sugar().Infof("successfully placed job in queue after %d retries", i)
				successfulRequest = true
				break
			}
			// configuration.Logger.Info(fmt.Sprintf("failed to publish job to queue. Sleeping for %v ms Retrying for the %d time: %+v", exponentialBackOffSleepTime, i, newJob))
			configuration.Logger.Info(fmt.Sprintf("failed to publish job to queue. Sleeping for %v ms Retrying for the %d time: %+v", sleepTimers[i], i, job))
		}

		if !successfulRequest {
			configuration.Logger.Error(fmt.Sprintf("failed to publish job to queue after %v and retrying %d times with job: %+v", time.Since(startTime), maxRetries, job))
			return err
		}
	}
	return nil
}
