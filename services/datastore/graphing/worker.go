package graphing

import (
	"fmt"
	"sync"
	"time"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/controller"
	dijkstra "github.com/iamcathal/dijkstra2"
	"github.com/neosteamfriendgraphing/common"
)

type CrawlJob struct {
	FromID int64
	ToID   int64
}

type GraphWorkerConfig struct {
	mainUser   common.UsersGraphData
	targetUser common.UsersGraphData
	jobMutex   *sync.Mutex
	resMutex   *sync.Mutex

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
			// configuration.Logger.Sugar().Infof("[ID:%d][jobs:%d][res:%d] dijkstra worker received job: %+v",
			// 	id, len(jobs), len(res), currentJob)

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
					// fmt.Println("done sending")
					workerConfig.resMutex.Unlock()
				}
			}
			// fmt.Printf("done sending jobs for %v\n", toUser.AccDetails.SteamID)

			// time.Sleep(100 * time.Millisecond)
			workerConfig.usersCrawledMutex.Lock()
			workerConfig.UsersCrawled++
			// fmt.Printf("\n\n\t\t\tUser %v crawled: %d/%d\n\n\n", currentJob.FromID, workerConfig.UsersCrawled, workerConfig.TotalUsersToCrawl)
			workerConfig.usersCrawledMutex.Unlock()
			time.Sleep(120 * time.Millisecond)
		}
	}
}

func GetShortestPathIDs(cntr controller.CntrInterface, userOne, userTwo common.UsersGraphData) (bool, []int64, error) {
	maxChanLen := 100000
	jobsChan := make(chan CrawlJob, maxChanLen)
	resChan := make(chan CrawlJob, maxChanLen)

	workerConfig := GraphWorkerConfig{}
	// fmt.Printf("totalToCrawl: %v (%d + %d)\n", workerConfig.TotalUsersToCrawl, len(userOne.FriendDetails), len(userTwo.FriendDetails))
	var jobMutex sync.Mutex
	var resMutex sync.Mutex
	var wg sync.WaitGroup
	var usersCrawledMutex sync.Mutex
	workerConfig.jobMutex = &jobMutex
	workerConfig.resMutex = &resMutex
	workerConfig.usersCrawledMutex = &usersCrawledMutex
	workerConfig.UsersCrawled = 0

	workerConfig.steamIDToUser = GetIDToUserMap(userOne, userTwo)
	usersToCrawl := 0
	for _, _ = range workerConfig.steamIDToUser {
		usersToCrawl++
	}
	workerConfig.TotalUsersToCrawl = usersToCrawl
	mainUserSteamID := toInt64(userOne.UserDetails.User.AccDetails.SteamID)
	targetUserSteamID := toInt64(userTwo.UserDetails.User.AccDetails.SteamID)
	// 1548 unique people
	workerAmount := 2
	var stopSignal chan bool = make(chan bool, 0)
	workersAreDone := false

	for i := 0; i < workerAmount; i++ {
		wg.Add(1)
		go graphWorker(i, stopSignal, cntr, &wg, &workerConfig, jobsChan, resChan)
	}

	graph := dijkstra.NewGraph()

	// for i, val := range workerConfig.steamIDToUser {
	// 	fmt.Printf("%v %s\n", i, val.AccDetails.SteamID)
	// }

	workerConfig.graphIDToSteamID = make(map[int]int64)
	workerConfig.steamIDToGraphID = make(map[int64]int)
	currGraphID := 0

	workerConfig.steamIDToGraphID[mainUserSteamID] = currGraphID
	workerConfig.graphIDToSteamID[currGraphID] = mainUserSteamID
	graph.AddVertex(currGraphID)
	// fmt.Printf("graphID is %d\n", currGraphID)
	currGraphID++

	workerConfig.steamIDToGraphID[targetUserSteamID] = currGraphID
	workerConfig.graphIDToSteamID[currGraphID] = targetUserSteamID
	graph.AddVertex(currGraphID)
	// fmt.Printf("graphID is %d\n", currGraphID)
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
			// fmt.Println(workerConfig.UsersCrawled, workerConfig.TotalUsersToCrawl)
			// fmt.Println("exiting!")
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
				// fmt.Printf("not seen before: adding vertex for %v (%v)\n", res.ToID, workerConfig.steamIDToGraphID[res.ToID])
				graph.AddVertex(workerConfig.steamIDToGraphID[res.ToID])
				// fmt.Printf("graphID is %d\n", currGraphID)
				currGraphID++
			}
			// fmt.Printf("control: got %v (%v) -> %v (%v) ..... sending %v (%v)\n",
			// 	res.FromID, workerConfig.steamIDToGraphID[res.FromID],
			// 	res.ToID, workerConfig.steamIDToGraphID[res.ToID])

			// fmt.Printf("adding arcs for %v (%v) -> %v (%v)\n", res.FromID, workerConfig.steamIDToGraphID[res.FromID], res.ToID, workerConfig.steamIDToGraphID[res.ToID])
			// fmt.Printf("adding arcs for %v (%v) -> %v (%v)\n", res.ToID, workerConfig.steamIDToGraphID[res.ToID], res.FromID, workerConfig.steamIDToGraphID[res.FromID])
			graph.AddArc(workerConfig.steamIDToGraphID[res.ToID], workerConfig.steamIDToGraphID[res.FromID], 1)
			graph.AddArc(workerConfig.steamIDToGraphID[res.FromID], workerConfig.steamIDToGraphID[res.ToID], 1)

			workerConfig.jobMutex.Lock()
			jobsChan <- CrawlJob{FromID: res.ToID}
			workerConfig.jobMutex.Unlock()

			// fmt.Println("")
			// for i, vertex := range graph.Verticies {
			// 	fmt.Printf("\t%d - %v\n", i, vertex)
			// }
			// fmt.Println("")
		default:
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
		return false, []int64{}, err
	}
	bestPathSteamIDs := []int64{}
	for _, graphID := range best.Path {
		bestPathSteamIDs = append(bestPathSteamIDs, workerConfig.graphIDToSteamID[graphID])
	}
	return true, bestPathSteamIDs, nil
}
