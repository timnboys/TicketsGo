package database

type DmOnOpen struct {
	Guild    uint64 `gorm:"column:GUILDID;unique;primary_key"`
	DmOnOpen *bool  `gorm:"column:DM_ON_OPEN"`
}

func (DmOnOpen) TableName() string {
	return "DmOnOpen"
}

func IsDmOnOpen(guild uint64, ch chan bool) {
	var row DmOnOpen
	Db.Where(DmOnOpen{Guild: guild}).First(&row)

	if row.DmOnOpen == nil {
		ch <- false
	} else {
		ch <- *row.DmOnOpen
	}
}
