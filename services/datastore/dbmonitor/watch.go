package dbmonitor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/neosteamfriendgraphing/common"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func Monitor() {
	usersCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("USER_COLLECTION"))

	matchPipeline := mongo.Pipeline{
		bson.D{
			{Key: "$match", Value: bson.D{
				{Key: "operationType", Value: "insert"},
			}},
		},
	}

	usersCollectionStream, err := usersCollection.Watch(context.TODO(), matchPipeline)
	if err != nil {
		configuration.Logger.Sugar().Fatalf("failed to init users collection stream: %+v", err)
		panic(err)
	}
	defer usersCollectionStream.Close(context.TODO())

	if err := configuration.DBClient.Ping(context.TODO(), readpref.Primary()); err != nil {
		configuration.Logger.Fatal(fmt.Sprintf("unable to ping mongoDB: %v", err))
		log.Fatal(err)
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

		// newAddUserEvent := datastructures.AddUserEvent{
		// 	PersonaName: user.AccDetails.Personaname,
		// 	ProfileURL:  user.AccDetails.Profileurl,
		// 	Avatar:      user.AccDetails.Avatar,
		// }

		// ship to websockets

	}
}
