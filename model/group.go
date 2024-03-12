package model

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/tristan-club/cascade/pconst"
	rdb "github.com/tristan-club/cascade/pkg/redisutil"
	"strconv"
)

type Group struct {
}

func GetGroupWithdraw(groupId int64) (uint64, error) {
	result, err := rdb.Default().HGet(context.Background(), pconst.RKGroupClaimed, strconv.FormatInt(groupId, 10)).Int64()
	if err != nil && err != redis.Nil {
		return 0, err
	}
	return uint64(result), nil
}

func AddGroupClaimed(groupId int64, claimed uint64) error {
	return rdb.Default().HIncrBy(context.Background(), pconst.RKGroupClaimed, strconv.FormatInt(groupId, 10), int64(claimed)).Err()
}

func GetGroupMemberCount(groupId int64) (int, error) {
	result, err := rdb.Default().HLen(context.Background(), pconst.GetRKScore(groupId)).Result()
	if err != nil && err != redis.Nil {
		return 0, err
	}
	return int(result), nil
}
