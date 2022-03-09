package dbmonitor

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/controller"
	"github.com/IamCathal/neo/services/datastore/datastructures"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/util"
	"github.com/segmentio/ksuid"
)

func emitRandomNewUsers() {
	profilers := []string{"https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/fb/fb9c36c36e54b8ca5f2e1cbd89c06574d1348af0.jpg",
		"https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/f3/f3d05db4d8557efbcdbfb337f4176abe9fcb5c1b.jpg",
		"https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/7d/7d6dd0cae5a80ae9ce7ec27c9173f20b5e5948f5.jpg",
		"https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/a4/a4e137a8dc641f3ed0234cd78fb4384961cae133.jpg",
		"https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/4b/4b9fa2d19b4fce90de72afe78b35e4b3225b89c7.jpg",
	}
	randomCountries := []string{"IE", "DE", "SK", "IT", "CN", "AU", "CA", "MX", "GL", "AR", "BR", "PK"}
	usernames := []string{"Ricky", "Bobandy", "Lahey", "Bubbles The Great", "The baby motel", "Barb", "Julian", "Trinity", "Ray", "Cyrus", "Sam", "Patrick Swayze", "Phil Collins"}
	i := 0
	for {
		time.Sleep(1500 * time.Millisecond)
		addUserEvent := datastructures.AddUserEvent{
			SteamID:     ksuid.New().String(),
			PersonaName: usernames[rand.Intn(len(profilers))],
			ProfileURL:  "https://cathaloc.dev",
			Avatar:      profilers[rand.Intn(len(profilers))],
			CountryCode: randomCountries[rand.Intn(len(randomCountries))],
			CrawlTime:   time.Now().Unix(),
		}
		writeNewUserEventToAllWebsockets(addUserEvent)
		i++
	}
}

func GetNewUserStreamWebsocketConnections() []WebsocketConn {
	return NewUserStreamWebsockets
}
func SetNewUserStreamWebsocketConnections(connections []WebsocketConn) {
	NewUserStreamWebsockets = connections
}

func GetCrawlingStatsStreamWebsocketConnections() []WebsocketConn {
	return CrawlingStatsStreamWebsockets
}
func SetCrawlingStatsStreamWebsocketConnections(connections []WebsocketConn) {
	CrawlingStatsStreamWebsockets = connections
}

func AddNewStreamWebsocketConnection(conn WebsocketConn, connections []WebsocketConn, lock *sync.Mutex) []WebsocketConn {
	lock.Lock()
	connections = append(connections, conn)
	configuration.Logger.Sugar().Infof("adding websocket connection %+v to websocket connections", conn)
	lock.Unlock()
	return connections
}

func RemoveAWebsocketConnection(websocketID string, connections []WebsocketConn, lock *sync.Mutex) ([]WebsocketConn, error) {
	lock.Lock()
	websocketFound := false
	for i, currWebsock := range connections {
		if currWebsock.ID == websocketID {
			websocketFound = true
			connections[i] = connections[len(connections)-1]
			connections = connections[:len(connections)-1]
			lock.Unlock()
			configuration.Logger.Sugar().Infof("removing websocket connection %+v from websocket connections", currWebsock)
		}
	}
	if websocketFound {
		return connections, nil
	}
	return []WebsocketConn{}, util.MakeErr(errors.New("tried to remove non existant websocket"), fmt.Sprintf("failed to remove non existant websocket %s from ws connection list", websocketID))
}

func addUserEventToMostRecent(event datastructures.AddUserEvent) {
	mostRecentEvents := reverseEvents(LastEightUserEvents)
	mostRecentEvents = append(mostRecentEvents, event)
	mostRecentEvents = reverseEvents(mostRecentEvents)
	if len(mostRecentEvents) >= 8 {
		LastEightUserEvents = mostRecentEvents[:8]
		return
	}
	LastEightUserEvents = mostRecentEvents
}

func reverseEvents(list []datastructures.AddUserEvent) []datastructures.AddUserEvent {
	for i, j := 0, len(list)-1; i < j; i, j = i+1, j-1 {
		list[i], list[j] = list[j], list[i]
	}
	return list
}

func GetRecentFinishedCrawlsAfterTimestamp(timestamp int64) []datastructures.FinishedCrawlWithItsUser {
	appropriateCrawlingStatuses := []datastructures.FinishedCrawlWithItsUser{}

	finishedCrawlsLock.Lock()
	defer finishedCrawlsLock.Unlock()
	for _, crawlingStatus := range LastTwelveFinishedCrawls {
		if crawlingStatus.CrawlingStatus.TimeStarted > timestamp {
			appropriateCrawlingStatuses = append(appropriateCrawlingStatuses, crawlingStatus)
		}
	}
	return appropriateCrawlingStatuses
}

func GetAssociatedUsersForFinishedCrawls(cntr controller.CntrInterface, allCrawls []common.CrawlingStatus) ([]datastructures.FinishedCrawlWithItsUser, error) {
	allUsers := []common.UserDocument{}
	allCrawlsWithAssociatedUsers := []datastructures.FinishedCrawlWithItsUser{}
	var allUsersLock sync.Mutex
	var waitG sync.WaitGroup

	for i := 0; i < len(allCrawls); i++ {
		waitG.Add(1)
		newUser := common.UserDocument{}
		asyncGetUser(cntr, allCrawls[i].OriginalCrawlTarget, &newUser, &waitG, &allUsersLock)
		allUsersLock.Lock()
		allUsers = append(allUsers, newUser)
		allUsersLock.Unlock()
	}

	waitG.Wait()

	for i, user := range allUsers {
		allCrawlsWithAssociatedUsers = append(allCrawlsWithAssociatedUsers, datastructures.FinishedCrawlWithItsUser{
			CrawlingStatus: allCrawls[i],
			User:           user,
		})
	}

	return allCrawlsWithAssociatedUsers, nil
}

func GetRecentFinishedShortestDistanceCrawlsAfterTimestamp(timestamp int64) []datastructures.ShortestDistanceInfo {
	appropriateCrawlingStatuses := []datastructures.ShortestDistanceInfo{}

	finishedShortestDistanceCrawlLock.Lock()
	defer finishedShortestDistanceCrawlLock.Unlock()
	for _, shortestDistanceCrawlingStatus := range LastTwelveFinishedShortestDistanceCrawls {
		if shortestDistanceCrawlingStatus.TimeStarted > timestamp {
			appropriateCrawlingStatuses = append(appropriateCrawlingStatuses, shortestDistanceCrawlingStatus)
		}
	}
	return appropriateCrawlingStatuses
}

func asyncGetUser(cntr controller.CntrInterface, ID string, user *common.UserDocument, waitG *sync.WaitGroup, userLock *sync.Mutex) {
	defer waitG.Done()
	retrievedUser, err := cntr.GetUser(context.TODO(), ID)
	if err != nil {
		configuration.Logger.Panic(err.Error())
	}
	userLock.Lock()
	*user = retrievedUser
	userLock.Unlock()
}
