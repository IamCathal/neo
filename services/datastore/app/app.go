package app

import (
	"context"
	"os"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/controller"
	"github.com/IamCathal/neo/services/datastore/datastructures"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/dtos"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func SaveUserToDB(cntr controller.CntrInterface, userDocument common.UserDocument) error {
	bsonObj, err := bson.Marshal(userDocument)
	if err != nil {
		return err
	}

	userCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("USER_COLLECTION"))
	_, err = cntr.InsertOne(context.TODO(), userCollection, bsonObj)
	return err
}

func SaveCrawlingStatsToDB(cntr controller.CntrInterface, saveUserDTO dtos.SaveUserDTO) error {
	crawlingStatsCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("CRAWLING_STATS_COLLECTION"))

	if saveUserDTO.CurrentLevel < saveUserDTO.MaxLevel {
		// Increment the users crawled counter by one and add len(friends) to
		// totaluserstocrawl as they need to be crawled
		updatedDoc := cntr.UpdateCrawlingStatus(context.TODO(), crawlingStatsCollection, saveUserDTO)

		if updatedDoc.Err() == mongo.ErrNoDocuments {
			crawlingStats := datastructures.CrawlingStatus{
				OriginalCrawlTarget: saveUserDTO.OriginalCrawlTarget,
				MaxLevel:            saveUserDTO.MaxLevel,
				TotalUsersToCrawl:   len(saveUserDTO.User.FriendIDs),
				UsersCrawled:        0,
			}
			bsonObj, err := bson.Marshal(crawlingStats)
			if err != nil {
				return err
			}

			_, err = cntr.InsertOne(context.TODO(), crawlingStatsCollection, bsonObj)
			if err != nil {
				return err
			}
			return nil
		}
		if updatedDoc.Err() != nil {
			return updatedDoc.Err()
		}
	} else {
		// Increment the users crawled counter by one
		updatedDoc := crawlingStatsCollection.FindOneAndUpdate(context.TODO(),
			bson.M{"originalcrawltarget": saveUserDTO.OriginalCrawlTarget},
			bson.D{
				{
					"$inc",
					bson.D{
						{"userscrawled", 1},
					},
				},
			})
		if updatedDoc.Err() == mongo.ErrNoDocuments {
			return errors.Errorf("failed to increment userscrawled on last level for DTO: '%+v'", saveUserDTO)
		}
	}

	configuration.Logger.Info("success on update crawling stats")
	return nil
}
