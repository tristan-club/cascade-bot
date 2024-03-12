package model

import (
	"github.com/tristan-club/cascade/pkg/dbutil"
	"time"
)

type SigninRecord struct {
	dbutil.BaseModelID
	GroupId     int64  `gorm:"column:group_id" json:"group_id" redis:"group_id"`
	OpenId      int64  `gorm:"column:open_id" json:"open_id" redis:"open_id"`
	Reward      uint64 `gorm:"column:reward" json:"reward" redis:"reward"`
	ConfigId    uint64 `gorm:"column:config_id" json:"config_id" redis:"config_id"`
	ExtraReward uint64 `gorm:"column:extra_reward" json:"extra_reward" redis:"extra_reward"`
	dbutil.BaseModelAt
}

func NewSignInUser(groupId int64, openId int64, er, reward uint64) *SigninRecord {
	return &SigninRecord{GroupId: groupId, OpenId: openId, Reward: reward, ExtraReward: er}
}

func SaveSignInRecord(sr *SigninRecord) error {
	return dbutil.Default().Save(sr).Error
}

func GetSignInRecord(groupId int64, fromTime time.Time) (list []*SigninRecord, err error) {
	return list, dbutil.Default().Model(&SigninRecord{}).Where("group_id = ? and created_at >= ? ", groupId, fromTime).Find(&list).Error
}
