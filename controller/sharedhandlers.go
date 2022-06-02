package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"swag/model"
	"text/template"

	"github.com/gin-gonic/gin"
)

func LogHandler(c *gin.Context) {

	var tpl = template.Must(template.ParseFiles("templates/log.html"))
	var logs []model.Log
	data, err := http.Get("http://127.0.0.1:2533/v1/api/logmsg")
	if err != nil {
		log.Fatal("err in logHandler")
	}
	// if data.StatusCode != http.StatusOK {
	// 	return
	// }
	bodyBytes, err := io.ReadAll(data.Body)
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(bodyBytes, &logs)
	fmt.Printf("%#v", logs)
	tpl.Execute(c.Writer, logs)
}

func (h handler) CreateMessage(c *gin.Context) {
	var msg model.Log
	c.BindJSON(&msg)
	err := model.CreateLogMsg(h.DB, &msg)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusCreated, msg)
}

func (h handler) GetMessages(c *gin.Context) {
	var msgs []model.Log
	c.BindJSON(msgs)
	err := model.GetAllLogMsgs(h.DB, &msgs)
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
	err := model.GetLogMsgById(h.DB, &msg, id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, msg)
}

func createMessage_notGIN(w http.ResponseWriter, r *http.Request) {
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Panic(err)
	}
	message := model.Log{}
	json.Unmarshal(requestBody, &message)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(message)
}

func getMessage(w http.ResponseWriter, r *http.Request) {

}
