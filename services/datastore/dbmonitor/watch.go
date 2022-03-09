package dbmonitor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/controller"
	"github.com/IamCathal/neo/services/datastore/datastructures"
	"github.com/gorilla/websocket"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/util"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	NewUserStreamWebsockets []WebsocketConn
	NewUserStreamLock       sync.Mutex
	LastEightUserEvents     []datastructures.AddUserEvent

	CrawlingStatsStreamWebsockets []WebsocketConn
	CrawlingStatsStreamLock       sync.Mutex

	LastTwelveFinishedCrawls                 []datastructures.FinishedCrawlWithItsUser
	finishedCrawlsLock                       sync.Mutex
	LastTwelveFinishedShortestDistanceCrawls []datastructures.ShortestDistanceInfo
	finishedShortestDistanceCrawlLock        sync.Mutex
)

type WebsocketConn struct {
	Ws      *websocket.Conn
	ID      string
	MatchOn string
}

func Monitor(cntr controller.CntrInterface) {
	go watchNewUsers()
	go watchCrawlingStatusUpdates()
	go watchRecentFinishedCrawls(cntr)
	go watchRecentFinishedShortestDistances(cntr)
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
		configuration.Logger.Sugar().Panicf("failed to init users collection stream: %+v", util.MakeErr(err))
	}
	defer usersCollectionStream.Close(context.TODO())

	if err := configuration.DBClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		configuration.Logger.Sugar().Panicf("unable to ping mongoDB: %v", util.MakeErr(err))
	}

	configuration.Logger.Info("watching users collection")

	go emitRandomNewUsers()

	for usersCollectionStream.Next(context.TODO()) {
		var user common.UserDocument
		var event bson.M

		if err := usersCollectionStream.Decode(&event); err != nil {
			configuration.Logger.Sugar().Panicf("failed to decode user from users collection stream: %+v", util.MakeErr(err))
		}
		jsonEvent, err := json.Marshal(event["fullDocument"])
		if err != nil {
			configuration.Logger.Sugar().Panicf("failed to marshal event from users collection stream: %+v", util.MakeErr(err))
		}
		err = json.Unmarshal(jsonEvent, &user)
		if err != nil {
			configuration.Logger.Sugar().Panicf("failed to unmarshal event from users collection stream: %+v", util.MakeErr(err))
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

func writeNewUserEventToAllWebsockets(event datastructures.AddUserEvent) error {
	addUserEventToMostRecent(event)
	websockets := GetNewUserStreamWebsocketConnections()

	jsonObj, err := json.Marshal(event)
	if err != nil {
		return util.MakeErr(err, "failed to marshal addUserEvent for websocket serving")
	}

	for _, ws := range websockets {
		err := ws.Ws.WriteMessage(websocket.TextMessage, jsonObj)
		if err != nil {
			return util.MakeErr(err, fmt.Sprintf("failed to write to websocket %s", ws.ID))
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
	crawlingStatsCollectionStream, err := crawlingStatsCollection.Watch(context.TODO(), matchOnlyUpdatesPipeline, options.ChangeStream().SetFullDocument(options.UpdateLookup))
	if err != nil {
		configuration.Logger.Sugar().Panicf("failed to init crawling stats collection stream: %+v", util.MakeErr(err))
	}
	defer crawlingStatsCollectionStream.Close(context.TODO())

	if err := configuration.DBClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		configuration.Logger.Sugar().Panicf("unable to ping mongoDB: %v", util.MakeErr(err))
	}

	// go emitRandomCrawlingStats()

	configuration.Logger.Info("watching crawling stats collection")

	for crawlingStatsCollectionStream.Next(context.TODO()) {
		var crawlingStat common.CrawlingStatus
		var event bson.M

		if err := crawlingStatsCollectionStream.Decode(&event); err != nil {
			configuration.Logger.Sugar().Panicf("failed to decode user from crawling stats collection stream: %+v", util.MakeErr(err))
		}
		jsonEvent, err := json.Marshal(event["fullDocument"])
		if err != nil {
			configuration.Logger.Sugar().Panicf("failed to marshal event from crawling stats collection stream: %+v", util.MakeErr(err))
		}
		err = json.Unmarshal(jsonEvent, &crawlingStat)
		if err != nil {
			configuration.Logger.Sugar().Panicf("failed to unmarshal event from users collection stream: %+v", util.MakeErr(err))
		}

		writeCrawlingStatsUpdateToAllWebsockets(crawlingStat)
	}
}

func writeCrawlingStatsUpdateToAllWebsockets(crawlingStat common.CrawlingStatus) error {
	websockets := GetCrawlingStatsStreamWebsocketConnections()

	jsonObj, err := json.Marshal(crawlingStat)
	if err != nil {
		return util.MakeErr(err, "failed to marshal crawlingStatUpdate for websocket serving")
	}

	for _, ws := range websockets {
		if crawlingStat.CrawlID == ws.MatchOn {
			err := ws.Ws.WriteMessage(websocket.TextMessage, jsonObj)
			if err != nil {
				return util.MakeErr(err, fmt.Sprintf("failed to write to websocket %s", ws.ID))
			}
		}
	}
	return nil
}

func watchRecentFinishedCrawls(cntr controller.CntrInterface) {
	numStatuses := int64(12)
	for {
		lastTwelveCrawls, err := cntr.GetNMostRecentFinishedCrawls(context.TODO(), numStatuses)
		if err != nil {
			configuration.Logger.Sugar().Panicf("failed to get %d most recent finished crawling statuses %+v", numStatuses, err)
		}
		lastTwelveCrawlsWithAssociatedUsers, err := GetAssociatedUsersForFinishedCrawls(cntr, lastTwelveCrawls)
		if err != nil {
			configuration.Logger.Sugar().Panicf("failed to associate %d most recent finished crawling statuses %+v", numStatuses, err)
		}
		finishedCrawlsLock.Lock()
		LastTwelveFinishedCrawls = lastTwelveCrawlsWithAssociatedUsers
		finishedCrawlsLock.Unlock()

		time.Sleep(30 * time.Second)
	}
}

func watchRecentFinishedShortestDistances(cntr controller.CntrInterface) {
	numStatuses := int64(12)
	for {
		lastTwelveShortestDistanceCrawls, err := cntr.GetNMostRecentFinishedShortestDistanceCrawls(context.TODO(), numStatuses)
		if err != nil {
			configuration.Logger.Sugar().Panicf("failed to get %d most recent finished shortest distance crawling statuses %+v", numStatuses, err)
		}
		finishedShortestDistanceCrawlLock.Lock()
		LastTwelveFinishedShortestDistanceCrawls = lastTwelveShortestDistanceCrawls
		finishedShortestDistanceCrawlLock.Unlock()

		time.Sleep(30 * time.Second)
	}
}
