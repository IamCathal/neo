package datastructures

import "github.com/neosteamfriendgraphing/common"

// SaveUserDTO is the input schema for saving users to the database. It takes
// the original crawl target user (that initally caused this crawl) and the
// current user to be saved
type SaveUserDTO struct {
	OriginalCrawlTarget string       `json:"orginalcrawltarget"`
	CurrentLevel        int          `json:"currentlevel"`
	MaxLevel            int          `json:"maxlevel"`
	User                UserDocument `json:"user"`
}

// UserDocument is the schema for information stored for a given user
type UserDocument struct {
	SteamID    string        `json:"steamid"`
	AccDetails common.Player `json:"accdetails"`
	Friends    []string      `json:"friends"`
	GamesOwned []GameInfo    `json:"gamesowned"`
}

// GamInfo is the schema for information stored for each steam game
type GameInfo struct {
	Name            string `json:"name"`
	PlaytimeForever int    `json:"playtimeforever"`
	Playtime2Weeks  int    `json:"playtime2weeks"`
	ImgIconURL      string `json:"imgiconurl"`
	ImgLogoURL      string `json:"imglogourl"`
}

// CrawlingStatus stores the total number of friends to crawl
// for a given user and the number of profiles that have been
// crawled so far. This is used to keep track of when a given
// user has been crawled completely and processing of their
// data should start
type CrawlingStatus struct {
	OriginalCrawlTarget string `json:"originalcrawltarget"`
	MaxLevel            int    `json:"maxlevel"`
	TotalUsersToCrawl   int    `json:"totaluserstocrawl"`
	UsersCrawled        int    `json:"userscrawled"`
}
