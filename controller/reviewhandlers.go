package controller

import (
	"log"
	"swag/model"
	"text/template"

	"github.com/gin-gonic/gin"
)

func PrepStageHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/reviewer/prepstage.html"))
	var logMsg model.Log
	err := model.GetLastLogMsg(&logMsg)
	if err != nil {
		log.Fatal("PrepStageHandler failed")
		return
	}

	proceed := false
	if logMsg.State > 3 {
		proceed = true
	}
	type Message struct {
		Proceed bool
	}
	msg := Message{
		Proceed: proceed,
	}
	tpl.Execute(c.Writer, msg)
}
