package graphing

func doesExistInMap(userMap map[string]bool, username string) bool {
	_, ok := userMap[username]
	if ok {
		return true
	}
	return false
}

func getAllSteamIDsFromJobsWithNoAssociatedUsernames(jobs []jobStruct) []string {
	steamIDs := []string{}
	for _, job := range jobs {
		if job.Username == "" {
			steamIDs = append(steamIDs, job.SteamID)
		}
	}
	return steamIDs
}
