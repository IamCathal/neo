package worker

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/iamcathal/neo/services/crawler/configuration"
	"github.com/iamcathal/neo/services/crawler/controller"
	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/neosteamfriendgraphing/common"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

var (
	testUser common.UserDocument
)

func TestMain(m *testing.M) {
	initTestData()
	c := zap.NewProductionConfig()
	c.OutputPaths = []string{"/dev/null"}
	log, err := c.Build()
	if err != nil {
		panic(err)
	}
	configuration.Logger = log

	code := m.Run()

	os.Exit(code)
}

func initTestData() {
	testUser = common.UserDocument{
		AccDetails: common.AccDetailsDocument{
			SteamID:        "76561197969081524",
			Personaname:    "persona name",
			Profileurl:     "profile url",
			Avatar:         "avatar url",
			Timecreated:    1223525546,
			Loccountrycode: "IE",
		},
		FriendIDs: []string{"1234", "5678"},
		GamesOwned: []common.GameOwnedDocument{
			{
				AppID:            102,
				Playtime_Forever: 1337,
			},
		},
	}
}

func TestPutFriendsIntoJobsQueue(t *testing.T) {
	mockController := &controller.MockCntrInterface{}
	currentJob := datastructures.Job{
		JobType:               "crawl",
		OriginalTargetSteamID: "12345",
		CurrentTargetSteamID:  "12345",
		CrawlID:               "2345345346546sdfdfbhfd",
		MaxLevel:              2,
		CurrentLevel:          1,
	}
	friendIDs := []string{"12455", "29456", "05838", "54954", "45967"}

	mockController.On("PublishToJobsQueue", mock.Anything).Return(nil)

	err := putFriendsIntoQueue(mockController, currentJob, friendIDs)

	assert.Nil(t, err)
	mockController.AssertNumberOfCalls(t, "PublishToJobsQueue", len(friendIDs))
}

func TestGetOwnedGamesReturnsAValidResponse(t *testing.T) {
	mockController := &controller.MockCntrInterface{}
	gameID := 123
	gameIconHash := "exampleHash"
	gameLogoHash := "anotherExampleHash"

	testResponse := common.GamesOwnedResponse{
		GameCount: 1,
		Games: []common.Game{
			{
				Appid:           gameID,
				Name:            "CS:GO",
				PlaytimeForever: 1377,
				Playtime2Weeks:  15,
				ImgIconURL:      gameIconHash,
				ImgLogoURL:      gameLogoHash,
			},
		},
	}
	mockController.On("CallGetOwnedGames", mock.AnythingOfType("string")).Return(testResponse, nil)

	gamesOwnedForCurrentUser, err := getGamesOwned(mockController, "exampleSteamID")

	assert.Nil(t, err)

	assert.Len(t, gamesOwnedForCurrentUser, 1)
	assert.Equal(t, gameIconHash, gamesOwnedForCurrentUser[0].ImgIconURL)
	assert.Equal(t, gameLogoHash, gamesOwnedForCurrentUser[0].ImgLogoURL)
}

func TestGetOwnedGamesEmptyWhenNoGamesFound(t *testing.T) {
	mockController := &controller.MockCntrInterface{}

	testResponse := common.GamesOwnedResponse{}

	mockController.On("CallGetOwnedGames", mock.AnythingOfType("string")).Return(testResponse, nil)

	gamesOwnedForCurrentUser, err := getGamesOwned(mockController, "exampleSteamID")

	assert.Nil(t, err)
	assert.Len(t, gamesOwnedForCurrentUser, 0)
}

