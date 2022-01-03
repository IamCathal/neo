package graphing

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/go-echarts/go-echarts/charts"
	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
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

	links []charts.GraphLink
	nodes []charts.GraphNode
}

// func graphWorker(id int, ctx context.Context, cntr controller.CntrInterface, workerConfig *GraphWorkerConfig, jobs <-chan jobStruct, res chan<- jobStruct) {
// 	for {
// 		workerConfig.usersCrawledMutex.Lock()
// 		// fmt.Println("lock userscrawled in worker")
// 		if workerConfig.UsersCrawled >= workerConfig.TotalUsersToCrawl &&
// 			workerConfig.TotalUsersToCrawl != 0 {
// 			configuration.Logger.Info(fmt.Sprint(id) + " graphWorker exiting.........")
// 			cancel()
// 		}
// 		workerConfig.usersCrawledMutex.Unlock()
// 		// fmt.Println("unlock userscrawled in worker")

// 		// Take from jobs
// 		// workerConfig.jobMutex.Lock()
// 		fmt.Printf("[%d] waiting for job...\n", id)
// 		currentJob := <-jobs
// 		emptyJob := jobStruct{}
// 		if currentJob == emptyJob {
// 			fmt.Println("EMPTY JOB")
// 			break
// 		}
// 		// workerConfig.jobMutex.Unlock()

// 		fmt.Printf("[%d] Got job in worker: %+v\n", id, currentJob)
// 		userGraphData, err := cntr.GetGraphableDataFromDataStore(currentJob.SteamID)
// 		if err != nil {
// 			configuration.Logger.Sugar().Fatalf("failed to get graphable data for %s: %+v", currentJob.SteamID, err)
// 			panic(err)
// 		}

// 		// fmt.Printf("graphable data: %+v\n", userGraphData)

// 		workerConfig.usersCrawledMutex.Lock()
// 		workerConfig.UsersCrawled++
// 		workerConfig.usersCrawledMutex.Unlock()

// 		if currentJob.CurrentLevel+1 <= workerConfig.MaxLevel {
// 			for _, friendID := range userGraphData.FriendIDs {
// 				newJob := jobStruct{
// 					SteamID: friendID,
// 					// Username: ,
// 					FromID:       currentJob.SteamID,
// 					FromUsername: userGraphData.Username,
// 					CurrentLevel: currentJob.CurrentLevel + 1,
// 					MaxLevel:     currentJob.MaxLevel,
// 				}
// 				workerConfig.resMutex.Lock()
// 				// fmt.Println("lock res in worker")
// 				res <- newJob
// 				// fmt.Printf("placed new job %+v\n", newJob)
// 				workerConfig.resMutex.Unlock()
// 				// fmt.Println("unlock res in worker")
// 				time.Sleep(2 * time.Millisecond)
// 			}
// 			fmt.Println("no more friends to place")
// 		} else {
// 			fmt.Printf("friends are too high level userscrawled: %d, totalusers: %d\n", workerConfig.UsersCrawled, workerConfig.TotalUsersToCrawl)
// 		}
// 	}
// 	configuration.Logger.Info(fmt.Sprint(id) + " graphWorker exiting..")
// }

// func graphWorker(id int, stopSignal <-chan bool, cntr controller.CntrInterface, wg *sync.WaitGroup, workerConfig *GraphWorkerConfig, jobs <-chan jobStruct, res chan<- jobStruct) {
// 	configuration.Logger.Sugar().Infof("%d graphWorker starting...\n", id)
// 	for {
// 		select {
// 		case <-stopSignal:
// 			configuration.Logger.Sugar().Infof("%d graphWorker exiting...\n", id)
// 			wg.Done()
// 			return
// 		case currentJob := <-jobs:
// 			emptyJob := jobStruct{}
// 			if currentJob == emptyJob {
// 				panic("EMPTY JOB, most likely means channel was closed and read from")
// 			}

// 			configuration.Logger.Sugar().Infof("[ID:%d][jobs:%d][res:%d] worker received job: %+v",
// 				id, len(jobs), len(res), currentJob)

// 			userGraphData, err := cntr.GetGraphableDataFromDataStore(currentJob.SteamID)
// 			if err != nil {
// 				configuration.Logger.Sugar().Fatalf("failed to get graphable data for %s: %+v", currentJob.SteamID, err)
// 				panic(err)
// 			}

// 			workerConfig.usersCrawledMutex.Lock()
// 			workerConfig.UsersCrawled++
// 			workerConfig.usersCrawledMutex.Unlock()

// 			if currentJob.CurrentLevel+1 <= workerConfig.MaxLevel {
// 				for _, friendID := range userGraphData.FriendIDs {
// 					newJob := jobStruct{
// 						SteamID: friendID,
// 						// Username: ,
// 						FromID:       userGraphData.SteamID,
// 						FromUsername: userGraphData.Username,
// 						CurrentLevel: currentJob.CurrentLevel + 1,
// 						MaxLevel:     currentJob.MaxLevel,
// 					}
// 					workerConfig.resMutex.Lock()
// 					res <- newJob
// 					workerConfig.resMutex.Unlock()
// 					time.Sleep(2 * time.Millisecond)
// 				}
// 			}
// 		}
// 	}
// }

