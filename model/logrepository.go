package model

import (
	"errors"
	"swag/database"
	"log"
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

// dbRresult := userHandler.db.Where("email = ?", email).First(&user)
// if errors.Is(dbRresult.Error, gorm.ErrRecordNotFound) {
//     // handle record not found
// }

func GetLogMsgByMsg(logmsg *Log, msg string) {
	result := database.DB.Where("log_msg = ?", msg).First(logmsg)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		log.Println("Error in GetLogMsgByMsg")
		return
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

func GetLastLogMsg(log *Log) (err error) {
	err = database.DB.Last(log).Error
	if err != nil {
		return err
	}
	return nil
}
