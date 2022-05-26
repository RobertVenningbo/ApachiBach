package controller

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"swag/db"
)


type Message struct {
	Id int 				`json:"id"`
	State int 			`json:"state"`
	LogMsg string 		`json:"logmsg"`
	FromUserID big.Int 	`json:"fromuserid"`
	Value []byte 		`json:"value"`
}

func createMessage(w http.ResponseWriter, r *http.Request) {
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Panic(err)
	}
	message := Message{}
	json.Unmarshal(requestBody, &message)
	db.Conn.Create(message)

	w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(message)
}

func getMessage(w http.ResponseWriter, r *http.Request) {
	requestBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Panic(err)
	}

	message := Message{}
	json.Unmarshal(requestBody, &message)
	db.Conn.Create(message)

	w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(message)
}