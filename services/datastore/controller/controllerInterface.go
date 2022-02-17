package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/IamCathal/neo/services/datastore/configuration"
	"github.com/IamCathal/neo/services/datastore/datastructures"
	"github.com/lib/pq"
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/util"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Cntr struct{}

type CntrInterface interface {
	// MongoDB related functions
	InsertOne(ctx context.Context, collection *mongo.Collection, bson []byte) (*mongo.InsertOneResult, error)
	UpdateCrawlingStatus(ctx context.Context, collection *mongo.Collection, crawlingStatus common.CrawlingStatus) (bool, error)
	GetUser(ctx context.Context, steamID string) (common.UserDocument, error)
	GetCrawlingStatusFromDBFromCrawlID(ctx context.Context, crawlID string) (common.CrawlingStatus, error)
	HasUserBeenCrawledBeforeAtLevel(ctx context.Context, level int, steamID string) (string, error)
	GetUsernames(ctx context.Context, steamIDs []string) (map[string]string, error)
	InsertGame(ctx context.Context, game common.BareGameInfo) (bool, error)
	GetDetailsForGames(ctx context.Context, IDList []int) ([]common.BareGameInfo, error)
	SaveShortestDistance(ctx context.Context, shortestDistanceInfo datastructures.ShortestDistanceInfo) (bool, error)
	GetShortestDistanceInfo(ctx context.Context, crawlIDs []string) (datastructures.ShortestDistanceInfo, error)
	// Postgresql related functions
	SaveProcessedGraphData(crawlID string, graphData common.UsersGraphData) (bool, error)
	GetProcessedGraphData(crawlID string) (common.UsersGraphData, error)
	DoesProcessedGraphDataExist(crawlID string) (bool, error)
}

func (control Cntr) InsertOne(ctx context.Context, collection *mongo.Collection, bson []byte) (*mongo.InsertOneResult, error) {
	insertionResult, err := collection.InsertOne(ctx, bson)
	if err != nil {
		// Sometimes duplicate users are inserted
		// This is not an issue
		if mongo.IsDuplicateKeyError(err) {
			return insertionResult, nil
		}
		return nil, util.MakeErr(err, "failed to insert document")
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
		return false, util.MakeErr(updatedDoc.Err())
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
		if err == mongo.ErrNoDocuments {
			return common.UserDocument{}, nil
		}
		return common.UserDocument{}, util.MakeErr(err)
	}
	return userDoc, nil
}

func (control Cntr) GetCrawlingStatusFromDBFromCrawlID(ctx context.Context, crawlID string) (common.CrawlingStatus, error) {
	crawlingStatsCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("CRAWLING_STATS_COLLECTION"))
	crawlingStatus := common.CrawlingStatus{}

	if err := crawlingStatsCollection.FindOne(ctx, bson.M{
		"crawlid": crawlID,
	}).Decode(&crawlingStatus); err != nil {
		if err == mongo.ErrNoDocuments {
			return common.CrawlingStatus{}, nil
		}
		return common.CrawlingStatus{}, util.MakeErr(err)
	}
	return crawlingStatus, nil
}

func (control Cntr) HasUserBeenCrawledBeforeAtLevel(ctx context.Context, level int, steamID string) (string, error) {
	crawlingStatsCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("CRAWLING_STATS_COLLECTION"))
	crawlStatus := common.CrawlingStatus{}

	projection := bson.D{
		{Key: "crawlid", Value: 1},
	}

	err := crawlingStatsCollection.FindOne(ctx, bson.M{
		"maxlevel":            level,
		"originalcrawltarget": steamID,
	}, options.FindOne().SetProjection(projection)).Decode(&crawlStatus)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", nil
		}
		return "", util.MakeErr(err, "failed to find existing crawlingstatus")
	}

	return crawlStatus.CrawlID, nil
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
		if err == mongo.ErrNoDocuments {
			return make(map[string]string), nil
		}
		return make(map[string]string), util.MakeErr(err)
	}
	defer cursor.Close(ctx)

	var allUsers []common.UserDocument
	var singleUser common.UserDocument
	for cursor.Next(ctx) {
		err = cursor.Decode(&singleUser)
		if err != nil {
			return make(map[string]string), util.MakeErr(err)
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

func (control Cntr) InsertGame(ctx context.Context, game common.BareGameInfo) (bool, error) {
	gamesCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection("games")

	bsonObj, err := bson.Marshal(game)
	if err != nil {
		return false, util.MakeErr(err)
	}

	_, err = gamesCollection.InsertOne(ctx, bsonObj)
	if err != nil {
		// Sometimes duplicate users are inserted
		// This is not an issue
		if mongo.IsDuplicateKeyError(err) {
			return true, nil
		}
		return false, util.MakeErr(err)
	}
	return true, nil
}

func (control Cntr) GetDetailsForGames(ctx context.Context, IDList []int) ([]common.BareGameInfo, error) {
	gamesCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection("games")

	configuration.Logger.Sugar().Infof("searching for details for the following %d games: %+v", len(IDList), IDList)
	cursor, err := gamesCollection.Find(ctx,
		bson.D{{Key: "appid", Value: bson.D{{Key: "$in", Value: IDList}}}})
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return []common.BareGameInfo{}, nil
		}
		return []common.BareGameInfo{}, util.MakeErr(err)
	}
	defer cursor.Close(ctx)

	var allGames []common.BareGameInfo
	var singleGame common.BareGameInfo
	for cursor.Next(ctx) {
		err = cursor.Decode(&singleGame)
		if err != nil {
			return []common.BareGameInfo{}, util.MakeErr(err)
		}
		allGames = append(allGames, singleGame)
	}
	return allGames, nil
}

