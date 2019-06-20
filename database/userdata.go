package database

type UserData struct {
	Id          int64  `gorm:"column:USERID;unique;primary_key"`
	Username      string `gorm:"column:USERNAME;type:text"`
	Discriminator string `gorm:"column:DISCRIM;type:varchar(4)"`
	Avatar        string `gorm:"column:AVATARHASH;type:varchar(100)"`
}

func (UserData) TableName() string {
	return "usernames"
}

func UpdateUser(id int64, name string, discrim string, avatarHash string) {
	var node UserData
	Db.Where(UserData{Id: id}).Assign(&UserData{Username: name, Discriminator: discrim, Avatar: avatarHash}).FirstOrCreate(&node)
}
