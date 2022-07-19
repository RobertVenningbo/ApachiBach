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

func UpdatePCKeys(keys []byte) (err error) {
	err = database.DB.Model(User{}).Where("usertype = ?", "pc").Update("public_keys", keys).Error
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

func GetReviewers(user *[]User) (err error){
	err = database.DB.Where("usertype = ?", "reviewer").Find(user).Error
	if err != nil {
		return err
	}
	return nil
}

func GetSubmitters(user *[]User) (err error){
	err = database.DB.Where("usertype = ?", "submitter").Find(user).Error
	if err != nil {
		return err
	}
	return nil
}