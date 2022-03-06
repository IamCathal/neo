package app

import (
	"github.com/neosteamfriendgraphing/common"
	"github.com/neosteamfriendgraphing/common/dtos"
)

var (
	// Two users who are directly friends with eachother
	userOneGraphData common.UsersGraphData
	userTwoGraphData common.UsersGraphData

	// Two users who share one common friend
	userOneWithOneSharedCommonFriendGraphData common.UsersGraphData
	userTwoWithOneSharedCommonFriendGraphData common.UsersGraphData
	commonFriendGraphData                     common.UsersGraphData
)

func initTestData() {
	testSaveUserDTO = dtos.SaveUserDTO{
		OriginalCrawlTarget: "testID",
		CurrentLevel:        1,
		MaxLevel:            3,
		User: common.UserDocument{
			AccDetails: common.AccDetailsDocument{
				SteamID:        "testID",
				Profileurl:     "profile url",
				Avatar:         "avatar url",
				Timecreated:    1223525546,
				Loccountrycode: "IE",
			},
			FriendIDs: []string{"1234", "5678"},
		},
		GamesOwnedFull: []common.GameInfoDocument{
			{
				Name:       "CS:GO",
				ImgIconURL: "example url",
				ImgLogoURL: "example url",
			},
		},
	}

	// Two users who are directly friends with eachother
	userOneID := "12334567"
	userTwoID := "342089525"

	userOneGraphData = common.UsersGraphData{
		UserDetails: common.UsersGraphInformation{
			User: common.UserDocument{
				AccDetails: common.AccDetailsDocument{
					SteamID: userOneID,
				},
				FriendIDs: []string{userTwoID},
			},
		},
		FriendDetails: []common.UsersGraphInformation{
			{
				User: common.UserDocument{
					AccDetails: common.AccDetailsDocument{
						SteamID: userTwoID,
					},
				},
			},
		},
	}
	userTwoGraphData = common.UsersGraphData{
		UserDetails: common.UsersGraphInformation{
			User: common.UserDocument{
				AccDetails: common.AccDetailsDocument{
					SteamID: userTwoID,
				},
				FriendIDs: []string{userOneID},
			},
		},
		FriendDetails: []common.UsersGraphInformation{
			{
				User: common.UserDocument{
					AccDetails: common.AccDetailsDocument{
						SteamID: userOneID,
					},
				},
			},
		},
	}

	// Two users who share one common friend
	userOneID = "12334567"
	userTwoID = "342089525"
	commonFriendID := "5556666"

	userOneWithOneSharedCommonFriendGraphData = common.UsersGraphData{
		UserDetails: common.UsersGraphInformation{
			User: common.UserDocument{
				AccDetails: common.AccDetailsDocument{
					SteamID: userOneID,
				},
				FriendIDs: []string{commonFriendID},
			},
		},
		FriendDetails: []common.UsersGraphInformation{
			{
				User: common.UserDocument{
					AccDetails: common.AccDetailsDocument{
						SteamID: commonFriendID,
					},
					FriendIDs: []string{userOneID, userTwoID},
				},
			},
		},
	}
	userTwoWithOneSharedCommonFriendGraphData = common.UsersGraphData{
		UserDetails: common.UsersGraphInformation{
			User: common.UserDocument{
				AccDetails: common.AccDetailsDocument{
					SteamID: userTwoID,
				},
				FriendIDs: []string{commonFriendID},
			},
		},
		FriendDetails: []common.UsersGraphInformation{
			{
				User: common.UserDocument{
					AccDetails: common.AccDetailsDocument{
						SteamID: commonFriendID,
					},
					FriendIDs: []string{userOneID, userTwoID},
				},
			},
		},
	}
	commonFriendGraphData = common.UsersGraphData{
		UserDetails: common.UsersGraphInformation{
			User: common.UserDocument{
				AccDetails: common.AccDetailsDocument{
					SteamID: commonFriendID,
				},
				FriendIDs: []string{userOneID, userTwoID},
			},
		},
		FriendDetails: []common.UsersGraphInformation{
			userOneWithOneSharedCommonFriendGraphData.UserDetails,
			userTwoWithOneSharedCommonFriendGraphData.UserDetails,
		},
	}
}
