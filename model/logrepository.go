package model

import (
	"swag/database"
)

//create a logmsg
func CreateLogMsg(log *Log) (err error) {
	err = database.DB.Create(log).Error
	if err != nil {
		return err
	}
	return nil
}

//get all logmsg by userid
func GetLogMsgsByUserId(log *[]Log) (err error) {
	err = database.DB.Find(log).Error
	if err != nil {
		return err
	}
	return nil
}

//get logmsg by id
func GetLogMsgById(log *Log, id string) (err error) {
	err = database.DB.Where("id = ?", id).First(log).Error
	if err != nil {
		return err
	}
	return nil
}

//get all logmsg
func GetAllLogMsgs(log *[]Log) (err error) {
	err = database.DB.Find(log).Error
	if err != nil {
		return err
	}
	return nil
}
