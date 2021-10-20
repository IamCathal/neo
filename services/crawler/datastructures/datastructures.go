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

// Steam web API data structs

// UserStatsStruct is the response from the steam web API
// for /getPlayerSummary calls
type UserStatsStruct struct {
	Response Response `json:"response"`
}

// Response is filler
type Response struct {
	Players []Player `json:"players"`
}

// Player holds all details for a given user returned by the steam web API for
// the /getPlayerSummary endpoint
type Player struct {
	Steamid                  string `json:"steamid"`
	Communityvisibilitystate int    `json:"communityvisibilitystate"`
	Profilestate             int    `json:"profilestate"`
	Personaname              string `json:"personaname"`
	Commentpermission        int    `json:"commentpermission"`
	Profileurl               string `json:"profileurl"`
	Avatar                   string `json:"avatar"`
	Avatarmedium             string `json:"avatarmedium"`
	Avatarfull               string `json:"avatarfull"`
	Avatarhash               string `json:"avatarhash"`
	Personastate             int    `json:"personastate"`
	Realname                 string `json:"realname"`
	Primaryclanid            string `json:"primaryclanid"`
	Timecreated              int    `json:"timecreated"`
	Personastateflags        int    `json:"personastateflags"`
	Loccountrycode           string `json:"loccountrycode"`
}

type UserDetails struct {
	SteamID int64       `json:"steamID"`
	Friends Friendslist `json:"friendsList"`
}

// FriensdList holds all friends for a given user
type Friendslist struct {
	Friends []Friend `json:"friends"`
}

// Friend holds basic details of a friend for a given user
type Friend struct {
	Username     string `json:"username"`
	Steamid      string `json:"steamid"`
	Relationship string `json:"relationship"`
	FriendSince  int    `json:"friend_since"`
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
