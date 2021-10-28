package datastructures

import (
	"time"
)

type UptimeResponse struct {
	Status string        `json:"status"`
	Uptime time.Duration `json:"uptime"`
}

type LoggingFields struct {
	NodeName string
	NodeDC   string
	LogPaths []string
	NodeIPV4 string
}

type CrawlUsersInput struct {
	FirstSteamID  string `json:"firstSteamID"`
	SecondSteamID string `json:"secondSteamID"`
	Level         int    `json:"level"`
}

type WorkerConfig struct {
	WorkerAmount int
}

type Job struct {
	JobType               string `json:"jobType"`
	OriginalTargetSteamID string `json:"originalTargetSteamID"`
	CurrentTargetSteamID  string `json:"currentTargetSteamID"`

	MaxLevel     int `json:"maxLevel"`
	CurrentLevel int `json:"currentLevel"`
}

type APIKeysInUse struct {
	APIKeys []APIKey
}

type APIKey struct {
	Key      string
	LastUsed time.Time
}

// Datastore service DTOs

type FriendsFromDB struct {
	Exists      bool        `json:"exists"`
	FriendsList Friendslist `json:"friends"`
}

type APIResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type UserDocument struct {
	SteamID            string     `json:"steamid"`
	AccDetails         Player     `json:"accDetails"`
	FriendIDs          []string   `json:"friends"`
	AmountOfGamesOwned int        `json:"amountOfGamesOwned"`
	GamesOwned         []GameInfo `json:"gamesOwned"`
}

type GameInfo struct {
	Name            string `json:"name"`
	PlaytimeForever int    `json:"playtimeForever"`
	Playtime2Weeks  int    `json:"playtime2Weeks"`
	ImgIconURL      string `json:"imgIconUrl"`
	ImgLogoURL      string `json:"imgLogoUrl"`
}
