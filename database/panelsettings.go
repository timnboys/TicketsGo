package database

type PanelSettings struct {
	GuildId int64 `gorm:"column:GUILDID"`
	Title string `gorm:"column:TITLE;type:VARCHAR(255)"`
	Content string `gorm:"column:CONTENT;type:TEXT"`
	Colour int `gorm:"column:COLOUR`
}

func (PanelSettings) TableName() string {
	return "panelsettings"
}

func GetPanelSettings(guildId int64, ch chan PanelSettings) {
	settings := PanelSettings{
		Title: "Open A Ticket",
		Content: "React with :envelope_with_arrow: to open a ticket",
		Colour: 2335514,
	}

	Db.Where(PanelSettings{GuildId: guildId}).Take(&settings)

	ch <- settings
}
