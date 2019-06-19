package database

import "github.com/satori/go.uuid"

type CannedResponse struct {
	Uuid string `gorm:"column:UUID;type:varchar(36);unique;primary_key"`
	Id string `gorm:"column:ID;type:varchar(16)"`
	Guild int64 `gorm:"column:GUILDID"`
	Content string `gorm:"column:TEXT;type:TEXT"`
}

func (CannedResponse) TableName() string {
	return "cannedresponses"
}

func GetCannedResponse(guild int64, id string, ch chan string) {
	var node CannedResponse
	Db.Where(CannedResponse{Id: id, Guild: guild}).Take(&node)
	ch <- node.Content
}

func GetCannedResponses(guild int64, ch chan []string) {
	var nodes []CannedResponse
	Db.Where(CannedResponse{Guild: guild}).Find(&nodes)

	ids := make([]string, 0)
	for _, node := range nodes {
		ids = append(ids, node.Id)
	}

	ch <- ids
}

func AddCannedResponse(guild int64, id string, content string) {
	Db.Create(&CannedResponse{
		Uuid: uuid.Must(uuid.NewV4()).String(),
		Id: id,
		Guild: guild,
		Content: content,
	})
}

func DeleteCannedResponse(guild int64, id string) {
	Db.Delete(&CannedResponse{Id: id, Guild: guild})
}