func TestGetOwnedGamesAnErrorWhenAPIThrowsOne(t *testing.T) {
	mockController := &controller.MockCntrInterface{}

	testErrorMsg := "all your base are belong to us"
	testError := errors.New(testErrorMsg)

	mockController.On("CallGetOwnedGames", mock.AnythingOfType("string")).Return(common.GamesOwnedResponse{}, testError)

	gamesOwnedForCurrentUser, err := getGamesOwned(mockController, "exampleSteamID")

	assert.ErrorIs(t, testError, err)
	assert.Len(t, gamesOwnedForCurrentUser, 0)
}
func TestVerifyFormatOfSteamIDsVerifiesTwoValidSteamIDs(t *testing.T) {
	expectedSteamIDs := []string{"12345678901234456", "72348978301996243"}
	inputData := datastructures.CrawlUsersInput{
		FirstSteamID:  expectedSteamIDs[0],
		SecondSteamID: expectedSteamIDs[1],
	}

	receivedValidSteamIDs, err := VerifyFormatOfSteamIDs(inputData)

	assert.Nil(t, err)
	assert.Equal(t, expectedSteamIDs, receivedValidSteamIDs, "expect two valid format steamIDs are returned")
}

func TestVerifyFormatOfSteamIDsReturnsNothingForTwoInvalidFormatSteamIDs(t *testing.T) {
	expectedSteamIDs := []string{"12345634456", "0"}
	inputData := datastructures.CrawlUsersInput{
		FirstSteamID:  expectedSteamIDs[0],
		SecondSteamID: expectedSteamIDs[1],
	}

	receivedValidSteamIDs, err := VerifyFormatOfSteamIDs(inputData)

	assert.Nil(t, err)
	assert.Len(t, receivedValidSteamIDs, 0, "expect to receive back no steamIDs for two invalid steamID inputs")
}

func TestExctractSteamIDsfromFriendsList(t *testing.T) {
	expectedList := []string{"1234", "5436", "6718"}
	friends := common.Friendslist{
		Friends: []common.Friend{
			{
				Steamid: "1234",
			},
			{
				Steamid: "5436",
			},
			{
				Steamid: "6718",
			},
		},
	}

	realList := extractSteamIDsfromFriendsList(friends)

	assert.Equal(t, expectedList, realList)
}

func TestBreakSteamIDsIntoListsOf100OrLessWith100IDs(t *testing.T) {
	idList := []string{}
	for i := 0; i < 100; i++ {
		idList = append(idList, strconv.Itoa(i))
	}
	URLFormattedSteamIDs := strings.Join(idList, ",")
	expectedSteamIDList := []string{URLFormattedSteamIDs}

	realSteamIDList := breakIntoStacksOf100OrLessSteamIDs(idList)

	assert.Equal(t, expectedSteamIDList, realSteamIDList)
}

func TestBreakSteamIDsIntoListsOf100OrLessWith120IDs(t *testing.T) {
	idList := []string{}
	for i := 0; i < 120; i++ {
		idList = append(idList, strconv.Itoa(i))
	}
	firstBatchOfURLFormattedSteamIDs := strings.Join(idList[:100], ",")
	remainderBatchOfURLFormattedSteamIDs := strings.Join(idList[100:], ",")

	expectedSteamIDList := []string{firstBatchOfURLFormattedSteamIDs, remainderBatchOfURLFormattedSteamIDs}

	realSteamIDList := breakIntoStacksOf100OrLessSteamIDs(idList)

	assert.Equal(t, expectedSteamIDList, realSteamIDList)
}

func TestBreakSteamIDsIntoListsOf100OrLess20IDs(t *testing.T) {
	idList := []string{}
	for i := 0; i < 20; i++ {
		idList = append(idList, strconv.Itoa(i))
	}
	URLFormattedSteamIDs := strings.Join(idList, ",")
	expectedSteamIDList := []string{URLFormattedSteamIDs}

	realSteamIDList := breakIntoStacksOf100OrLessSteamIDs(idList)

	assert.Equal(t, expectedSteamIDList, realSteamIDList)
}

func TestBreakSteamIDsIntoListsOf100OrLessWith1911IDs(t *testing.T) {
	idList := []string{}
	for i := 0; i < 1911; i++ {
		idList = append(idList, strconv.Itoa(i))
	}

	realSteamIDList := breakIntoStacksOf100OrLessSteamIDs(idList)

	assert.Len(t, realSteamIDList, 20)
}

