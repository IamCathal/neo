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
	UpdateCrawlingStatus(ctx context.Context, collection *mongo.Collection, saveUserDTO dtos.SaveUserDTO) *mongo.SingleResult
}

func (control Cntr) InsertOne(ctx context.Context, collection *mongo.Collection, bson []byte) (*mongo.InsertOneResult, error) {
	insertionResult, err := collection.InsertOne(ctx, bson)
	if err != nil {
		return nil, err
	}
	return insertionResult, nil
}

func (control Cntr) UpdateCrawlingStatus(ctx context.Context, collection *mongo.Collection, saveUserDTO dtos.SaveUserDTO) *mongo.SingleResult {
	updatedDoc := collection.FindOneAndUpdate(context.TODO(),
		bson.M{"originalcrawltarget": saveUserDTO.OriginalCrawlTarget},
		bson.D{
			{
				"$inc",
				bson.D{
					{"totaluserstocrawl", len(saveUserDTO.User.FriendIDs)},
					{"userscrawled", 1},
				},
			},
		})
	return updatedDoc
}
