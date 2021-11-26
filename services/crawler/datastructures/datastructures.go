package datastructures

import (
	"time"
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
