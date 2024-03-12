package model

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/tristan-club/cascade/pconst"
	rdb "github.com/tristan-club/cascade/pkg/redisutil"
	tlog "github.com/tristan-club/kit/log"
	"strconv"
)

func GetUserScore(groupId, openId int64) (int, error) {
	result, err := rdb.Default().HGet(context.Background(), pconst.GetRKScore(groupId), strconv.FormatInt(openId, 10)).Int()
	if err != nil && err != redis.Nil {
		return 0, err
	}
	return result, nil
}

func AddUserScore(groupId, openId, score int64) error {
	return rdb.Default().HIncrBy(context.Background(), pconst.GetRKScore(groupId), strconv.FormatInt(openId, 10), score).Err()
}

func GetUserScores(groupId int64) (map[int64]int, error) {
	m := make(map[int64]int, 0)

	ctx := context.Background()
	results, err := rdb.Default().HGetAll(ctx, pconst.GetRKScore(groupId)).Result()
	if err != nil && err != redis.Nil {
		return nil, err
	}

	for openIdStr, scoreStr := range results {

		openId, err := strconv.ParseInt(openIdStr, 10, 64)
		if err != nil {
			tlog.Error().Fields(map[string]interface{}{"action": "parse openId error", "error": err.Error()}).Send()
		} else {
			score, err := strconv.ParseInt(scoreStr, 10, 64)
			if err != nil {
				tlog.Error().Fields(map[string]interface{}{"action": "invalid score", "error": err.Error(), "openId": openIdStr}).Send()
			}
			m[openId] = int(score)
		}
	}

	return m, nil
}