func TestGetUsersProfileSummaryFromSliceReturnsTheSearchedForProfile(t *testing.T) {
	expectedUserProfile := common.Player{
		Steamid:  "54290543656",
		Realname: "Eddie Durcan",
	}
	exampleSummaries := []common.Player{
		{
			Steamid:  "213023525435",
			Realname: "Buzz Mc Donell",
		},
		expectedUserProfile,
		{
			Steamid:  "5647568578975",
			Realname: "The Boogenhagen",
		},
	}

	found, realUserProfile := getUsersProfileSummaryFromSlice(expectedUserProfile.Steamid, exampleSummaries)

	assert.True(t, found)
	assert.Equal(t, expectedUserProfile, realUserProfile)
}

func TestGetUsersProfileSummaryFromSliceReturnsFalseWhenNotFound(t *testing.T) {
	nonExistantSteamID := "45356346547567"
	exampleSummaries := []common.Player{
		{
			Steamid:  "213023525435",
			Realname: "Buzz Mc Donell",
		},
		{
			Steamid:  "5647568578975",
			Realname: "The Boogenhagen",
		},
	}

	found, realUserProfile := getUsersProfileSummaryFromSlice(nonExistantSteamID, exampleSummaries)

	assert.False(t, found)
	assert.Empty(t, realUserProfile)
}

func TestGetSteamIDsFromPlayersReturnsAllSteamIDs(t *testing.T) {
	examplePlayers := []common.Player{
		{
			Steamid:  "213023525435",
			Realname: "Buzz Mc Donell",
		},
		{
			Steamid:  "54290543656",
			Realname: "Eddie Durcan",
		},
		{
			Steamid:  "5647568578975",
			Realname: "The Boogenhagen",
		},
	}
	expectedSteamIDList := []string{examplePlayers[0].Steamid, examplePlayers[1].Steamid, examplePlayers[2].Steamid}

	realSteamIDs := getSteamIDsFromPlayers(examplePlayers)

	assert.Equal(t, expectedSteamIDList, realSteamIDs)
}

func TestGetSteamIDsFromPlayersFromAnEmptySliceReturnsNothing(t *testing.T) {
	examplePlayers := []common.Player{}

	realSteamIDs := getSteamIDsFromPlayers(examplePlayers)

	assert.Empty(t, realSteamIDs)
}

func TestGetPublicProfilesReturnsOnlyPublicProfiles(t *testing.T) {
	expectedPublicProfile := common.Player{
		Steamid:                  "54290543656",
		Realname:                 "Eddie Durcan",
		Communityvisibilitystate: 3,
	}
	examplePlayers := []common.Player{
		{
			Steamid:                  "213023525435",
			Realname:                 "Buzz Mc Donell",
			Communityvisibilitystate: 2,
		},
		expectedPublicProfile,
		{
			Steamid:                  "5647568578975",
			Realname:                 "The Boogenhagen",
			Communityvisibilitystate: 1,
		},
	}

	realPublicProfiles := getPublicProfiles(examplePlayers)

	assert.Equal(t, expectedPublicProfile, realPublicProfiles[0])
}

func TestInitWorkerConfig(t *testing.T) {
	expectedWorkerAmount := 20
	configuration.WorkerConfig.WorkerAmount = expectedWorkerAmount

	workerConfig := InitWorkerConfig()

	assert.Equal(t, expectedWorkerAmount, workerConfig.WorkerAmount)
}

func TestCrawlUser(t *testing.T) {
	mockController := controller.MockCntrInterface{}
	mockController.On("PublishToJobsQueue", mock.Anything).Return(nil)

	CrawlUser(&mockController, "testSteamID", "testcrawlID", 4)
}

func TestCrawlUserWhenErrorIsReturnedPublishingJobToQueue(t *testing.T) {
	mockController := controller.MockCntrInterface{}
	mockController.On("PublishToJobsQueue", mock.Anything).Return(errors.New("test error"))

	CrawlUser(&mockController, "testSteamID", "testcrawlID", 4)
}

