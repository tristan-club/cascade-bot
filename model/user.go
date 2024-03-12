package model

import (
	"errors"
	"github.com/tristan-club/cascade/pkg/dbutil"
	"github.com/tristan-club/kit/log"
	"gorm.io/gorm"
)

type User struct {
	dbutil.BaseModelID
	OpenId    int64  `gorm:"column:open_id" json:"open_id" redis:"open_id"`
	Username  string `gorm:"column:username" json:"username" redis:"username"`
	FirstName string `gorm:"column:first_name" json:"first_name" redis:"first_name"`
	LastName  string `gorm:"column:last_name" json:"last_name" redis:"last_name"`
	Addr      string `gorm:"column:addr" json:"addr" redis:"addr"`
	dbutil.BaseModelAt
}

func (u *User) HaveAddress() bool {
	return u.Addr != ""
}

func NewUser(openId int64, username string, firstName string, lastName string) *User {
	return &User{OpenId: openId, Username: username, FirstName: firstName, LastName: lastName}
}

func AddUser(u *User) error {

	var _u *User
	err := dbutil.Default().Model(&User{}).Where("open_id = ? ", u.OpenId).First(&_u).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Error().Fields(map[string]interface{}{"action": "get user error", "error": err.Error()}).Send()
		return err
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = dbutil.Default().Model(&User{}).Save(&u).Error
		if err != nil {
			log.Error().Fields(map[string]interface{}{"action": "save user error", "error": err.Error()}).Send()
			return err
		}
	}

	return nil
}

func GetUser(openId int64) (*User, error) {
	var u *User
	err := dbutil.Default().Model(&User{}).Where("open_id = ? ", openId).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return u, nil
}

func GetUserByUsername(username string) (*User, error) {
	var u *User
	err := dbutil.Default().Model(&User{}).Where("username = ? ", username).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return u, nil
}

func GetUserByAddress(addr string) (*User, error) {
	var u *User
	err := dbutil.Default().Model(&User{}).Where("addr = ? ", addr).First(&u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		} else {
			return nil, err
		}
	}
	return u, nil
}

func BatchGetUser(openIds []int64) ([]*User, error) {
	list := make([]*User, 0)
	err := dbutil.Default().Model(&User{}).Where("open_id in (?) ", openIds).Find(&list).Order("id").Error
	return list, err
}

func SaveUser(u *User) error {
	return dbutil.Default().Save(u).Error
}
