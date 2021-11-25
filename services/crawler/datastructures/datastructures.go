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

type BaseResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
