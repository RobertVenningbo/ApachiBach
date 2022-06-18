package backend

import (
	"swag/model"
)

var tree = *Trae

func DatabaseToTree() *Tree {
	var msgs []model.Log
	model.GetAllLogMsgs(&msgs)

	for _, msg := range msgs {
		tree.Put(msg.LogMsg, msg.Value)
	}

	return &tree
}

func CheckStringAgainstDB(str string) {
	var msg model.Log
	model.GetLogMsgByMsg(&msg, str)

	// if err != nil {
	// 	log.Printf("String not found in Database \n")
	// 	return
	// }

	Trae.Put(msg.LogMsg, msg.Value)
}

func CheckStringAgainstDBStruct(str string) {
	var msg model.Log
	model.GetLogMsgByMsg(&msg, str)

	msglog := ValueSignature{
		Value:     msg.Value,
		Signature: msg.Signature,
	}

	Trae.Put(msg.LogMsg, EncodeToBytes(msglog))
}
