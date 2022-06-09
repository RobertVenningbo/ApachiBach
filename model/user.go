package model

type User struct {
	Id         int    `json:"id" gorm:"primarykey"`
	Username   string `json:"username"`
	Usertype   string `json:"usertype"`
	PublicKeys []byte `json:"publickeys"`
}
