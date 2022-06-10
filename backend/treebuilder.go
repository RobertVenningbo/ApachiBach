package backend

import (
	"log"
	"swag/model"
)

var tree = *Trae

func DatabaseToTree() (*Tree) {
	var msgs []model.Log
	model.GetAllLogMsgs(&msgs)

	for _, msg := range msgs {
		tree.Put(msg.LogMsg, msg.Value)
	}

	return &tree
}

func CheckStringAgainstDB(str string) {
	var msg model.Log
	err := model.GetLogMsgByMsg(&msg, str)
	
	if err != nil {
		log.Fatalf("String not found in Database")
		return
	}
	
	Trae.Put(msg.LogMsg, msg.Value)
}