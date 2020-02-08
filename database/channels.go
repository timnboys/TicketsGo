package database

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/sentry"
	"strings"
)

type Channel struct {
	ChannelId int64  `gorm:"column:CHANNELID"`
	GuildId   int64  `gorm:"column:GUILDID"`
	Name      string `gorm:"column:NAME;type:VARCHAR(32)"`
	Type      int    `gorm:"column:CHANNELTYPE;type:TINYINT(1)"`
}

func (Channel) TableName() string {
	return "Channel"
}

func StoreChannel(channelId, guildId int64, name string, channelType int) {
	channel := Channel{
		ChannelId: channelId,
		GuildId:   guildId,
		Name:      name,
		Type:      channelType,
	}

	Db.Where(&Channel{ChannelId:channelId}).Assign(&channel).FirstOrCreate(&Channel{})
}

func DeleteChannel(channelId int64) {
	var node Channel
	Db.Where(Channel{ChannelId: channelId}).Take(&node)
	Db.Delete(&node)
}

func GetCachedChannelsByGuild(guildId int64, res chan []Channel) {
	var nodes []Channel
	Db.Where(Channel{GuildId: guildId}).Find(&nodes)
	res <- nodes
}

func InsertChannels(data []Channel) {
	records := make([]interface{}, 0)
	for _, record := range data {
		records = append(records, record)
	}

	bulkInsertChannels(data)
}

func bulkInsertChannels(data []Channel) {
	chunks := make([][]Channel, 0)
	temp := make([]Channel, 0)

	// MySQL has a variable limit, so split into chunks
	for _, record := range data {
		temp = append(temp, record)

		if len(temp) >= 2000 {
			chunks = append(chunks, temp)
			temp = make([]Channel, 0)
		}
	}

	for _, chunk := range chunks {
		values := make([]string, 0)
		args := make([]interface{}, 0)

		for _, record := range chunk {
			values = append(values, "(?, ?, ?, ?)")
			args = append(args, record.ChannelId)
			args = append(args, record.GuildId)
			args = append(args, record.Name)
			args = append(args, record.Type)
		}

		statement := fmt.Sprintf("INSERT INTO Channel(CHANNELID, GUILDID, NAME, CHANNELTYPE) VALUES %s ON DUPLICATE KEY UPDATE NAME=VALUES(NAME);", strings.Join(values, ","))
		if _, err := Db.DB.DB().Exec(statement, args...); err != nil {
			sentry.Error(err)
		}
	}
}
