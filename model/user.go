package model

type User struct {
	Id         int    `json:"id" gorm:"primarykey"`
	State      int    `json:"state"`
	LogMsg     string `json:"logmsg"`
	FromUserID int    `json:"fromuserid"`
	Value      []byte `json:"value"`
}
