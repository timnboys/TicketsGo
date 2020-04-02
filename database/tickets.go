package database

import (
	"fmt"
	"github.com/TicketsBot/TicketsGo/sentry"
	"strconv"
	"strings"
	"time"
)

type Ticket struct {
	Uuid             string  `gorm:"column:UUID;type:varchar(36);unique;primary_key"`
	Id               int     `gorm:"column:ID"`
	Guild            uint64  `gorm:"column:GUILDID"`
	Channel          *uint64 `gorm:"column:CHANNELID;nullable"`
	Owner            uint64  `gorm:"column:OWNERID"`
	Members          string  `gorm:"column:MEMBERS;type:text"` // TODO: Refactor to make this actually acceptable
	IsOpen           bool    `gorm:"column:OPEN"`
	OpenTime         *int64 `gorm:"column:OPENTIME;nullable"`
	WelcomeMessageId *uint64 `gorm:"column:WELCOMEMESSAGEID;nullable"`
}

func (Ticket) TableName() string {
	return "tickets"
}

func GetNextId(guild uint64, ch chan int) {
	var node Ticket
	Db.Where(Ticket{Guild: guild}).Order("ID desc").First(&node)
	ch <- node.Id + 1
}

func CreateTicket(uuid string, guild, owner uint64, ch chan int) {
	idChan := make(chan int)
	go GetNextId(guild, idChan)
	id := <-idChan

	now := time.Now().UnixNano() / int64(time.Millisecond)

	node := Ticket{
		Uuid:     uuid,
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

func SetTicketChannel(id int, guild uint64, channel uint64) {
	var node Ticket

	Db.Where(Ticket{Id: id, Guild: guild}).First(&node)
	Db.Model(&node).Update("CHANNELID", channel)
}

func GetMembers(ticket int, guild uint64, ch chan []uint64) {
	node := Ticket{
		Id:    ticket,
		Guild: guild,
	}

	Db.Where(node).Take(&node)

	ids := make([]uint64, 0)
	for _, id := range strings.Split(node.Members, ",") {
		parsed, err := strconv.ParseUint(id, 10, 64); if err != nil {
			sentry.Error(err)
			continue
		}

		ids = append(ids, parsed)
	}

	ch <- ids
}

func AddMember(ticket int, guild, user uint64) {
	node := Ticket{
		Id:    ticket,
		Guild: guild,
	}

	Db.Where(node).Take(&node)

	updated := fmt.Sprintf("%d,", user)

	Db.Where(node).Assign("MEMBERS", updated).FirstOrCreate(&node)
}

func RemoveMember(ticket int, guild, user uint64) {
	memberChan := make(chan []uint64)
	go GetMembers(ticket, guild, memberChan)
	members := <-memberChan

	for i, member := range members {
		if member == user || member == 0 {
			members[i] = members[len(members)-1]
			members = members[:len(members)-1]
		}
	}

	var s string
	for _, member := range members {
		s += strconv.FormatUint(member, 10)
		s += ","
	}

	var node Ticket
	Db.Where(Ticket{Id: ticket, Guild: guild}).Assign("MEMBERS", s).FirstOrCreate(&node)
}

func IsTicketChannel(channel uint64, ch chan bool) {
	var count int
	Db.Table(Ticket{}.TableName()).Where(Ticket{Channel: &channel}).Count(&count)
	ch <- count > 0
}

func GetTicketById(guild uint64, id int, ch chan Ticket) {
	var node Ticket
	Db.Where(Ticket{Guild: guild, Id: id}).Take(&node)
	ch <- node
}

func GetTicketByUuid(uuid string, ch chan Ticket) {
	var node Ticket
	Db.Where(Ticket{Uuid: uuid}).Take(&node)
	ch <- node
}

func GetTicketByChannel(channel uint64, ch chan Ticket) {
	var node Ticket
	Db.Where(Ticket{Channel: &channel}).Take(&node)
	ch <- node
}

func GetTicketId(channel uint64, ch chan int) {
	var node Ticket
	Db.Where(Ticket{Channel: &channel}).Take(&node)
	ch <- node.Id
}

func GetOwner(ticket int, guild uint64, ch chan uint64) {
	var node Ticket
	Db.Where(Ticket{Guild: guild, Id: ticket}).Take(&node)
	ch <- node.Owner
}

func GetOwnerByChannel(channel uint64, ch chan uint64) {
	var node Ticket
	Db.Where(Ticket{Channel: &channel}).Take(&node)
	ch <- node.Owner
}

func GetTicketsOpenedBy(guild, owner uint64, ch chan map[uint64]int) {
	var nodes []Ticket
	Db.Where(Ticket{Guild: guild, Owner: owner}).Find(&nodes)

	tickets := make(map[uint64]int)
	for _, ticket := range nodes {
		if ticket.Channel != nil {
			tickets[*ticket.Channel] = ticket.Id
		}
	}

	ch <- tickets
}

func GetTicketUuid(channel uint64, ch chan string) {
	var node Ticket
	Db.Where(Ticket{Channel: &channel}).Take(&node)
	ch <- node.Uuid
}

func GetOpenTickets(guild uint64, ch chan []string) {
	var nodes []Ticket
	Db.Where(Ticket{Guild: guild, IsOpen: true}).Find(&nodes)

	tickets := make([]string, 0)
	for _, ticket := range nodes {
		tickets = append(tickets, ticket.Uuid)
	}

	ch <- tickets
}

func GetOpenTicketChannelIds(guild uint64, ch chan []*uint64) {
	var nodes []Ticket
	Db.Where(Ticket{Guild: guild, IsOpen: true}).Find(&nodes)

	tickets := make([]*uint64, 0)
	for _, ticket := range nodes {
		tickets = append(tickets, ticket.Channel)
	}

	ch <- tickets
}

func GetOpenTicketStructs(guild uint64, ch chan []Ticket) {
	var nodes []Ticket
	Db.Where(Ticket{Guild: guild, IsOpen: true}).Find(&nodes)

	ch <- nodes
}

func GetOpenTicketsOpenedBy(guild, user uint64, ch chan []string) {
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

func GetOpenTimes(guild uint64, ch chan map[string]*int64) {
	var nodes []Ticket
	Db.Where(Ticket{Guild: guild}).Find(&nodes)

	times := make(map[string]*int64, 0)
	for _, node := range nodes {
		times[node.Uuid] = node.OpenTime
	}

	ch <- times
}

func GetTotalTicketCount(guild uint64, ch chan int) {
	var count int
	Db.Table(Ticket{}.TableName()).Where(Ticket{Guild: guild}).Count(&count)
	ch <- count
}

func GetTotalTicketsFromUser(guild uint64, user uint64, ch chan int) {
	var count int
	Db.Where(Ticket{Guild: guild, Owner: user}).Count(&count)
	ch <- count
}

func GetGlobalTicketCount(ch chan int) {
	var count int
	Db.Where(Ticket{}).Count(&count)
	ch <- count
}

func Close(guild uint64, ticket int) {
	node := Ticket{Id: ticket, Guild: guild}
	Db.Model(&node).Where(node).Update("OPEN", false)
}

func CloseByChannel(channel uint64) {
	node := Ticket{Channel: &channel}
	Db.Model(&node).Where(node).Update("OPEN", false)
}

func SetWelcomeMessageId(ticketId int, guildId, msgId uint64) {
	node := Ticket{Guild: guildId, Id: ticketId}
	Db.Model(&node).Where(node).Update("WELCOMEMESSAGEID", msgId)
}
