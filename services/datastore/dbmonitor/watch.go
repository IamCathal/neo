package dbmonitor

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/datastructures"
	"github.com/gorilla/websocket"
	"github.com/neosteamfriendgraphing/common"
	"github.com/segmentio/ksuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	NewUserStreamWebsockets []WebsocketConn
	NewUserStreamLock       sync.Mutex
	LastEightUserEvents     []datastructures.AddUserEvent

	CrawlingStatsStreamWebsockets []WebsocketConn
	CrawlingStatsStreamLock       sync.Mutex
)

type WebsocketConn struct {
	Ws      *websocket.Conn
	ID      string
	MatchOn string
}

func Monitor() {
	go watchNewUsers()
	go watchCrawlingStatusUpdates()
}

func watchNewUsers() {
	usersCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("USER_COLLECTION"))

	matchOnlyInsertsPipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "operationType", Value: "insert"},
			}},
		},
	}

	usersCollectionStream, err := usersCollection.Watch(context.TODO(), matchOnlyInsertsPipeline)
	if err != nil {
		configuration.Logger.Sugar().Fatalf("failed to init users collection stream: %+v", err)
		panic(err)
	}
	defer usersCollectionStream.Close(context.TODO())

	if err := configuration.DBClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		configuration.Logger.Fatal(fmt.Sprintf("unable to ping mongoDB: %v", err))
		panic(err)
	}

	configuration.Logger.Info("watching users collection")

	go emitRandomNewUsers()

	for usersCollectionStream.Next(context.TODO()) {
		var user common.UserDocument
		var event bson.M

		if err := usersCollectionStream.Decode(&event); err != nil {
			configuration.Logger.Sugar().Fatalf("failed to decode user from users collection stream: %+v", err)
			panic(err)
		}
		jsonEvent, err := json.Marshal(event["fullDocument"])
		if err != nil {
			configuration.Logger.Sugar().Fatalf("failed to marshal event from users collection stream: %+v", err)
			panic(err)
		}
		err = json.Unmarshal(jsonEvent, &user)
		if err != nil {
			configuration.Logger.Sugar().Fatalf("failed to unmarshal event from users collection stream: %+v", err)
			panic(err)
		}

		addUserEvent := datastructures.AddUserEvent{
			SteamID:     user.AccDetails.SteamID,
			PersonaName: user.AccDetails.Personaname,
			ProfileURL:  user.AccDetails.Profileurl,
			Avatar:      user.AccDetails.Avatar,
			CountryCode: user.AccDetails.Loccountrycode,
			CrawlTime:   time.Now().Unix(),
		}
		writeNewUserEventToAllWebsockets(addUserEvent)
	}
}
func emitRandomNewUsers() {
	profilers := []string{"https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/fb/fb9c36c36e54b8ca5f2e1cbd89c06574d1348af0.jpg",
		"https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/f3/f3d05db4d8557efbcdbfb337f4176abe9fcb5c1b.jpg",
		"https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/7d/7d6dd0cae5a80ae9ce7ec27c9173f20b5e5948f5.jpg",
		"https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/a4/a4e137a8dc641f3ed0234cd78fb4384961cae133.jpg",
		"https://steamcdn-a.akamaihd.net/steamcommunity/public/images/avatars/4b/4b9fa2d19b4fce90de72afe78b35e4b3225b89c7.jpg",
	}

	usernames := []string{"Ricky", "Bobandy", "Lahey", "Bubbles The Great", "The baby motel", "Barb", "Julian", "Trinity", "Ray", "Cyrus", "Sam", "Patrick Swayze", "Phil Collins"}
	i := 0
	for {
		time.Sleep(1500 * time.Millisecond)
		addUserEvent := datastructures.AddUserEvent{
			SteamID:     ksuid.New().String(),
			PersonaName: usernames[rand.Intn(len(profilers))],
			ProfileURL:  "https://cathaloc.dev",
			Avatar:      profilers[rand.Intn(len(profilers))],
			CountryCode: "IE",
			CrawlTime:   time.Now().Unix(),
		}
		writeNewUserEventToAllWebsockets(addUserEvent)
		i++
	}
}

