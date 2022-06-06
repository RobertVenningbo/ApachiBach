package model

import (
	"swag/database"
)

func CreateUser(user *User) (err error) {
	err = database.DB.Create(user).Error
	if err != nil {
		return err
	}
	return nil
}


func UpdateUsername(id int, username string) {
	database.DB.Model(User{}).Where("id = ?", id).Update("username", username)
}