func TestGetFriendsWhenFriendIsFoundFromDatastore(t *testing.T) {
	mockController := controller.MockCntrInterface{}
	mockController.On("GetUserFromDataStore", mock.AnythingOfType("string")).Return(testUser, nil)

	didExistInDatastore, friends, err := GetFriends(&mockController, testUser.AccDetails.SteamID)

	mockController.AssertNotCalled(t, "CallGetFriends")

	assert.True(t, didExistInDatastore)
	assert.Equal(t, testUser.FriendIDs, friends)
	assert.Nil(t, err)
}

func TestGetFriendsWhenFriendWhenAnErrorIsReturnedFromDatastoreTheSteamAPIIsUsed(t *testing.T) {
	mockController := controller.MockCntrInterface{}
	noUserFound := common.UserDocument{}
	mockController.On("GetUserFromDataStore", mock.AnythingOfType("string")).Return(noUserFound, errors.New("test error"))
	mockController.On("CallGetFriends", mock.AnythingOfType("string")).Return(testUser.FriendIDs, nil)

	didExistInDatastore, friends, err := GetFriends(&mockController, testUser.AccDetails.SteamID)

	mockController.AssertNumberOfCalls(t, "GetUserFromDataStore", 1)
	mockController.AssertNumberOfCalls(t, "CallGetFriends", 1)

	assert.False(t, didExistInDatastore)
	assert.Equal(t, testUser.FriendIDs, friends)
	assert.Nil(t, err)
}
func TestGetFriendsWhenFriendIsNotFoundFromDatastoreAndSteamAPIIsCalled(t *testing.T) {
	mockController := controller.MockCntrInterface{}
	noUserFound := common.UserDocument{}
	mockController.On("GetUserFromDataStore", mock.AnythingOfType("string")).Return(noUserFound, nil)
	mockController.On("CallGetFriends", mock.AnythingOfType("string")).Return(noUserFound.FriendIDs, nil)

	didExistInDatastore, friends, err := GetFriends(&mockController, testUser.AccDetails.SteamID)

	mockController.AssertNumberOfCalls(t, "CallGetFriends", 1)

	assert.False(t, didExistInDatastore)
	assert.Equal(t, noUserFound.FriendIDs, friends)
	assert.Nil(t, err)
}

func TestGetFriendsWhenFriendIsNotFoundFromDatastoreAndNoUserIsFoundInSteamAPI(t *testing.T) {
	mockController := controller.MockCntrInterface{}
	noUserFound := common.UserDocument{}
	mockController.On("GetUserFromDataStore", mock.AnythingOfType("string")).Return(noUserFound, nil)
	mockController.On("CallGetFriends", mock.AnythingOfType("string")).Return(noUserFound.FriendIDs, errors.New("no users found error"))

	didExistInDatastore, friends, err := GetFriends(&mockController, testUser.AccDetails.SteamID)

	mockController.AssertNumberOfCalls(t, "CallGetFriends", 1)

	assert.False(t, didExistInDatastore)
	assert.Equal(t, []string{}, friends)
	assert.NotNil(t, err)
}

func TestGetTopTwentyOrFewerGames(t *testing.T) {
	expectedFirstGame := "CS Source"
	expectedSecondGame := "CS:GO"
	gamesList := []common.Game{
		{
			Name:            "CS:GO",
			PlaytimeForever: 1337,
		},
		{
			Name:            expectedFirstGame,
			PlaytimeForever: 199999,
		},
	}

	sortedGames := getTopTwentyOrFewerGames(gamesList)

	assert.Equal(t, expectedFirstGame, sortedGames[0].Name)
	assert.Equal(t, expectedSecondGame, sortedGames[1].Name)
	assert.Len(t, sortedGames, 2)
}

func TestGetTopTwentyOrFewerGamesOnlyReturnsTwentyOrFewerGames(t *testing.T) {
	gamesList := []common.Game{}
	for i := 0; i < 22; i++ {
		gamesList = append(gamesList, common.Game{
			Appid:           i,
			PlaytimeForever: i,
		})
	}

	sortedGames := getTopTwentyOrFewerGames(gamesList)

	assert.Len(t, sortedGames, 20)
}