func graphWorker(id int, stopSignal <-chan bool, cntr controller.CntrInterface, wg *sync.WaitGroup, workerConfig *GraphWorkerConfig, jobs <-chan datastructures.CrawlJob, res chan<- datastructures.ResStruct) {
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
				configuration.Logger.Sugar().Fatalf("failed tgo get user data for %s: %+v", currentJob.SteamID, err)
				panic(err)
			}

			if currentJob.CurrentLevel <= workerConfig.MaxLevel {
				newJob := datastructures.ResStruct{
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

func Control2Func(cntr controller.CntrInterface, steamID string, workerConfig GraphWorkerConfig) ([]datastructures.ResStruct, error) {
	jobsChan := make(chan datastructures.CrawlJob, 70000)
	resChan := make(chan datastructures.ResStruct, 70000)

	var jobMutex sync.Mutex
	var resMutex sync.Mutex
	var wg sync.WaitGroup
	var usersCrawledMutex sync.Mutex
	workerConfig.jobMutex = &jobMutex
	workerConfig.resMutex = &resMutex
	workerConfig.usersCrawledMutex = &usersCrawledMutex

	allUsersGraphData := []datastructures.ResStruct{}

	firstJob := datastructures.CrawlJob{
		SteamID:      steamID,
		FromID:       steamID,
		CurrentLevel: 1,
		MaxLevel:     workerConfig.MaxLevel,
	}
	jobsChan <- firstJob

	workerAmount := 3
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
			return []datastructures.ResStruct{}, err
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

	usersDataForGraphWithOnlyTop5Games := []datastructures.ResStruct{}
	for _, friend := range usersDataForGraph {
		if len(friend.User.GamesOwned) >= 5 {
			friend.User.GamesOwned = friend.User.GamesOwned[:5]
		}
		usersDataForGraphWithOnlyTop5Games = append(usersDataForGraphWithOnlyTop5Games, friend)
	}

	usersDataForGraphWithFriends := datastructures.UsersGraphData{
		UserDetails:   usersDataForGraphWithOnlyTop5Games[0],
		FriendDetails: usersDataForGraphWithOnlyTop5Games[1:],
	}

	json, err := json.Marshal(usersDataForGraphWithFriends)
	if err != nil {
		configuration.Logger.Fatal("failed to marshal json for user graph data")
		panic(err)
	}

	err = os.WriteFile("hellok.json", json, 0644)
	if err != nil {
		configuration.Logger.Fatal("failed to write json for user graph data")
		panic(err)
	}

	success, err := cntr.SaveProcessedGraphDataToDataStore(crawlID, usersDataForGraphWithFriends)
	if err != nil || success == false {
		configuration.Logger.Sugar().Fatalf("failed to save processed graph data to datastore: %+v", err)
		panic(err)
	}
}

// func ControlFunc(cntr controller.CntrInterface, steamID string, workerConfig GraphWorkerConfig) ([]jobStruct, error) {
// 	jobsChan := make(chan jobStruct, 25000)
// 	resChan := make(chan jobStruct, 25000)

// 	var jobMutex sync.Mutex
// 	var resMutex sync.Mutex
// 	var wg sync.WaitGroup
// 	var usersCrawledMutex sync.Mutex
// 	workerConfig.jobMutex = &jobMutex
// 	workerConfig.resMutex = &resMutex
// 	workerConfig.usersCrawledMutex = &usersCrawledMutex

// 	allUsersGraphData := []jobStruct{}

// 	firstJob := jobStruct{
// 		SteamID:      steamID,
// 		MaxLevel:     workerConfig.MaxLevel,
// 		CurrentLevel: 1,
// 		FromID:       steamID,
// 	}
// 	jobsChan <- firstJob
// 	allUsersGraphData = append(allUsersGraphData, firstJob)

// 	workerAmount := 2
// 	var stopSignal chan bool = make(chan bool, 0)
// 	workersAreDone := false
// 	oneOrMoreUsersHasNoUsername := false

// 	for i := 0; i < workerAmount; i++ {
// 		wg.Add(1)
// 		go graphWorker(i, stopSignal, cntr, &wg, &workerConfig, jobsChan, resChan)
// 	}

// 	for {
// 		if workersAreDone {
// 			break
// 		}
// 		if workerConfig.UsersCrawled >= workerConfig.TotalUsersToCrawl &&
// 			workerConfig.TotalUsersToCrawl != 0 {
// 			workersAreDone = true
// 			for i := 0; i < workerAmount; i++ {
// 				stopSignal <- true
// 			}
// 			workersAreDone = true
// 		}
// 		if workersAreDone {
// 			break
// 		}

// 		select {
// 		case res := <-resChan:
// 			fmt.Printf("got new job in controlfunc: %+v\n", res)
// 			if res.Username == "" && !oneOrMoreUsersHasNoUsername {
// 				fmt.Printf("%+v\n", res)
// 				oneOrMoreUsersHasNoUsername = true
// 			}
// 			allUsersGraphData = append(allUsersGraphData, res)
// 			if res.CurrentLevel <= res.MaxLevel {
// 				workerConfig.jobMutex.Lock()
// 				jobsChan <- res
// 				workerConfig.jobMutex.Unlock()
// 			}
// 		default:
// 			temp := false
// 			if temp {
// 				temp = false
// 			}
// 		}
// 	}

// 	close(jobsChan)
// 	close(resChan)

// 	configuration.Logger.Info("waiting for all to be done")
// 	wg.Wait()
// 	configuration.Logger.Sugar().Infof("all %d users have been found", len(allUsersGraphData))

// 	if oneOrMoreUsersHasNoUsername {
// 		configuration.Logger.Info("one or more users had no username, retrieving and correlating all usernames now")
// 		steamIDsWithoutAssociatedUsernames := getAllSteamIDsFromJobsWithNoAssociatedUsernames(allUsersGraphData)
// 		steamIDsToUsernames, err := cntr.GetUsernamesForSteamIDs(steamIDsWithoutAssociatedUsernames)
// 		if err != nil {
// 			return []jobStruct{}, err
// 		}

// 		for _, job := range allUsersGraphData {
// 			if job.Username == "" {
// 				job.Username = steamIDsToUsernames[job.SteamID]
// 			}
// 		}
// 	}

// 	return allUsersGraphData, nil
// }
