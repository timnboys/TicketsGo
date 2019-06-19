package database

import (
	uuid "github.com/satori/go.uuid"
)

type PremiumKeys struct {
	Key    uuid.UUID `gorm:"column:KEY;type:uuid;unique;primary_key"`
	Length int64     `gorm:"column:EXPIRY"`
}

func (PremiumKeys) TableName() string {
	return "premiumkeys"
}

func AddKey(length int64, ch chan uuid.UUID) {
	uuid := uuid.Must(uuid.NewV4())

	Db.Create(&PremiumKeys{
		Key: uuid,
		Length: length,
	})

	ch <- uuid
}

func PopKey(key uuid.UUID, ch chan int64) {
	var node PremiumKeys
	Db.Where(PremiumKeys{Key: key}).Take(&node)

	length := node.Length

	Db.Delete(&node)

	ch <- length
}

func KeyExists(key uuid.UUID, ch chan bool) {
	var count int
	Db.Where(PremiumKeys{Key: key}).Count(&count)
	ch <- count > 0
}
