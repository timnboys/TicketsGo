package database

import (
	"fmt"
	"github.com/satori/go.uuid"
	"strings"
	"time"
)

type Ticket struct {
	Uuid     string `gorm:"column:UUID;type:varchar(36);unique;primary_key"`
	Id       int    `gorm:"column:ID"`
	Guild    int64  `gorm:"column:GUILDID"`
	Channel  *int64  `gorm:"column:CHANNELID;nullable"`
	Owner    int64  `gorm:"column:OWNERID"`
	Members  string `gorm:"column:MEMBERS;type:text"`
	IsOpen   bool   `gorm:"column:OPEN"`
	OpenTime *int64 `gorm:"column:OPENTIME;nullable"`
}

func (Ticket) TableName() string {
	return "tickets"
}

func GetNextId(guild int64, ch chan int) {
	var node Ticket
	Db.Where(Ticket{Guild: guild}).Order("ID desc").First(&node)
	ch <- node.Id + 1
}

func CreateTicket(guild, owner int64, ch chan int) {
	idChan := make(chan int)
	go GetNextId(guild, idChan)
	id := <-idChan

	now := time.Now().UnixNano() / int64(time.Millisecond)

	node := Ticket{
		Uuid:     uuid.Must(uuid.NewV4()).String(),
		Id:       id,
		Guild:    guild,
		Channel:  nil,
		Owner:    owner,
		Members:  "",
		IsOpen:   true,
		OpenTime: &now,
	}

	Db.Create(&node)

	ch <- id
}

func SetTicketChannel(id int, guild int64, channel int64) {
	var node Ticket

	Db.Where(Ticket{Id: id, Guild: guild}).First(&node)
	Db.Model(&node).Update("CHANNELID", channel)
}

func GetMembers(ticket int, guild int64, ch chan []string) {
	node := Ticket{
		Id:    ticket,
		Guild: guild,
	}

	Db.Where(node).Take(&node)

	ch <- strings.Split(node.Members, ",")
}

func AddMember(ticket int, guild int64, user string) {
	node := Ticket{
		Id:    ticket,
		Guild: guild,
	}

	Db.Where(node).Take(&node)

	updated := fmt.Sprintf("%s,", user)

	Db.Where(node).Assign("MEMBERS", updated).FirstOrCreate(&node)
}

func RemoveMember(ticket int, guild int64, user string) {
	memberChan := make(chan []string)
	go GetMembers(ticket, guild, memberChan)
	members := <-memberChan

	for i, member := range members {
		if member == user || member == "" {
			members[i] = members[len(members)-1]
			members = members[:len(members)-1]
		}
	}

	var node Ticket
	Db.Where(Ticket{Id: ticket, Guild: guild}).Assign("MEMBERS", strings.Join(members, ",")).FirstOrCreate(&node)
}

func IsTicketChannel(channel int64, ch chan bool) {
	var count int
	Db.Table(Ticket{}.TableName()).Where(Ticket{Channel: &channel}).Count(&count)
	ch <- count > 0
}

func GetTicketById(guild int64, id int, ch chan Ticket) {
	var node Ticket
	Db.Where(Ticket{Guild: guild, Id: id}).Take(&node)
	ch <- node
}

func GetTicketByChannel(channel int64, ch chan Ticket) {
	var node Ticket
	Db.Where(Ticket{Channel: &channel}).Take(&node)
	ch <- node
}

func GetTicketId(channel int64, ch chan int) {
	var node Ticket
	Db.Where(Ticket{Channel: &channel}).Take(&node)
	ch <- node.Id
}

func GetOwner(ticket int, guild int64, ch chan int64) {
	var node Ticket
	Db.Where(Ticket{Guild: guild, Id: ticket}).Take(&node)
	ch <- node.Owner
}

func GetOwnerByChannel(channel int64, ch chan int64) {
	var node Ticket
	Db.Where(Ticket{Channel: &channel}).Take(&node)
	ch <- node.Owner
}

func GetTicketsOpenedBy(guild, owner int64, ch chan map[int64]int) {
	var nodes []Ticket
	Db.Where(Ticket{Guild: guild, Owner: owner}).Find(&nodes)

	tickets := make(map[int64]int)
	for _, ticket := range nodes {
		if ticket.Channel != nil {
			tickets[*ticket.Channel] = ticket.Id
		}
	}

	ch <- tickets
}

func GetTicketUuid(channel int64, ch chan string) {
	var node Ticket
	Db.Where(Ticket{Channel: &channel}).Take(&node)
	ch <- node.Uuid
}

func GetOpenTickets(guild int64, ch chan []string) {
	var nodes []Ticket
	Db.Where(Ticket{Guild: guild, IsOpen: true}).Find(&nodes)

	tickets := make([]string, 0)
	for _, ticket := range nodes {
		tickets = append(tickets, ticket.Uuid)
	}

	ch <- tickets
}

func GetOpenTicketChannelIds(guild int64, ch chan []*int64) {
	var nodes []Ticket
	Db.Where(Ticket{Guild: guild, IsOpen: true}).Find(&nodes)

	tickets := make([]*int64, 0)
	for _, ticket := range nodes {
		tickets = append(tickets, ticket.Channel)
	}

	ch <- tickets
}

func GetOpenTicketStructs(guild int64, ch chan []Ticket) {
	var nodes []Ticket
	Db.Where(Ticket{Guild: guild, IsOpen: true}).Find(&nodes)

	ch <- nodes
}

func GetOpenTicketsOpenedBy(guild, user int64, ch chan []string) {
	var nodes []Ticket
	Db.Where(Ticket{Guild: guild, Owner: user, IsOpen: true}).Find(&nodes)

	tickets := make([]string, 0)
	for _, ticket := range nodes {
		tickets = append(tickets, ticket.Uuid)
	}

	ch <- tickets
}

func GetOpenTime(uuid string, ch chan *int64) {
	var node Ticket
	Db.Where(Ticket{Uuid: uuid}).Take(&node)
	ch <- node.OpenTime
}

func GetOpenTimes(guild int64, ch chan map[string]*int64) {
	var nodes []Ticket
	Db.Where(Ticket{Guild: guild}).Find(&nodes)

	times := make(map[string]*int64, 0)
	for _, node := range nodes {
		times[node.Uuid] = node.OpenTime
	}

	ch <- times
}

func GetTotalTicketCount(guild int64, ch chan int) {
	var count int
	Db.Table(Ticket{}.TableName()).Where(Ticket{Guild: guild}).Count(&count)
	ch <- count
}

func GetTotalTicketsFromUser(guild int64, user int64, ch chan int) {
	var count int
	Db.Where(Ticket{Guild: guild, Owner: user}).Count(&count)
	ch <- count
}

func GetGlobalTicketCount(ch chan int) {
	var count int
	Db.Where(Ticket{}).Count(&count)
	ch <- count
}

func Close(guild int64, ticket int) {
	node := Ticket{Id: ticket, Guild: guild}
	Db.Model(&node).Where(node).Update("OPEN", false)
}

func CloseByChannel(channel int64) {
	node := Ticket{Channel: &channel}
	Db.Model(&node).Where(node).Update("OPEN", false)
}
