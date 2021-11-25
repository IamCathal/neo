package app

import (
	"context"
	"os"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/controller"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/dtos"
	"github.com/pkg/errors"
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

func SaveCrawlingStatsToDB(cntr controller.CntrInterface, saveUserDTO dtos.SaveUserDTO) error {
	crawlingStatsCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("CRAWLING_STATS_COLLECTION"))

	if (saveUserDTO.CurrentLevel < saveUserDTO.MaxLevel) || (saveUserDTO.CurrentLevel == 1 && saveUserDTO.MaxLevel == 1) {
		// Increment the users crawled counter by one and add len(friends) to
		// totaluserstocrawl as they need to be crawled
		docExisted, err := cntr.UpdateCrawlingStatus(context.TODO(), crawlingStatsCollection, saveUserDTO, len(saveUserDTO.User.FriendIDs), 1)
		if err != nil {
			return err
		}

		if !docExisted {
			crawlingStats := common.CrawlingStatus{
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
	} else {
		// Increment the users crawled counter by one
		docExisted, err := cntr.UpdateCrawlingStatus(context.TODO(),
			crawlingStatsCollection,
			saveUserDTO,
			0, 1)
		if err != nil {
			return err
		}
		if !docExisted {
			return errors.Errorf("failed to increment userscrawled on last level for DTO: '%+v'", saveUserDTO.User.AccDetails.SteamID)
		}
	}

	configuration.Logger.Info("success on update crawling stats and user document")
	return nil
}

func GetUserFromDB(cntr controller.CntrInterface, steamID string) (common.UserDocument, error) {
	user, err := cntr.GetUser(context.TODO(), steamID)
	if err != nil {
		return common.UserDocument{}, err
	}
	return user, nil
}
