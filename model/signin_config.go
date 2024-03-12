package model

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/tristan-club/cascade/pconst"
	rdb "github.com/tristan-club/cascade/pkg/redisutil"
	"github.com/tristan-club/kit/log"
)

const (
	DefaultSignInConfigId = 100
)

type SigninConfig struct {
	GroupId             int64  `gorm:"column:group_id" json:"group_id" redis:"group_id"`
	ConfigId            int64  `gorm:"column:config_id" json:"config_id" redis:"config_id"`
	CreatorId           int64  `gorm:"column:creator_id" json:"creator_id" redis:"creator_id"`
	Enable              bool   `gorm:"column:enable" json:"enable" redis:"enable"`
	Text                string `gorm:"column:text" json:"text" redis:"text"`
	Reward              int    `gorm:"column:reward" json:"reward" redis:"reward"`
	Symbol              string `gorm:"column:symbol" json:"symbol" redis:"symbol"`
	ExtraRewardAmount   int    `gorm:"column:extra_reward_amount" json:"extra_reward_amount" redis:"extra_reward_amount"`
	ExtraRewardQuantity int    `gorm:"column:extra_reward_quantity" json:"extra_reward_quantity" redis:"extra_reward_quantity"`
}

func (sc *SigninConfig) Active() bool {
	return sc.Enable && sc.Text != ""
}

func GetSignInConfig(groupId int64) (*SigninConfig, error) {

	sc := SigninConfig{GroupId: groupId}
	err := rdb.Default().HGetAll(context.Background(), pconst.GetRkSignInConfig(groupId)).Scan(&sc)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	return &sc, nil
}

func SaveSignInConfig(sc *SigninConfig) error {
	err := rdb.Default().HSet(context.Background(), pconst.GetRkSignInConfig(sc.GroupId), sc).Err()
	if err != nil {
		return err
	}

	err = rdb.Default().RPush(context.Background(), pconst.GetRkSignInConfigList(sc.GroupId), sc).Err()
	if err != nil {
		log.Error().Fields(map[string]interface{}{"action": "save signup config list error", "error": err.Error()}).Send()
	}

	return nil
}
