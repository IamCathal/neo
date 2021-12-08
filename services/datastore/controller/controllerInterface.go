package controller

import (
	"context"
	"os"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/neosteamfriendgraphing/common"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Cntr struct{}

type CntrInterface interface {
	InsertOne(ctx context.Context, collection *mongo.Collection, bson []byte) (*mongo.InsertOneResult, error)
	UpdateCrawlingStatus(ctx context.Context, collection *mongo.Collection, crawlingStatus common.CrawlingStatus) (bool, error)
	GetUser(ctx context.Context, steamID string) (common.UserDocument, error)
	GetCrawlingStatusFromDB(ctx context.Context, collection *mongo.Collection, crawlID string) (common.CrawlingStatus, error)
	GetUsernames(ctx context.Context, steamIDs []string) (map[string]string, error)
}

func (control Cntr) InsertOne(ctx context.Context, collection *mongo.Collection, bson []byte) (*mongo.InsertOneResult, error) {
	insertionResult, err := collection.InsertOne(ctx, bson)
	if err != nil {
		// Sometimes duplicate users are inserted
		// This is not an issue
		if mongo.IsDuplicateKeyError(err) {
			return insertionResult, nil
		}
		return nil, err
	}
	return insertionResult, nil
}

func (control Cntr) UpdateCrawlingStatus(ctx context.Context, collection *mongo.Collection, crawlingStatus common.CrawlingStatus) (bool, error) {
	updatedDoc := collection.FindOneAndUpdate(context.TODO(),
		bson.M{"crawlid": crawlingStatus.CrawlID},
		bson.D{
			primitive.E{
				Key: "$inc",
				Value: bson.D{
					primitive.E{Key: "totaluserstocrawl", Value: crawlingStatus.TotalUsersToCrawl},
					primitive.E{Key: "userscrawled", Value: 1},
				},
			},
		})
	// If the document did not exists
	if updatedDoc.Err() == mongo.ErrNoDocuments {
		return false, nil
	}
	// Document did exist but a different error was returned
	if updatedDoc.Err() != nil {
		return false, updatedDoc.Err()
	}
	// Document did exist (best case)
	return true, nil
}

func (control Cntr) GetUser(ctx context.Context, steamID string) (common.UserDocument, error) {
	userCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("USER_COLLECTION"))
	userDoc := common.UserDocument{}

	if err := userCollection.FindOne(ctx, bson.M{
		"accdetails.steamid": steamID,
	}).Decode(&userDoc); err != nil {
		return common.UserDocument{}, err
	}
	return userDoc, nil
}

func (control Cntr) GetCrawlingStatusFromDB(ctx context.Context, crawlingStatusCollection *mongo.Collection, crawlID string) (common.CrawlingStatus, error) {
	crawlingStatus := common.CrawlingStatus{}

	if err := crawlingStatusCollection.FindOne(ctx, bson.M{
		"crawlid": crawlID,
	}).Decode(&crawlingStatus); err != nil {
		return common.CrawlingStatus{}, err
	}

	return crawlingStatus, nil
}

func (control Cntr) GetUsernames(ctx context.Context, steamIDs []string) (map[string]string, error) {
	userCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("USER_COLLECTION"))
	steamIDToUsernameMap := make(map[string]string)

	projection := bson.D{
		{Key: "accdetails.steamid", Value: 1},
		{Key: "accdetails.personaname", Value: 1},
	}
	cursor, err := userCollection.Find(ctx,
		bson.D{{Key: "accdetails.steamid", Value: bson.D{{Key: "$in", Value: steamIDs}}}}, options.Find().SetProjection(projection))
	if err != nil {
		return make(map[string]string), err
	}
	defer cursor.Close(ctx)

	var allUsers []common.UserDocument
	var singleUser common.UserDocument
	for cursor.Next(ctx) {
		err = cursor.Decode(&singleUser)
		if err != nil {
			return make(map[string]string), err
		}
		allUsers = append(allUsers, singleUser)
	}

	if len(allUsers) != len(steamIDs) {
		return make(map[string]string), errors.Errorf("queried %d steamIDs and only received back %d", len(steamIDs), len(allUsers))
	}

	for _, result := range allUsers {
		steamIDToUsernameMap[result.AccDetails.SteamID] = result.AccDetails.Personaname
	}
	return steamIDToUsernameMap, nil
}
