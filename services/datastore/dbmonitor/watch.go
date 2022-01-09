package dbmonitor

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/datastructures"
	"github.com/IamCathal/neo/services/datastore/endpoints"
	"github.com/gorilla/websocket"
	"github.com/neosteamfriendgraphing/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func Monitor() {
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
			PersonaName: user.AccDetails.Personaname,
			ProfileURL:  user.AccDetails.Profileurl,
			Avatar:      user.AccDetails.Avatar,
			CountryCode: user.AccDetails.Loccountrycode,
		}
		WriteNewUserEventToAllWebsockets(addUserEvent)
	}
}

func WriteNewUserEventToAllWebsockets(event datastructures.AddUserEvent) error {
	websockets := endpoints.GetNewUserStreamWebsocketConnections()

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
