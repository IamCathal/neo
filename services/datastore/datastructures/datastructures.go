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

// From crawler

type UserDocument struct {
	SteamID            string     `json:"steamid"`
	AccDetails         Player     `json:"accDetails"`
	FriendIDs          []string   `json:"friends"`
	AmountOfGamesOwned int        `json:"amountOfGamesOwned"`
	GamesOwned         []GameInfo `json:"gamesOwned"`
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

type GameInfo struct {
	Name            string `json:"name"`
	PlaytimeForever int    `json:"playtimeForever"`
	Playtime2Weeks  int    `json:"playtime2Weeks"`
	ImgIconURL      string `json:"imgIconUrl"`
	ImgLogoURL      string `json:"imgLogoUrl"`
}
