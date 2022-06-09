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


func UpdateUsername(id int, username string) (err error) {
	err = database.DB.Model(User{}).Where("id = ?", id).Update("username", username).Error
	if err != nil {
		return err
	}
	return nil
}

func InsertPublicKeys(id int, keys []byte) (err error) {
	err = database.DB.Model(User{}).Where("id = ?", id).Update("publickeys", keys).Error
	if err != nil {
		return err
	}
	return nil
}

func GetPC(user *User) (err error){
	err = database.DB.Where("usertype = ?", "pc").First(user).Error
	if err != nil {
		return err
	}
	return nil
}

func GetReviewers(user []User) (err error){
	err = database.DB.Where("usertype = ?", "reviewer").Error
	if err != nil {
		return err
	}
	return nil
}