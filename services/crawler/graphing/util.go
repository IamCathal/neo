package graphing

import "github.com/iamcathal/neo/services/crawler/datastructures"

func doesExistInMap(userMap map[string]bool, username string) bool {
	_, ok := userMap[username]
	if ok {
		return true
	}
	return false
}

func getAllSteamIDsFromJobsWithNoAssociatedUsernames(jobs []datastructures.ResStruct) []string {
	steamIDs := []string{}
	for _, job := range jobs {
		if job.User.AccDetails.Personaname == "" {
			steamIDs = append(steamIDs, job.User.AccDetails.SteamID)
		}
	}
	return steamIDs
}
