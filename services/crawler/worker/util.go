package worker

import (
	"regexp"
	"strconv"

	"github.com/iamcathal/neo/services/crawler/datastructures"
)

func VerifyFormatOfSteamIDs(input datastructures.CrawlUsersInput) ([]int64, error) {
	validSteamIDs := []int64{}
	match, err := regexp.MatchString("([0-9]){17}", strconv.FormatInt(input.FirstSteamID, 10))
	if err != nil {
		return validSteamIDs, err
	}
	if match {
		validSteamIDs = append(validSteamIDs, input.FirstSteamID)
	}

	match, err = regexp.MatchString("([0-9]){17}", strconv.FormatInt(input.SecondSteamID, 10))
	if err != nil {
		return validSteamIDs, err
	}
	if match {
		validSteamIDs = append(validSteamIDs, input.SecondSteamID)
	}
	return validSteamIDs, nil
}
