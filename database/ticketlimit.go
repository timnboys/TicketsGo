package database

type TicketLimit struct {
	Guild uint64 `gorm:"column:GUILDID;unique;primary_key"`
	Limit int   `gorm:"column:TICKETLIMIT"`
}

func (TicketLimit) TableName() string {
	return "ticketlimit"
}

func GetTicketLimit(guild uint64, ch chan int) {
	var node TicketLimit
	Db.Where(TicketLimit{Guild: guild}).First(&node)

	if node.Limit == 0 {
		ch <- 5
	} else {
		ch <- node.Limit
	}
}

func SetTicketLimit(guild uint64, limit int) {
	var node TicketLimit
	Db.Where(TicketLimit{Guild: guild}).Assign(TicketLimit{Limit: limit}).FirstOrCreate(&node)
}
