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
	UpdateCrawlingStatus(ctx context.Context, collection *mongo.Collection, crawlingStatus datastructures.CrawlingStatus) (bool, error)
	GetUser(ctx context.Context, steamID string) (common.UserDocument, error)
	GetCrawlingStatusFromDBFromCrawlID(ctx context.Context, crawlID string) (datastructures.CrawlingStatus, error)
	HasUserBeenCrawledBeforeAtLevel(ctx context.Context, level int, steamID string) (string, error)
	GetUsernames(ctx context.Context, steamIDs []string) (map[string]string, error)
	InsertGame(ctx context.Context, game datastructures.BareGameInfo) (bool, error)
	GetDetailsForGames(ctx context.Context, IDList []int) ([]datastructures.BareGameInfo, error)
	// Postgresql related functions
	SaveProcessedGraphData(crawlID string, graphData datastructures.UsersGraphData) (bool, error)
	GetProcessedGraphData(crawlID string) (datastructures.UsersGraphData, error)
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
		return nil, fmt.Errorf("failed to insert document: %+v", err)
	}
	return insertionResult, nil
}

func (control Cntr) UpdateCrawlingStatus(ctx context.Context, collection *mongo.Collection, crawlingStatus datastructures.CrawlingStatus) (bool, error) {
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

func (control Cntr) GetCrawlingStatusFromDBFromCrawlID(ctx context.Context, crawlID string) (datastructures.CrawlingStatus, error) {
	crawlingStatsCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("CRAWLING_STATS_COLLECTION"))
	crawlingStatus := datastructures.CrawlingStatus{}

	if err := crawlingStatsCollection.FindOne(ctx, bson.M{
		"crawlid": crawlID,
	}).Decode(&crawlingStatus); err != nil {
		return datastructures.CrawlingStatus{}, err
	}

	return crawlingStatus, nil
}

func (control Cntr) HasUserBeenCrawledBeforeAtLevel(ctx context.Context, level int, steamID string) (string, error) {
	crawlingStatsCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection(os.Getenv("CRAWLING_STATS_COLLECTION"))
	crawlStatus := datastructures.CrawlingStatus{}

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
		return "", fmt.Errorf("failed to find existing crawlingstatus: %+v", err)
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

func (control Cntr) InsertGame(ctx context.Context, game datastructures.BareGameInfo) (bool, error) {
	gamesCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection("games")

	bsonObj, err := bson.Marshal(game)
	if err != nil {
		return false, err
	}

	_, err = gamesCollection.InsertOne(ctx, bsonObj)
	if err != nil {
		// Sometimes duplicate users are inserted
		// This is not an issue
		if mongo.IsDuplicateKeyError(err) {
			return true, nil
		}
		return false, err
	}
	return true, nil
}

func (control Cntr) GetDetailsForGames(ctx context.Context, IDList []int) ([]datastructures.BareGameInfo, error) {
	gamesCollection := configuration.DBClient.Database(os.Getenv("DB_NAME")).Collection("games")

	cursor, err := gamesCollection.Find(ctx,
		bson.D{{Key: "appid", Value: bson.D{{Key: "$in", Value: IDList}}}})
	if err != nil {
		return []datastructures.BareGameInfo{}, err
	}
	defer cursor.Close(ctx)

	var allGames []datastructures.BareGameInfo
	var singleGame datastructures.BareGameInfo
	for cursor.Next(ctx) {
		err = cursor.Decode(&singleGame)
		if err != nil {
			return []datastructures.BareGameInfo{}, err
		}
		allGames = append(allGames, singleGame)
	}

	return allGames, nil
}

func (control Cntr) SaveProcessedGraphData(crawlID string, graphData datastructures.UsersGraphData) (bool, error) {
	jsonBody, err := json.Marshal(graphData)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal graphdata json: %+v", err)
	}

	queryString := `INSERT INTO graphdata (crawlid, graphdata) VALUES ($1, $2)`
	_, err = configuration.SQLClient.Exec(queryString, crawlID, string(jsonBody))
	if err != nil {
		if err, ok := err.(*pq.Error); ok {
			// Attempts to duplicate an insert are not an issue
			if err.Code.Name() != "unique_violation" {
				return false, fmt.Errorf("failed to exec insert into graphdata: %+v", err)
			} else {
				configuration.Logger.Sugar().Infof("duplicate insert of crawlid %s was attempted", crawlID)
			}
		}
	}
	return true, nil
}

func (control Cntr) GetProcessedGraphData(crawlID string) (datastructures.UsersGraphData, error) {
	graphData := datastructures.UsersGraphData{}

	queryString := `SELECT * FROM graphdata WHERE crawlid = $1`
	res, err := configuration.SQLClient.Query(queryString, crawlID)
	if err != nil {
		return datastructures.UsersGraphData{}, err
	}
	graphDataJSON := ""
	for res.Next() {
		crawlID := ""
		if err := res.Scan(&crawlID, &graphDataJSON); err != nil {
			return datastructures.UsersGraphData{}, fmt.Errorf("failed to scan returned row: %+v", err)
		}
	}
	if len(graphDataJSON) == 0 {
		return datastructures.UsersGraphData{}, nil
	}
	err = json.Unmarshal([]byte(graphDataJSON), &graphData)
	if err != nil {
		return datastructures.UsersGraphData{}, fmt.Errorf("failed to unmarshal returned data for crawlid %s: %+v", crawlID, err)
	}
	return graphData, nil
}

func (control Cntr) DoesProcessedGraphDataExist(crawlID string) (bool, error) {
	queryString := `SELECT crawlid FROM graphdata WHERE crawlid = $1`
	res, err := configuration.SQLClient.Query(queryString, crawlID)
	if err != nil {
		return false, err
	}
	crawlIDFromRow := ""
	for res.Next() {
		crawlID := ""
		if err := res.Scan(&crawlID, &crawlIDFromRow); err != nil {
			return false, fmt.Errorf("failed to scan returned row: %+v", err)
		}
	}
	fmt.Println(len(crawlIDFromRow))
	if len(crawlIDFromRow) == 0 {
		return false, nil
	}
	return true, nil
}
