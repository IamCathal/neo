package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/controller"
	"github.com/neosteamfriendgraphing/common"
	"go.mongodb.org/mongo-driver/bson"
)

func SaveUserToDB(cntr controller.CntrInterface, userDocument common.UserDocument) error {
	gamesOwnedSlimmedDown := []common.GameOwnedDocument{}
	for _, game := range userDocument.GamesOwned {
		currentGame := common.GameOwnedDocument{
			AppID:            game.AppID,
			Playtime_Forever: game.Playtime_Forever,
		}
		gamesOwnedSlimmedDown = append(gamesOwnedSlimmedDown, currentGame)
	}

	UserDocument := common.UserDocument{
		AccDetails: userDocument.AccDetails,
		FriendIDs:  userDocument.FriendIDs,
		GamesOwned: gamesOwnedSlimmedDown,
	}

	bsonObj, err := bson.Marshal(UserDocument)
	if err != nil {
		return err
	}

	userCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("USER_COLLECTION"))
	_, err = cntr.InsertOne(context.TODO(), userCollection, bsonObj)
	return err
}

func SaveCrawlingStatsToDB(cntr controller.CntrInterface, currentLevel int, crawlingStatus common.CrawlingStatus) error {
	crawlingStatsCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("CRAWLING_STATS_COLLECTION"))
	if (currentLevel < crawlingStatus.MaxLevel) || (currentLevel == 1 && crawlingStatus.MaxLevel == 1) {
		// Increment the users crawled counter by one and add len(friends) to
		// totaluserstocrawl as they need to be crawled
		crawlingStatus.TimeStarted = time.Now()
		docExisted, err := cntr.UpdateCrawlingStatus(context.TODO(), crawlingStatsCollection, crawlingStatus)
		if err != nil {
			return err
		}

		if !docExisted {
			crawlingStatus.TimeStarted = time.Now()
			crawlingStatus.UsersCrawled = 0
			if crawlingStatus.MaxLevel == 1 {
				crawlingStatus.UsersCrawled = 1
				crawlingStatus.TotalUsersToCrawl = 1
			}
			// jsonObj, _ := json.Marshal(crawlingStats)
			// fmt.Println(string(jsonObj))
			bsonObj, err := bson.Marshal(crawlingStatus)
			if err != nil {
				return err
			}

			_, err = cntr.InsertOne(context.TODO(), crawlingStatsCollection, bsonObj)
			if err != nil {
				return err
			}
			return nil
		}
	} else {
		// Increment the users crawled counter by one but
		// do not increment users to crawl since we're at max level
		crawlingStatus.TotalUsersToCrawl = 0
		docExisted, err := cntr.UpdateCrawlingStatus(context.TODO(),
			crawlingStatsCollection,
			crawlingStatus)
		if err != nil {
			return err
		}
		if !docExisted {
			// For when the crawling status document has been deleted but
			// some jobs still remain in the queue that must be killed off
			warningMsg := fmt.Sprintf("crawlID '%s' originalcrawltarget '%s' has no crawling status entry", crawlingStatus.CrawlID, crawlingStatus.OriginalCrawlTarget)
			configuration.Logger.Warn(warningMsg)
			return nil
		}
	}

	configuration.Logger.Info("success on update crawling stats to db")
	return nil
}

func GetUserFromDB(cntr controller.CntrInterface, steamID string) (common.UserDocument, error) {
	user, err := cntr.GetUser(context.TODO(), steamID)
	if err != nil {
		return common.UserDocument{}, err
	}
	return user, nil
}
