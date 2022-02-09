package graphing

import (
	"sync"
	"time"

	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/neosteamfriendgraphing/common"
)

type jobStruct struct {
	SteamID      string
	Username     string
	FromID       string
	FromUsername string
	MaxLevel     int
	CurrentLevel int
}

type GraphWorkerConfig struct {
	jobMutex *sync.Mutex
	resMutex *sync.Mutex

	// CrawlingStatus related variables
	usersCrawledMutex *sync.Mutex
	TotalUsersToCrawl int
	UsersCrawled      int
	MaxLevel          int
}

func graphWorker(id int, stopSignal <-chan bool, cntr controller.CntrInterface, wg *sync.WaitGroup, workerConfig *GraphWorkerConfig, jobs <-chan datastructures.CrawlJob, res chan<- common.UsersGraphInformation) {
	configuration.Logger.Sugar().Infof("%d graphWorker starting...\n", id)
	for {
		select {
		case <-stopSignal:
			configuration.Logger.Sugar().Infof("%d graphWorker exiting...\n", id)
			wg.Done()
			return
		case currentJob := <-jobs:
			emptyJob := datastructures.CrawlJob{}
			if currentJob == emptyJob {
				panic("EMPTY JOB, most likely means channel was closed and read from")
			}

			configuration.Logger.Sugar().Infof("[ID:%d][jobs:%d][res:%d] worker received job: %+v",
				id, len(jobs), len(res), currentJob)

			userGraphData, err := cntr.GetUserFromDataStore(currentJob.SteamID)
			if err != nil {
				configuration.Logger.Sugar().Fatalf("failed to get user data for %s: %+v", currentJob.SteamID, err)
				panic(err)
			}

			if currentJob.CurrentLevel <= workerConfig.MaxLevel {
				newJob := common.UsersGraphInformation{
					User:         userGraphData,
					FromID:       currentJob.FromID,
					CurrentLevel: currentJob.CurrentLevel,
					MaxLevel:     currentJob.MaxLevel,
				}
				workerConfig.resMutex.Lock()
				res <- newJob
				workerConfig.resMutex.Unlock()
				time.Sleep(2 * time.Millisecond)
			}

			workerConfig.usersCrawledMutex.Lock()
			workerConfig.UsersCrawled++
			workerConfig.usersCrawledMutex.Unlock()
		}
	}
}

func Control2Func(cntr controller.CntrInterface, steamID string, workerConfig GraphWorkerConfig) ([]common.UsersGraphInformation, error) {
	jobsChan := make(chan datastructures.CrawlJob, 70000)
	resChan := make(chan common.UsersGraphInformation, 70000)

	var jobMutex sync.Mutex
	var resMutex sync.Mutex
	var wg sync.WaitGroup
	var usersCrawledMutex sync.Mutex
	workerConfig.jobMutex = &jobMutex
	workerConfig.resMutex = &resMutex
	workerConfig.usersCrawledMutex = &usersCrawledMutex

	allUsersGraphData := []common.UsersGraphInformation{}

	firstJob := datastructures.CrawlJob{
		SteamID:      steamID,
		FromID:       steamID,
		CurrentLevel: 1,
		MaxLevel:     workerConfig.MaxLevel,
	}
	jobsChan <- firstJob

	workerAmount := 6
	var stopSignal chan bool = make(chan bool, 0)
	workersAreDone := false
	oneOrMoreUsersHasNoUsername := false

	for i := 0; i < workerAmount; i++ {
		wg.Add(1)
		go graphWorker(i, stopSignal, cntr, &wg, &workerConfig, jobsChan, resChan)
	}

	for {
		if workersAreDone {
			break
		}
		if workerConfig.UsersCrawled >= workerConfig.TotalUsersToCrawl &&
			workerConfig.TotalUsersToCrawl != 0 {
			workersAreDone = true
			for i := 0; i < workerAmount; i++ {
				stopSignal <- true
			}
			workersAreDone = true
		}
		if workersAreDone {
			break
		}

		select {
		case res := <-resChan:
			if res.User.AccDetails.Personaname == "" && !oneOrMoreUsersHasNoUsername {
				oneOrMoreUsersHasNoUsername = true
			}

			allUsersGraphData = append(allUsersGraphData, res)
			if res.CurrentLevel < res.MaxLevel {
				for _, friendID := range res.User.FriendIDs {
					newCrawlJob := datastructures.CrawlJob{
						SteamID:      friendID,
						FromID:       res.User.AccDetails.SteamID,
						CurrentLevel: res.CurrentLevel + 1,
						MaxLevel:     res.MaxLevel,
					}
					workerConfig.jobMutex.Lock()
					jobsChan <- newCrawlJob
					workerConfig.jobMutex.Unlock()
				}
			}
		default:
			temp := false
			if temp {
				temp = false
			}
		}
	}

	close(jobsChan)
	close(resChan)

	configuration.Logger.Info("waiting for all jobs to be done")
	wg.Wait()
	configuration.Logger.Sugar().Infof("all %d users have been found", len(allUsersGraphData))

	if oneOrMoreUsersHasNoUsername {
		configuration.Logger.Info("one or more users had no username, retrieving and correlating all usernames now")
		steamIDsWithoutAssociatedUsernames := getAllSteamIDsFromJobsWithNoAssociatedUsernames(allUsersGraphData)
		steamIDsToUsernames, err := cntr.GetUsernamesForSteamIDs(steamIDsWithoutAssociatedUsernames)
		if err != nil {
			return []common.UsersGraphInformation{}, err
		}

		for _, job := range allUsersGraphData {
			if job.User.AccDetails.Personaname == "" {
				job.User.AccDetails.Personaname = steamIDsToUsernames[job.User.AccDetails.SteamID]
			}
		}
	}
	return allUsersGraphData, nil
}

func CollectGraphData(cntr controller.CntrInterface, steamID, crawlID string, workerConfig GraphWorkerConfig) {
	usersDataForGraph, err := Control2Func(cntr, steamID, workerConfig)
	if err != nil {
		configuration.Logger.Sugar().Fatalf("failed to gather data for crawlID %s: %+v", crawlID, err)
		panic(err)
	}

	usersDataForGraphWithOnlyTop40Games := []common.UsersGraphInformation{}
	for _, friend := range usersDataForGraph {
		if len(friend.User.GamesOwned) >= 40 {
			friend.User.GamesOwned = friend.User.GamesOwned[:40]
		}
		usersDataForGraphWithOnlyTop40Games = append(usersDataForGraphWithOnlyTop40Games, friend)
	}

	topOverallGameDetails, err := getTopTenOverallGameNames(cntr, usersDataForGraphWithOnlyTop40Games)
	if err != nil {
		configuration.Logger.Sugar().Fatalf("failed to get top 10 game detail: %+v", err)
		panic(err)
	}

	usersDataForGraphWithFriends := common.UsersGraphData{
		UserDetails:    usersDataForGraphWithOnlyTop40Games[0],
		FriendDetails:  usersDataForGraphWithOnlyTop40Games[1:],
		TopGameDetails: topOverallGameDetails,
	}

	success, err := cntr.SaveProcessedGraphDataToDataStore(crawlID, usersDataForGraphWithFriends)
	if err != nil || success == false {
		configuration.Logger.Sugar().Fatalf("failed to save processed graph data to datastore: %+v", err)
		panic(err)
	}
	configuration.Logger.Sugar().Infof("successfully collected graph data for crawlID: %s", crawlID)
}
