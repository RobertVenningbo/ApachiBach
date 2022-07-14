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

//get a single log entry by a string
func GetLogMsgByMsg(logmsg *Log, msg string) {
	result := database.DB.Where("log_msg = ?", msg).First(logmsg)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Println("Error in GetLogMsgByMsg")
	}
}

//get all log entries
func GetAllLogMsgs(log *[]Log) (err error) {
	err = database.DB.Select("id, state, log_msg, from_user_id, value, signature").Find(log).Error
	if err != nil {
		return err
	}
	return nil
}

//Retrieves the entire column "log_msg"
func GetAllLogMsgsLogMsgs(logmsg *[]Log, msg string) { 
	err := database.DB.Select("log_msg").Where("log_msg = ?", msg).Find(logmsg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println("Error in GetAllLogMsgsLogMsgs")
	}
	
}

func GetAllLogMsgsFromSender(logmsg *[]Log, id int) {
	err := database.DB.Select("from_user_id").Where("from_user_id = ?", id).Find(&logmsg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println("Error in GetAllLogMsgsFromSender")
	}
}

func GetAllLogMsgsByState(logmsg *[]Log, state int) {
	err := database.DB.Select("state, log_msg, from_user_id, value, signature").Where("state = ?", state).Find(&logmsg).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println("Error in GetAllLogMsgsByState")
	}
}

//Checks if log msg exists in the database
func DoesLogMsgExist(msg string) bool {
	var exists bool
	err := database.DB.Model(Log{}).Select("count(*) > 0").Where("log_msg = ?", msg).Find(&exists).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Println("Error in DoesLogMsgExist")
	}
	return exists

}

//get all log entries by a string
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
