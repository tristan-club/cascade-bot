package dbutil

import (
	"time"
)

func GetOffsetAndLimit(page, limit int64) (int, int) {
	offset := int64(0)
	if limit <= 0 {
		limit = 10
	}
	if page > 1 {
		offset = (page - 1) * limit
	}
	return int(offset), int(limit)
}

type BaseModelID struct {
	Id uint `gorm:"primary_key;AUTO_INCREMENT;column:id" json:"id"`
}

type BaseModelAt struct {
	CreatedAt time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at" json:"updated_at"`
	DeletedAt *time.Time `sql:"index" json:"-"`
}
