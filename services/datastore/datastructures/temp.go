package datastructures

import "github.com/neosteamfriendgraphing/common"

type UsersGraphData struct {
	UserDetails   ResStruct   `json:"userdetails"`
	FriendDetails []ResStruct `json:"frienddetails"`
}

type ResStruct struct {
	// data frields
	User common.UserDocument
	// Job fields
	FromID       string
	MaxLevel     int
	CurrentLevel int
}
