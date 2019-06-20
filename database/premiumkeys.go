package database

import (
	uuid "github.com/satori/go.uuid"
)

type PremiumKeys struct {
	Key    string `gorm:"column:KEY;type:varchar(36);unique;primary_key"`
	Length int64     `gorm:"column:EXPIRY"`
}

func (PremiumKeys) TableName() string {
	return "premiumkeys"
}

func AddKey(length int64, ch chan uuid.UUID) {
	uuid := uuid.Must(uuid.NewV4())

	Db.Create(&PremiumKeys{
		Key: uuid.String(),
		Length: length,
	})

	ch <- uuid
}

func PopKey(key uuid.UUID, ch chan int64) {
	var node PremiumKeys
	Db.Where(PremiumKeys{Key: key.String()}).Take(&node)

	length := node.Length

	Db.Delete(&node)

	ch <- length
}

func KeyExists(key uuid.UUID, ch chan bool) {
	var count int
	Db.Table(PremiumKeys{}.TableName()).Where(PremiumKeys{Key: key.String()}).Count(&count)
	ch <- count > 0
}
