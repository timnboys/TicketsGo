package database

import (
	"github.com/satori/go.uuid"
)

type CannedResponse struct {
	Uuid string `gorm:"column:UUID;type:varchar(36);unique;primary_key"`
	Id string `gorm:"column:ID;type:varchar(16)"`
	Guild uint64 `gorm:"column:GUILDID"`
	Content string `gorm:"column:TEXT;type:TEXT"`
}

func (CannedResponse) TableName() string {
	return "cannedresponses"
}

func GetCannedResponse(guild uint64, id string, ch chan string) {
	var node CannedResponse
	Db.Where(CannedResponse{Id: id, Guild: guild}).Take(&node)
	ch <- node.Content
}

func GetCannedResponses(guild uint64, ch chan []string) {
	var nodes []CannedResponse
	Db.Where(CannedResponse{Guild: guild}).Find(&nodes)

	ids := make([]string, 0)
	for _, node := range nodes {
		ids = append(ids, node.Id)
	}

	ch <- ids
}

func AddCannedResponse(guild uint64, id string, content string) {
	Db.Create(&CannedResponse{
		Uuid: uuid.NewV4().String(),
		Id: id,
		Guild: guild,
		Content: content,
	})
}

func DeleteCannedResponse(guild uint64, id string) {
	Db.Where(map[string]interface{}{
		"ID": id,
		"GUILDID": guild,
	}).Delete(&CannedResponse{})
}