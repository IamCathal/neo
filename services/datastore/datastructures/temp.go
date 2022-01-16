package datastructures

import "github.com/neosteamfriendgraphing/common"

// type UsersGraphData struct {
// 	UserDetails       ResStruct      `json:"userdetails"`
// 	FriendDetails     []ResStruct    `json:"frienddetails"`
// 	TopTenGameDetails []BareGameInfo `json:"toptengamedetails"`
// }

// type ResStruct struct {
// 	// data frields
// 	User common.UserDocument
// 	// Job fields
// 	FromID       string
// 	MaxLevel     int
// 	CurrentLevel int
// }

// type BareGameInfo struct {
// 	AppID int    `json:"appid"`
// 	Name  string `json:"name"`
// }

// type GetDetailsForGamesDTO struct {
// 	GameIDs []int `json:"gameids"`
// }

type GetProcessedGraphDataDTO struct {
	Status        string                `json:"status"`
	UserGraphData common.UsersGraphData `json:"usergraphdata"`
}

type AddUserEvent struct {
	SteamID     string `json:"steamid"`
	PersonaName string `json:"personaname"`
	ProfileURL  string `json:"profileurl"`
	Avatar      string `json:"avatar"`
	CountryCode string `json:"countrycode"`
	CrawlTime   int64  `json:"crawltime"`
}
