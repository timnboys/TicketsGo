package database

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/sentry"
	"strings"
)

type UserData struct {
	UserId        int64  `gorm:"column:USERID;unique;primary_key"`
	Username      string `gorm:"column:USERNAME;type:text CHARACTER SET utf8 COLLATE utf8_unicode_ci"`
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

	bulkInsert(data)
}

func GetUsername(id int64, ch chan string) {
	var node UserData
	Db.Where(UserData{UserId: id}).Take(&node)
	ch <- node.Username
}

func bulkInsert(data []UserData) {
	chunks := make([][]UserData, 0)
	temp := make([]UserData, 0)

	// MySQL has a variable limit, so split into chunks
	for _, record := range data {
		temp = append(temp, record)

		if len(temp) >= 2000 {
			chunks = append(chunks, temp)
			temp = make([]UserData, 0)
		}
	}

	for _, chunk := range chunks {
		values := make([]string, 0)
		args := make([]interface{}, 0)

		for _, record := range chunk {
			values = append(values, "(?, ?, ?, ?)")
			args = append(args, record.UserId)
			args = append(args, record.Username)
			args = append(args, record.Discriminator)
			args = append(args, record.Avatar)
		}

		statement := fmt.Sprintf("INSERT INTO usernames(USERID, USERNAME, DISCRIM, AVATARHASH) VALUES %s ON DUPLICATE KEY UPDATE USERNAME=VALUES(USERNAME), DISCRIM=VALUES(DISCRIM), AVATARHASH=VALUES(AVATARHASH)", strings.Join(values, ","))
		if _, err := Db.DB.DB().Exec(statement, args...); err != nil {
			sentry.Error(err)
		}
	}
}

