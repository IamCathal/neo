package worker

import (
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/neosteamfriendgraphing/common"
)

func InitWorkerConfig() datastructures.WorkerConfig {
	return datastructures.WorkerConfig{
		WorkerAmount: configuration.WorkerConfig.WorkerAmount,
	}
}

func StartUpWorkers(cntr controller.CntrInterface) {
	for i := 0; i < 40; i++ {
		go ControlFunc(cntr)
	}
}

func VerifyFormatOfSteamIDs(input datastructures.CrawlUsersInput) ([]string, error) {
	validSteamIDs := []string{}
	match, err := regexp.MatchString("([0-9]){17}", input.FirstSteamID)
	if err != nil {
		return validSteamIDs, err
	}
	if match {
		validSteamIDs = append(validSteamIDs, input.FirstSteamID)
	}

	match, err = regexp.MatchString("([0-9]){17}", input.SecondSteamID)
	if err != nil {
		return validSteamIDs, err
	}
	if match {
		validSteamIDs = append(validSteamIDs, input.SecondSteamID)
	}
	return validSteamIDs, nil
}

func putFriendsIntoQueue(cntr controller.CntrInterface, currentJob datastructures.Job, friendIDs []string) error {
	startTime := time.Now()

	nextLevel := currentJob.CurrentLevel + 1
	if nextLevel > currentJob.MaxLevel {
		// configuration.Logger.Info("not putting on friends")
		return nil
	}

	for _, ID := range friendIDs {
		newJob := datastructures.Job{
			JobType:               "crawl",
			OriginalTargetSteamID: currentJob.OriginalTargetSteamID,
			CurrentTargetSteamID:  ID,

			CrawlID:      currentJob.CrawlID,
			MaxLevel:     currentJob.MaxLevel,
			CurrentLevel: nextLevel,
		}

		configuration.Logger.Sugar().Infof("pushing job: %+v", newJob)
		err := publishJobWithThrottling(cntr, newJob)
		if err != nil {
			configuration.Logger.Sugar().Fatalf("failed to publish job after all retries: %+v", err)
			panic(err)
		}
	}

	configuration.Logger.Sugar().Infof("took %v to publish %d jobs to queue", time.Since(startTime), len(friendIDs))
	return nil
}

func getGamesOwned(cntr controller.CntrInterface, steamID string) ([]common.Game, error) {
	gamesInfo := []common.Game{}
	ownedGamesResponse, err := cntr.CallGetOwnedGames(steamID)
	if err != nil {
		return gamesInfo, err
	}
	// for _, game := range ownedGamesResponse.Games {

	// 	// Filter out some fields which will never be used
	// 	currentGame := common.GameInfoDocument{
	// 		AppID:           game.Appid,
	// 		Name:            game.Name,
	// 		PlaytimeForever: game.PlaytimeForever,
	// 		ImgIconURL:      game.ImgIconURL,
	// 		ImgLogoURL:      game.ImgLogoURL,
	// 	}

	// 	gamesInfo = append(gamesInfo, currentGame)
	// }
	return ownedGamesResponse.Games, nil
}

func getPlayerSummaries(cntr controller.CntrInterface, job datastructures.Job, friendIDs []string) ([]common.Player, error) {
	// Only 100 steamIDs can be queried per call
	stacksOfSteamIDs := breakIntoStacksOf100OrLessSteamIDs(friendIDs)

	allPlayerSummaries := []common.Player{}
	for i := 0; i < len(stacksOfSteamIDs); i++ {
		batchOfPlayerSummaries, err := cntr.CallGetPlayerSummaries(stacksOfSteamIDs[i])
		if err != nil {
			return []common.Player{}, err
		}
		allPlayerSummaries = append(allPlayerSummaries, batchOfPlayerSummaries...)
	}

	onlyPublicProfiles := getPublicProfiles(allPlayerSummaries)
	return onlyPublicProfiles, nil
}

