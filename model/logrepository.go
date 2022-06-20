package model

import (
	"errors"
	"log"
	"swag/database"

	"gorm.io/gorm"
)

//create a logmsg
func CreateLogMsg(log *Log) (err error) {
	err = database.DB.Create(log).Error
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

func GetLogMsgByMsg(logmsg *Log, msg string) {
	result := database.DB.Where("log_msg = ?", msg).First(logmsg)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Println("Error in GetLogMsgByMsg")
	}

}

//get all logmsg
func GetAllLogMsgs(log *[]Log) (err error) {
	err = database.DB.Find(log).Error
	if err != nil {
		return err
	}
	return nil
}

//get all logmsg
func GetAllLogMsgsByMsg(logmsg *[]Log, msg string) {
	err := database.DB.Where("log_msg = ?", msg).Find(logmsg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println("Error in GetLogMsgByMsg")
	}
}

func GetLastLogMsg(log *Log) (err error) {
	err = database.DB.Last(log).Error
	if err != nil {
		return err
	}
	return nil
}
