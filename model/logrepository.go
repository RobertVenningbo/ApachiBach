package model

import(
	"gorm.io/gorm"
)


//create a logmsg
func CreateLogMsg(db *gorm.DB, log *Log) (err error) {
	err = db.Create(log).Error
	if err != nil {
	   return err
	}
	return nil
 }
 
 //get all logmsg by userid
 func GetLogMsgsByUserId(db *gorm.DB, log *[]Log) (err error) {
	err = db.Find(log).Error
	if err != nil {
	   return err
	}
	return nil
 }
 
 //get logmsg by id
 func GetLogMsgById(db *gorm.DB, log *Log, id string) (err error) {
	err = db.Where("id = ?", id).First(log).Error
	if err != nil {
	   return err
	}
	return nil
 }
 
  //get all logmsg
  func GetAllLogMsgs(db *gorm.DB, log *[]Log) (err error) {
	err = db.Find(log).Error
	if err != nil {
	   return err
	}
	return nil
 }