package datastructures

import (
	"sync"
	"time"

	"github.com/streadway/amqp"
)

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

type AmqpChannel struct {
	Channel amqp.Channel
	Lock    *sync.Mutex
}

type CrawlJob struct {
	CrawlID string
	SteamID string

	FromID       string
	MaxLevel     int
	CurrentLevel int
}

type CrawlUserTempDTO struct {
	Level    int      `json:"level"`
	SteamIDs []string `json:"steamids"`
}

type CrawlResponseDTO struct {
	Status   string   `json:"status"`
	CrawlIDs []string `json:"crawlids"`
}
