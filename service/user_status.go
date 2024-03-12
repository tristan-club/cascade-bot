package service

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/tristan-club/cascade/pconst"
	rdb "github.com/tristan-club/cascade/pkg/redisutil"
	tlog "github.com/tristan-club/kit/log"
	"strconv"
)

func GetUserStatus(openId int64) int {
	status, err := rdb.Default().HGet(context.Background(), pconst.RKUserStatus, strconv.FormatInt(openId, 10)).Int()
	if err != nil && err != redis.Nil {
		tlog.Error().Fields(map[string]interface{}{"action": "get user status error", "openId": openId, "error": err.Error()}).Send()
	}
	return status
}

func UpdateUserStatus(openId int64, status int) {
	err := rdb.Default().HSet(context.Background(), pconst.RKUserStatus, strconv.FormatInt(openId, 10), status).Err()
	if err != nil && err != redis.Nil {
		tlog.Error().Fields(map[string]interface{}{"action": "set user status error", "openId": openId, "error": err.Error()}).Send()
	}
}

func ResetUserStatus(openId int64) {
	UpdateUserStatus(openId, -1)
}
