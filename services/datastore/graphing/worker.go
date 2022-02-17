package graphing

import (
	"fmt"
	"sync"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/controller"
	dijkstra "github.com/iamcathal/dijkstra2"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/util"
)

type CrawlJob struct {
	FromID int64
	ToID   int64
}

type GraphWorkerConfig struct {
	jobMutex *sync.Mutex
	resMutex *sync.Mutex

	steamIDToUser    map[string]common.UserDocument
	steamIDToGraphID map[int64]int
	graphIDToSteamID map[int]int64

	// CrawlingStatus related variables
	usersCrawledMutex *sync.Mutex
	UsersCrawled      int
	TotalUsersToCrawl int
}

func graphWorker(id int, stopSignal <-chan bool, cntr controller.CntrInterface, wg *sync.WaitGroup, workerConfig *GraphWorkerConfig, jobs <-chan CrawlJob, res chan<- CrawlJob) {
	configuration.Logger.Sugar().Infof("%d dijkstra graphWorker starting...\n", id)
	for {
		select {
		case <-stopSignal:
			configuration.Logger.Sugar().Infof("%d dijkstra graphWorker exiting...\n", id)
			wg.Done()
			return
		case currentJob := <-jobs:
			emptyJob := CrawlJob{}
			if currentJob == emptyJob {
				panic("EMPTY JOB, most likely means channel was closed and read from")
			}

			toUser := workerConfig.steamIDToUser[fmt.Sprint(currentJob.FromID)]
			for _, friendID := range toUser.FriendIDs {
				// If the user is within range and has a userDocument saved already
				if exists := ifKeyExists(friendID, workerConfig.steamIDToUser); exists {
					newJob := CrawlJob{
						FromID: currentJob.FromID,
						ToID:   toInt64(friendID),
					}
					workerConfig.resMutex.Lock()
					res <- newJob
					workerConfig.resMutex.Unlock()
				}
			}

			workerConfig.usersCrawledMutex.Lock()
			workerConfig.UsersCrawled++
			workerConfig.usersCrawledMutex.Unlock()
		}
	}
}

func GetShortestPathIDs(cntr controller.CntrInterface, userOne, userTwo common.UsersGraphData) (bool, []int64, error) {
	maxChanLen := 100000
	jobsChan := make(chan CrawlJob, maxChanLen)
	resChan := make(chan CrawlJob, maxChanLen)

	workerConfig := GraphWorkerConfig{}
	var jobMutex sync.Mutex
	var resMutex sync.Mutex
	var wg sync.WaitGroup
	var usersCrawledMutex sync.Mutex
	workerConfig.jobMutex = &jobMutex
	workerConfig.resMutex = &resMutex
	workerConfig.usersCrawledMutex = &usersCrawledMutex
	workerConfig.UsersCrawled = 0

	workerConfig.steamIDToUser = GetIDToUserMap(userOne, userTwo)
	usersToCrawl := len(workerConfig.steamIDToUser)

	workerConfig.TotalUsersToCrawl = usersToCrawl
	mainUserSteamID := toInt64(userOne.UserDetails.User.AccDetails.SteamID)
	targetUserSteamID := toInt64(userTwo.UserDetails.User.AccDetails.SteamID)

	workerAmount := 2
	var stopSignal chan bool = make(chan bool, 0)
	workersAreDone := false

	for i := 0; i < workerAmount; i++ {
		wg.Add(1)
		go graphWorker(i, stopSignal, cntr, &wg, &workerConfig, jobsChan, resChan)
	}

	graph := dijkstra.NewGraph()

	workerConfig.graphIDToSteamID = make(map[int]int64)
	workerConfig.steamIDToGraphID = make(map[int64]int)
	currGraphID := 0

	workerConfig.steamIDToGraphID[mainUserSteamID] = currGraphID
	workerConfig.graphIDToSteamID[currGraphID] = mainUserSteamID
	graph.AddVertex(currGraphID)
	currGraphID++

	workerConfig.steamIDToGraphID[targetUserSteamID] = currGraphID
	workerConfig.graphIDToSteamID[currGraphID] = targetUserSteamID
	graph.AddVertex(currGraphID)
	currGraphID++

	firstJob := CrawlJob{
		FromID: mainUserSteamID,
		ToID:   mainUserSteamID,
	}
	jobsChan <- firstJob

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
			if exists := ifSteamIDSeenBefore(res.ToID, workerConfig.steamIDToGraphID); !exists {
				workerConfig.steamIDToGraphID[res.ToID] = currGraphID
				workerConfig.graphIDToSteamID[currGraphID] = res.ToID
				graph.AddVertex(workerConfig.steamIDToGraphID[res.ToID])
				currGraphID++
			}

			graph.AddArc(workerConfig.steamIDToGraphID[res.ToID], workerConfig.steamIDToGraphID[res.FromID], 1)
			graph.AddArc(workerConfig.steamIDToGraphID[res.FromID], workerConfig.steamIDToGraphID[res.ToID], 1)

			workerConfig.jobMutex.Lock()
			jobsChan <- CrawlJob{FromID: res.ToID}
			workerConfig.jobMutex.Unlock()

		default:
			// Just some filler
			temp := false
			if temp {
				temp = false
			}
		}
	}

	configuration.Logger.Info("waiting for all jobs to be done")
	wg.Wait()
	close(jobsChan)
	close(resChan)

	best, err := graph.Shortest(workerConfig.steamIDToGraphID[toInt64(userOne.UserDetails.User.AccDetails.SteamID)], workerConfig.steamIDToGraphID[toInt64(userTwo.UserDetails.User.AccDetails.SteamID)])
	if err != nil {
		return false, []int64{}, util.MakeErr(err)
	}
	bestPathSteamIDs := []int64{}
	for _, graphID := range best.Path {
		bestPathSteamIDs = append(bestPathSteamIDs, workerConfig.graphIDToSteamID[graphID])
	}
	return true, bestPathSteamIDs, nil
}