// breakIntoStacksOf100OrLessSteaMIDs divides a list of steam IDs into stacks of one
// hundred IDs or less. The GetPlayerSummary API only accepts up to 100 steam IDs per
// call
// Returns lists of formatted IDs that the steam API will take like
// {"1233,4324,5435,5677,2432","34689,1035,7847,4673,9384"}
func breakIntoStacksOf100OrLessSteamIDs(friendIDs []string) []string {
	totalFriendCount := len(friendIDs)
	steamIDList := []string{}

	fullOneHundredIDLists, remainder := divideAndGetRemainder(totalFriendCount, 100)
	// If less than one hundred total IDs
	if fullOneHundredIDLists == 0 {
		idList := []string{}
		for _, ID := range friendIDs {
			idList = append(idList, ID)
		}
		steamIDList = append(steamIDList, strings.Join(idList, ","))
		return steamIDList
	}

	// There are 100 or more total IDs
	for i := 0; i < fullOneHundredIDLists; i++ {
		// For each batch of 100 users
		idList := []string{}
		for j := 0; j < 100; j++ {
			idList = append(idList, friendIDs[j+(i*100)])
		}
		steamIDList = append(steamIDList, strings.Join(idList, ","))
	}
	if remainder == 0 {
		// There were a clean multiple of 100 steamIDs
		return steamIDList
	}

	// There was not a clean multiple of 100 steamIDs
	firstIndex := fullOneHundredIDLists * 100
	idList := []string{}
	for i := firstIndex; i < firstIndex+remainder; i++ {
		idList = append(idList, friendIDs[i])
	}
	steamIDList = append(steamIDList, strings.Join(idList, ","))

	return steamIDList
}

func divideAndGetRemainder(numerator, denominator int) (quotient, remainder int) {
	quotient = numerator / denominator
	remainder = numerator % denominator
	return
}

func extractSteamIDsFromPlayersList(friends []common.Player) []string {
	steamIDs := []string{}
	for _, friend := range friends {
		steamIDs = append(steamIDs, friend.Steamid)
	}
	return steamIDs
}

func extractSteamIDsfromFriendsList(friends common.Friendslist) []string {
	steamIDs := []string{}
	for _, friend := range friends.Friends {
		steamIDs = append(steamIDs, friend.Steamid)
	}
	return steamIDs
}

func getPublicProfiles(users []common.Player) []common.Player {
	publicProfiles := []common.Player{}
	for i := 0; i < len(users); i++ {
		// Visibility state of 1 or 2 means some level of privacy
		if users[i].Communityvisibilitystate == 3 {
			publicProfiles = append(publicProfiles, users[i])
		}
	}
	return publicProfiles
}

func getSteamIDsFromPlayers(users []common.Player) []string {
	steamIDs := []string{}
	for _, user := range users {
		steamIDs = append(steamIDs, user.Steamid)
	}
	return steamIDs
}

func getUsersProfileSummaryFromSlice(steamID string, playerSummaries []common.Player) (bool, common.Player) {
	for _, player := range playerSummaries {
		if player.Steamid == steamID {
			return true, player
		}
	}
	return false, common.Player{}
}

// getTopTwentyOrFewerGames gets the top twenty games ordered by playtime_forever.
// If there are less than twenty games then all of them are returned in sorted
// sorted order
func getTopTwentyOrFewerGames(allGames []common.Game) []common.Game {
	if len(allGames) == 0 {
		return []common.Game{}
	}
	gamesRankedByPlayTime := allGames
	sort.Slice(gamesRankedByPlayTime, func(i, j int) bool {
		return allGames[i].PlaytimeForever > allGames[j].PlaytimeForever
	})

	if len(allGames) >= 20 {
		return gamesRankedByPlayTime[:20]
	}

	return gamesRankedByPlayTime
}

func GetSlimmedDownOwnedGames(games []common.Game) []common.GameOwnedDocument {
	slimmedDownOwnedGames := []common.GameOwnedDocument{}
	for _, game := range games {
		currentGame := common.GameOwnedDocument{
			AppID:            game.Appid,
			Playtime_Forever: game.PlaytimeForever,
		}
		slimmedDownOwnedGames = append(slimmedDownOwnedGames, currentGame)
	}
	return slimmedDownOwnedGames
}

func GetSlimmedDownGames(games []common.Game) []common.GameInfoDocument {
	slimmedDownGames := []common.GameInfoDocument{}
	for _, game := range games {
		currentGame := common.GameInfoDocument{
			AppID:      game.Appid,
			Name:       game.Name,
			ImgIconURL: game.ImgIconURL,
			ImgLogoURL: game.ImgLogoURL,
		}
		slimmedDownGames = append(slimmedDownGames, currentGame)
	}
	return slimmedDownGames
}
