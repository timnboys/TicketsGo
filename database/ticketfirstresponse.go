package database

import "time"

type TicketFirstResponse struct {
	Ticket string `gorm:"column:UUID;type:varchar(36);unique;primary_key"`
	Guild int64 `gorm:"column:GUILDID"`
	Responder int64 `gorm:"column:USERID"`
	Time int64 `gorm:"column:RESPONSETIME"`
}

func (TicketFirstResponse) TableName() string {
	return "ticketfirstresponse"
}

func AddResponseTime(ticket string, guild int64, responder int64) {
	openTimeChan := make(chan *int64)
	go GetOpenTime(ticket, openTimeChan)
	openTime := <-openTimeChan
	if openTime == nil {
		return
	}

	current := time.Now().UnixNano() / int64(time.Millisecond)

	Db.Create(&TicketFirstResponse{
		Ticket: ticket,
		Guild: guild,
		Responder: responder,
		Time: current - *openTime,
	})
}

func HasResponse(ticket string, ch chan bool) {
	var count int
	Db.Table(TicketFirstResponse{}.TableName()).Where(TicketFirstResponse{Ticket: ticket}).Count(&count)
	ch <- count != 0
}

func GetGuildResponseTimes(guild int64, ch chan map[string]int64) {
	var nodes []TicketFirstResponse
	Db.Where(TicketFirstResponse{Guild: guild}).Find(&nodes)

	times := make(map[string]int64)
	for _, node := range nodes {
		times[node.Ticket] = node.Time
	}

	ch <- times
}

func GetUserResponseTimes(guild int64, user int64, ch chan map[string]int64) {
	var nodes []TicketFirstResponse
	Db.Where(TicketFirstResponse{Guild: guild, Responder: user}).Find(&nodes)

	times := make(map[string]int64)
	for _, node := range nodes {
		times[node.Ticket] = node.Time
	}

	ch <- times
}
