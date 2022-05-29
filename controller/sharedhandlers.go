package controller

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"swag/model"

	"github.com/gin-gonic/gin"
)

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
	c.BindJSON(&msgs)
	err := model.GetAllLogMsgs(h.DB, &msgs)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, msgs)
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