func emitRandomCrawlingStats() {

	crawlID := "23PxM8mLuAoNW4SImvYXllgCVbZ"
	timeStarted := time.Now().Unix()
	fmt.Println(crawlID)
	i := 0
	for {
		time.Sleep(150 * time.Millisecond)
		crawlingStatUpdate := datastructures.CrawlingStatus{
			TimeStarted:       timeStarted,
			CrawlID:           crawlID,
			TotalUsersToCrawl: 150 + (i * 2),
			UsersCrawled:      89 + i,
		}
		writeCrawlingStatsUpdateToAllWebsockets(crawlingStatUpdate)
		i++
	}
}

func writeNewUserEventToAllWebsockets(event datastructures.AddUserEvent) error {
	addUserEventToMostRecent(event)
	websockets := GetNewUserStreamWebsocketConnections()

	jsonObj, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal addUserEvent for websocket serving: %+v", err)
	}

	for _, ws := range websockets {
		err := ws.Ws.WriteMessage(websocket.TextMessage, jsonObj)
		if err != nil {
			return fmt.Errorf("failed to write to websocket %s: %+v", ws.ID, err)
		}
	}
	return nil
}

func watchCrawlingStatusUpdates() {
	crawlingStatsCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("CRAWLING_STATS_COLLECTION"))

	matchOnlyUpdatesPipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "operationType", Value: "update"},
			}},
		},
	}
	crawlingStatsCollectionStream, err := crawlingStatsCollection.Watch(context.TODO(), matchOnlyUpdatesPipeline)
	if err != nil {
		configuration.Logger.Sugar().Fatalf("failed to init crawling stats collection stream: %+v", err)
		panic(err)
	}
	defer crawlingStatsCollectionStream.Close(context.TODO())

	if err := configuration.DBClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		configuration.Logger.Fatal(fmt.Sprintf("unable to ping mongoDB: %v", err))
		panic(err)
	}

	go emitRandomCrawlingStats()

	configuration.Logger.Info("watching crawling stats collection")

	for crawlingStatsCollectionStream.Next(context.TODO()) {
		var crawlingStat datastructures.CrawlingStatus
		var event bson.M

		if err := crawlingStatsCollectionStream.Decode(&event); err != nil {
			configuration.Logger.Sugar().Fatalf("failed to decode user from crawling stats collection stream: %+v", err)
			panic(err)
		}
		jsonEvent, err := json.Marshal(event["fullDocument"])
		if err != nil {
			configuration.Logger.Sugar().Fatalf("failed to marshal event from crawling stats collection stream: %+v", err)
			panic(err)
		}
		err = json.Unmarshal(jsonEvent, &crawlingStat)
		if err != nil {
			configuration.Logger.Sugar().Fatalf("failed to unmarshal event from crawling stats collection stream: %+v", err)
			panic(err)
		}

		writeCrawlingStatsUpdateToAllWebsockets(crawlingStat)
	}
}

func writeCrawlingStatsUpdateToAllWebsockets(crawlingStat datastructures.CrawlingStatus) error {
	websockets := GetCrawlingStatsStreamWebsocketConnections()

	jsonObj, err := json.Marshal(crawlingStat)
	if err != nil {
		return fmt.Errorf("failed to marshal crawlingStatUpdate for websocket serving: %+v", err)
	}

	for _, ws := range websockets {
		if crawlingStat.CrawlID == ws.MatchOn {
			// fmt.Printf("matched on, sending: %+v\n", string(jsonObj))
			err := ws.Ws.WriteMessage(websocket.TextMessage, jsonObj)
			if err != nil {
				return fmt.Errorf("failed to write to websocket %s: %+v", ws.ID, err)
			}
		}
	}
	return nil
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
	return []WebsocketConn{}, fmt.Errorf("failed to remove non existant websocket %s from ws connection list", websocketID)
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
