package app

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/controller"
	"github.com/IamCathal/neo/services/datastore/datastructures"
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

func SaveCrawlingStatsToDB(cntr controller.CntrInterface, currentLevel int, crawlingStatus datastructures.CrawlingStatus) error {
	crawlingStatsCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("CRAWLING_STATS_COLLECTION"))
	if (currentLevel < crawlingStatus.MaxLevel) || (currentLevel == 1 && crawlingStatus.MaxLevel == 1) {
		// Increment the users crawled counter by one and add len(friends) to
		// totaluserstocrawl as they need to be crawled
		crawlingStatus.TimeStarted = time.Now().Unix()
		docExisted, err := cntr.UpdateCrawlingStatus(context.TODO(), crawlingStatsCollection, crawlingStatus)
		if err != nil {
			return err
		}

		if !docExisted {
			crawlingStatus.TimeStarted = time.Now().Unix()
			crawlingStatus.UsersCrawled = 0
			if crawlingStatus.MaxLevel == 1 {
				crawlingStatus.UsersCrawled = 1
				crawlingStatus.TotalUsersToCrawl = 1
			}

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

func GetCrawlingStatsFromDBFromCrawlID(cntr controller.CntrInterface, crawlID string) (datastructures.CrawlingStatus, error) {
	crawlingStatus, err := cntr.GetCrawlingStatusFromDBFromCrawlID(context.TODO(), crawlID)
	if err != nil {
		return datastructures.CrawlingStatus{}, err
	}
	return crawlingStatus, nil
}

func IsCurrentlyBeingCrawled(cntr controller.CntrInterface, crawlID string) (bool, string, error) {
	crawlingStatus, err := cntr.GetCrawlingStatusFromDBFromCrawlID(context.TODO(), crawlID)
	if err != nil {
		return false, "", nil
	}
	if crawlingStatus.UsersCrawled < crawlingStatus.TotalUsersToCrawl {
		return true, crawlingStatus.OriginalCrawlTarget, nil
	}
	return false, "", nil
}
