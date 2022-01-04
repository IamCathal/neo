package datastructures

import (
	"time"

	"github.com/neosteamfriendgraphing/common"
)

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
	CrawlID               string `json:"crawlID"`

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

type CreateGraph struct {
	CrawlID string `json:"crawlid"`
}

type UsersGraphData struct {
	UserDetails       ResStruct      `json:"userdetails"`
	FriendDetails     []ResStruct    `json:"frienddetails"`
	TopTenGameDetails []BareGameInfo `json:"toptengamedetails"`
}

type CrawlJob struct {
	SteamID string

	FromID       string
	MaxLevel     int
	CurrentLevel int
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
	Status string         `json:"status"`
	Games  []BareGameInfo `json:"games"`
}

type GetDetailsForGamesInputDTO struct {
	GameIDs []int `json:"gameids"`
}
