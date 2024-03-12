package pconst

import "fmt"

const (
	RKPrefixGroupId = "group:id:"
)

const (
	RKUserData         = "user:data"
	RKSignInConfig     = "signin:config:"
	RKSignInLimiter    = "signin:limiter:"
	RkSignInConfigList = "signin:config-list:"
	RKScore            = "score:"
	RKUserStatus       = "user:status"
	RKGroupClaimed     = "group:claimed"
)

func GetRKUser(openId int64) string {
	return fmt.Sprintf("%s%d", RKUserData, openId)
}

func GetRKScore(groupId int64) string {
	return fmt.Sprintf("%s%d", RKScore, groupId)
}

func GetRkSignInConfig(groupId int64) string {
	return fmt.Sprintf("%s%d", RKSignInConfig, groupId)
}

func GetRkSignInConfigList(groupId int64) string {
	return fmt.Sprintf("%s%d", RkSignInConfigList, groupId)
}
