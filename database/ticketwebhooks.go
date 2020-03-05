package database

type TicketWebhook struct {
	Uuid     string `gorm:"column:UUID;type:varchar(36);unique;primary_key"`
	WebhookUrl   string `gorm:"column:CDNURL;type:varchar(200)"`
}

func (TicketWebhook) TableName() string {
	return "webhooks"
}

func (w *TicketWebhook) AddWebhook() {
	Db.Create(w)
}

func DeleteWebhookByUuid(uuid string) {
	Db.Where(TicketWebhook{Uuid: uuid}).Delete(TicketWebhook{})
}

func GetWebhookByUuid(uuid string, res chan *string) {
	var row TicketWebhook
	Db.Where(TicketWebhook{Uuid: uuid}).Take(&row)

	if row.WebhookUrl == "" {
		res <- nil
	} else {
		res <- &row.WebhookUrl
	}
}
