package controller

import (
	"context"
	"os"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/dtos"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Cntr struct{}

type CntrInterface interface {
	InsertOne(ctx context.Context, collection *mongo.Collection, bson []byte) (*mongo.InsertOneResult, error)
	UpdateCrawlingStatus(ctx context.Context, collection *mongo.Collection, saveUserDTO dtos.SaveUserDTO, moreUsersToCrawl, usersCrawled int) (bool, error)
	GetUser(ctx context.Context, steamID string) (common.UserDocument, error)
}

func (control Cntr) InsertOne(ctx context.Context, collection *mongo.Collection, bson []byte) (*mongo.InsertOneResult, error) {
	insertionResult, err := collection.InsertOne(ctx, bson)
	if err != nil {
		return nil, err
	}
	return insertionResult, nil
}

func (control Cntr) UpdateCrawlingStatus(ctx context.Context, collection *mongo.Collection, saveUserDTO dtos.SaveUserDTO, moreUsersToCrawl, usersCrawled int) (bool, error) {
	updatedDoc := collection.FindOneAndUpdate(context.TODO(),
		bson.M{"originalcrawltarget": saveUserDTO.OriginalCrawlTarget},
		bson.D{
			{
				"$inc",
				bson.D{
					{"totaluserstocrawl", moreUsersToCrawl},
					{"userscrawled", usersCrawled},
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
		"steamid": steamID,
	}).Decode(&userDoc); err != nil {
		return common.UserDocument{}, err
	}
	return userDoc, nil
}
