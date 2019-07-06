package database

import (
	"github.com/TicketsBot/TicketsGo/sentry"
	"github.com/t-tiger/gorm-bulk-insert"
)

type UserData struct {
	UserId        int64  `gorm:"column:USERID;unique;primary_key"`
	Username      string `gorm:"column:USERNAME;type:text"`
	Discriminator string `gorm:"column:DISCRIM;type:varchar(4)"`
	Avatar        string `gorm:"column:AVATARHASH;type:varchar(100)"`
}

func (UserData) TableName() string {
	return "usernames"
}

func UpdateUser(id int64, name string, discrim string, avatarHash string) {
	var node UserData
	Db.Where(UserData{UserId: id}).Assign(&UserData{Username: name, Discriminator: discrim, Avatar: avatarHash}).FirstOrCreate(&node)
}

// We don't need to update / upsert because this should be for initial data only when we first receive the guold
func InsertUsers(data []UserData) {
	records := make([]interface{}, 0)
	for _, record := range data {
		records = append(records, record)
	}

	if err := gormbulk.BulkInsert(Db.DB, records, 2000); err != nil {
		sentry.Error(err)
	}
}

func GetUsername(id int64, ch chan string) {
	var node UserData
	Db.Where(UserData{UserId: id}).Take(&node)
	ch <- node.Username
}
