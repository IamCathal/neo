package worker

import (
	"os"
	"testing"

	"github.com/iamcathal/neo/services/crawler/datastructures"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}

func TestVerifyFormatOfSteamIDsVerifiesTwoValidSteamIDs(t *testing.T) {
	expectedSteamIDs := []int64{12345678901234456, 72348978301996243}
	inputData := datastructures.CrawlUsersInput{
		FirstSteamID:  expectedSteamIDs[0],
		SecondSteamID: expectedSteamIDs[1],
	}

	receivedValidSteamIDs, err := VerifyFormatOfSteamIDs(inputData)

	assert.Nil(t, err)
	assert.Equal(t, expectedSteamIDs, receivedValidSteamIDs, "expect two valid format steamIDs are returned")
}

func TestVerifyFormatOfSteamIDsReturnsNothingForTwoInvalidFormatSteamIDs(t *testing.T) {
	expectedSteamIDs := []int64{12345634456, 0}
	inputData := datastructures.CrawlUsersInput{
		FirstSteamID:  expectedSteamIDs[0],
		SecondSteamID: expectedSteamIDs[1],
	}

	receivedValidSteamIDs, err := VerifyFormatOfSteamIDs(inputData)

	assert.Nil(t, err)
	assert.Len(t, receivedValidSteamIDs, 0, "expect to receive back no steamIDs for two invalid steamID inputs")
}
