package controller

import (
	"sort"
	"swag/model"
	"text/template"

	"github.com/gin-gonic/gin"
)

func LogHandler(c *gin.Context) {

	var tpl = template.Must(template.ParseFiles("templates/public/log.html"))
	var logs []model.Log
	var logsView []model.Log // has to do ugly copying as GetAllLogMsgs binds to logs struct. I.e. you can't mutate.
	model.GetAllLogMsgs(&logs)

	for i := range logs {
		logsView = append(logsView, logs[i])
		logsView[i].Value = []byte{69, 69, 69, 69}
	}
	sort.SliceStable(logsView, func(i, j int) bool {
		return logsView[i].Id < logsView[j].Id
	})
	tpl.Execute(c.Writer, &logsView)
}

func TestPlatform(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/pc/decision.html"))

	tpl.Execute(c.Writer, nil)
}