func (control Cntr) SaveShortestDistance(ctx context.Context, shortestDistanceInfo datastructures.ShortestDistanceInfo) (bool, error) {
	shortestDistanceCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("SHORTEST_DISTANCE_COLLECTION"))
	bsonObj, err := bson.Marshal(shortestDistanceInfo)
	if err != nil {
		return false, util.MakeErr(err)
	}
	_, err = control.InsertOne(context.TODO(), shortestDistanceCollection, bsonObj)
	if err != nil {
		return false, util.MakeErr(err)
	}
	return true, nil
}

func (control Cntr) GetShortestDistanceInfo(ctx context.Context, crawlIDs []string) (datastructures.ShortestDistanceInfo, error) {
	userCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("SHORTEST_DISTANCE_COLLECTION"))
	shortestDistanceInfo := datastructures.ShortestDistanceInfo{}

	if err := userCollection.FindOne(ctx,
		bson.D{{Key: "crawlids", Value: bson.D{{Key: "$all", Value: crawlIDs}}}}).Decode(&shortestDistanceInfo); err != nil {
		if err == mongo.ErrNoDocuments {
			return datastructures.ShortestDistanceInfo{}, nil
		}
		return datastructures.ShortestDistanceInfo{}, util.MakeErr(err)
	}
	return shortestDistanceInfo, nil
}

func (control Cntr) SaveProcessedGraphData(crawlID string, graphData common.UsersGraphData) (bool, error) {
	jsonBody, err := json.Marshal(graphData)
	if err != nil {
		return false, util.MakeErr(err, "failed to unmarshal graphdata json")
	}

	queryString := `INSERT INTO graphdata (crawlid, graphdata) VALUES ($1, $2)`
	_, err = configuration.SQLClient.Exec(queryString, crawlID, string(jsonBody))
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// Attempts to duplicate an insert are not an issue
			if err.Code.Name() != "unique_violation" {
				return false, util.MakeErr(err, "failed to exec insert into graphdata")
			} else {
				configuration.Logger.Sugar().Infof("duplicate insert processed data for crawlid %s was attempted", crawlID)
			}
		}
	}
	return true, nil
}

func (control Cntr) GetProcessedGraphData(crawlID string) (common.UsersGraphData, error) {
	graphData := common.UsersGraphData{}

	queryString := `SELECT * FROM graphdata WHERE crawlid = $1`
	res, err := configuration.SQLClient.Query(queryString, crawlID)
	if err != nil {
		return common.UsersGraphData{}, util.MakeErr(err)
	}
	graphDataJSON := ""
	for res.Next() {
		crawlID := ""
		if err := res.Scan(&crawlID, &graphDataJSON); err != nil {
			return common.UsersGraphData{}, util.MakeErr(err, "failed to scan returned row")
		}
	}
	if len(graphDataJSON) == 0 {
		return common.UsersGraphData{}, nil
	}
	err = json.Unmarshal([]byte(graphDataJSON), &graphData)
	if err != nil {
		return common.UsersGraphData{}, fmt.Errorf("failed to unmarshal returned data for crawlid %s: %+v", crawlID, err)
	}
	return graphData, nil
}

func (control Cntr) DoesProcessedGraphDataExist(crawlID string) (bool, error) {
	queryString := `SELECT crawlid FROM graphdata WHERE crawlid = $1`
	res, err := configuration.SQLClient.Query(queryString, crawlID)
	if err != nil {
		return false, util.MakeErr(err)
	}
	crawlIDFromRow := ""
	for res.Next() {
		if err := res.Scan(&crawlIDFromRow); err != nil {
			return false, util.MakeErr(err, "failed to scan returned row")
		}
	}
	if len(crawlIDFromRow) == 0 {
		return false, nil
	}
	return true, nil
}
