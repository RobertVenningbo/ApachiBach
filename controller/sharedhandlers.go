package controller

import (
	"net/http"
	"swag/model"
	"text/template"

	"github.com/gin-gonic/gin"
)

func LogHandler(c *gin.Context) {

	var tpl = template.Must(template.ParseFiles("templates/log.html"))
	var logs []model.Log
	// do something with different the getting method.
	model.GetAllLogMsgs(&logs)

	tpl.Execute(c.Writer, logs)
}

func (h handler) CreateMessage(c *gin.Context) {
	var msg model.Log
	c.BindJSON(&msg)
	err := model.CreateLogMsg(&msg)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusCreated, msg)
}

func (h handler) GetMessages(c *gin.Context) {
	var msgs []model.Log
	c.BindJSON(msgs)
	err := model.GetAllLogMsgs(&msgs)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, msgs)
}

func (h handler) GetMessage(c *gin.Context) {
	id, _ := c.Params.Get("id")
	var msg model.Log
	c.BindJSON(msg)
	err := model.GetLogMsgById(&msg, id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, msg)
}

func TestPlatform(c *gin.Context) {
	tpl = template.Must(template.ParseFiles("templates/reviewer/makereview.html"))

	tpl.Execute(c.Writer, nil)
}
