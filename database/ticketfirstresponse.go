package database

import "time"

type TicketFirstResponse struct {
	Ticket    string `gorm:"column:UUID;type:varchar(36);unique;primary_key"`
	Guild     uint64 `gorm:"column:GUILDID"`
	Responder uint64 `gorm:"column:USERID"`
	Time      int64  `gorm:"column:RESPONSETIME"`
}

func (TicketFirstResponse) TableName() string {
	return "ticketfirstresponse"
}

func AddResponseTime(ticket string, guild uint64, responder uint64) {
	openTimeChan := make(chan *int64)
	go GetOpenTime(ticket, openTimeChan)
	openTime := <-openTimeChan
	if openTime == nil {
		return
	}

	current := time.Now().UnixNano() / int64(time.Millisecond)

	Db.Create(&TicketFirstResponse{
		Ticket:    ticket,
		Guild:     guild,
		Responder: responder,
		Time:      current - *openTime,
	})
}

func HasResponse(ticket string, ch chan bool) {
	var count int
	Db.Table(TicketFirstResponse{}.TableName()).Where(TicketFirstResponse{Ticket: ticket}).Count(&count)
	ch <- count != 0
}

func GetGuildResponseTimes(guild uint64, ch chan map[string]int64) {
	var nodes []TicketFirstResponse
	Db.Where(TicketFirstResponse{Guild: guild}).Find(&nodes)

	times := make(map[string]int64)
	for _, node := range nodes {
		times[node.Ticket] = node.Time
	}

	ch <- times
}

func GetUserResponseTimes(guild uint64, user uint64, ch chan map[string]int64) {
	var nodes []TicketFirstResponse
	Db.Where(TicketFirstResponse{Guild: guild, Responder: user}).Find(&nodes)

	times := make(map[string]int64)
	for _, node := range nodes {
		times[node.Ticket] = node.Time
	}

	ch <- times
}
