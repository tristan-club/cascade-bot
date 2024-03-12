package redisutil

import (
	"context"
	"github.com/tristan-club/kit/log"
	"strconv"
)
import "github.com/redis/go-redis/v9"

var rdb = &redis.Client{}

func Default() *redis.Client {
	if rdb == nil {
		return &redis.Client{}
	}
	return rdb
}

func InitRedis(svc string, db string) error {
	_db, err := strconv.ParseInt(db, 10, 64)
	if err != nil {
		return err
	}
	rdb = redis.NewClient(&redis.Options{
		Addr: svc,
		DB:   int(_db),
	})

	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "init redis client error", "error": err.Error()}).Send()
		return err
	}

	log.Info().Msgf("init redis client addr %s db %s success", svc, db)

	return nil
}
