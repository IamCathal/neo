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
	LogPath  string
	NodeIPV4 string
}

type CrawlUsersInput struct {
	FirstSteamID  int64 `json:"firstSteamID"`
	SecondSteamID int64 `json:"secondSteamID"`
	Level         int   `json:"level"`
}

type WorkerConfig struct {
	WorkerAmount int
}

type Job struct {
	JobType               string `json:"jobType"`
	OriginalTargetSteamID int64  `json:"originalTargetSteamID"`
	CurrentTargetSteamID  int64  `json:"currentTargetSteamID"`

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
