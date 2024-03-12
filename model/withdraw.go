package model

import (
	"github.com/tristan-club/cascade/pkg/dbutil"
	"gorm.io/gorm"
)

type Withdraw struct {
	dbutil.BaseModelID
	GroupId   int64  `gorm:"column:group_id" json:"group_id" redis:"group_id"`
	OpenId    int64  `gorm:"column:open_id" json:"open_id" redis:"open_id"`
	Username  string `gorm:"column:username" json:"username" redis:"username"`
	Address   string `gorm:"column:address" json:"address"`
	Amount    uint64 `gorm:"column:amount" json:"amount"`
	TxHash    string `gorm:"column:tx_hash" json:"tx_hash"`
	IsProceed bool   `gorm:"column:is_proceed" json:"is_proceed"`
	dbutil.BaseModelAt
}

func NewWithdraw(groupId int64, openId int64, username string, address string, amount uint64) *Withdraw {
	return &Withdraw{GroupId: groupId, OpenId: openId, Username: username, Address: address, Amount: amount}
}

func AddWithdraw(tx *gorm.DB, entity *Withdraw) error {
	return dbutil.GetDb(tx).Create(entity).Error
}

func BatchUpdateWithdraw(ids []uint, param map[string]interface{}) error {
	return dbutil.Default().Model(&Withdraw{}).
		Where("id in (?) ", ids).
		Updates(param).Error
}

func GetShouldProcessWithdrawList() (list []*Withdraw, err error) {
	err = dbutil.Default().Model(&Withdraw{}).
		Where("is_proceed = 0").
		Find(&list).Error
	return
}
