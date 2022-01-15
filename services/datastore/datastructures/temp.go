package datastructures

import (
	"github.com/neosteamfriendgraphing/common"
)

type UsersGraphData struct {
	UserDetails       ResStruct      `json:"userdetails"`
	FriendDetails     []ResStruct    `json:"frienddetails"`
	TopTenGameDetails []BareGameInfo `json:"toptengamedetails"`
}

type ResStruct struct {
	// data frields
	User common.UserDocument
	// Job fields
	FromID       string
	MaxLevel     int
	CurrentLevel int
}

type BareGameInfo struct {
	AppID int    `json:"appid"`
	Name  string `json:"name"`
}

type GetDetailsForGamesDTO struct {
	GameIDs []int `json:"gameids"`
}

type GetProcessedGraphDataDTO struct {
	Status        string         `json:"status"`
	UserGraphData UsersGraphData `json:"usergraphdata"`
}

// SaveUserDTO is the input schema for saving users to the database. It takes
// the original crawl target user (that initally caused this crawl) and the
// current user to be saved
type SaveUserDTO struct {
	OriginalCrawlTarget string              `json:"orginalcrawltarget"`
	CrawlID             string              `json:"crawlid"`
	CurrentLevel        int                 `json:"currentlevel"`
	MaxLevel            int                 `json:"maxlevel"`
	User                common.UserDocument `json:"user"`
}

type CrawlingStatus struct {
	TimeStarted         int64  `json:"timestarted"`
	CrawlID             string `json:"crawlid"`
	OriginalCrawlTarget string `json:"originalcrawltarget"`
	MaxLevel            int    `json:"maxlevel"`
	TotalUsersToCrawl   int    `json:"totaluserstocrawl"`
	UsersCrawled        int    `json:"userscrawled"`
}

type SaveCrawlingStatsDTO struct {
	CurrentLevel   int            `json:"currentlevel"`
	CrawlingStatus CrawlingStatus `json:"crawlingstatus"`
}

type GetCrawlingStatusDTO struct {
	Status         string         `json:"status"`
	CrawlingStatus CrawlingStatus `json:"crawlingstatus"`
}

type HasBeenCrawledBeforeInputDTO struct {
	Level   int    `json:"level"`
	SteamID string `json:"steamid"`
}

type AddUserEvent struct {
	SteamID     string `json:"steamid"`
	PersonaName string `json:"personaname"`
	ProfileURL  string `json:"profileurl"`
	Avatar      string `json:"avatar"`
	CountryCode string `json:"countrycode"`
	CrawlTime   int64  `json:"crawltime"`
}

type DoesProcessedGraphDataExistDTO struct {
	Status string `json:"status"`
	Exists string `json:"exists"`
}
