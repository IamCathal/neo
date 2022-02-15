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

func publishJobWithThrottling(cntr controller.CntrInterface, job datastructures.Job) error {
	jobPublishLock.Lock()
	for {
		if time.Since(lastPublishedTime) > time.Duration(2*time.Nanosecond) {
			lastPublishedTime = time.Now()

			err := publishJob(cntr, job)
			jobPublishLock.Unlock()
			if err != nil {
				return err
			}
			return nil
		}
	}
}

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
		sleepTimers := []int{80, 500, 8500}

		for i := 0; i < maxRetries; i++ {
			cntr.Sleep(time.Duration(sleepTimers[i]) * time.Millisecond)
			err = amqpchannelmanager.PublishToJobsQueue(cntr, jobJSON)
			if err == nil {
				configuration.Logger.Sugar().Infof("successfully placed job in queue after %d retries", i)
				successfulRequest = true
				break
			}
			configuration.Logger.Info(fmt.Sprintf("failed to publish job to queue. Sleeping for %v ms Retrying for the %d time: %+v", sleepTimers[i], i, job))
		}

		if !successfulRequest {
			configuration.Logger.Sugar().Errorf("failed to publish job to queue after %v and retrying %d times with job: %+v", time.Since(startTime), maxRetries, job)
			return err
		}
	}
	return nil
}
