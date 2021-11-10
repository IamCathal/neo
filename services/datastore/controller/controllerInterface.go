package controller

import (
	"context"

	"github.com/neosteamfriendgraphing/common/dtos"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Cntr struct{}

type CntrInterface interface {
	InsertOne(ctx context.Context, collection *mongo.Collection, bson []byte) (*mongo.InsertOneResult, error)
	UpdateCrawlingStatus(ctx context.Context, collection *mongo.Collection, saveUserDTO dtos.SaveUserDTO, moreUsersToCrawl, usersCrawled int) (bool, error)
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
