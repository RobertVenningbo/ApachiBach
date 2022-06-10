package controller

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"swag/backend"
	"swag/model"
	"text/template"

	"github.com/gin-gonic/gin"
)

var reviewer backend.Reviewer
var reviewerexists bool

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

	if reviewerexists {
		return
	}
	fmt.Println("tjek")
	url := strings.Split(c.Request.Host, ":")
	portAsInt, _ := strconv.Atoi(url[1])

	keys := backend.NewKeys()
	pubkeys := backend.EncodeToBytes(keys.PublicKey)
	user := model.User{
		Id:         portAsInt,
		Usertype:   "reviewer",
		PublicKeys: pubkeys,
	}
	reviewer = backend.Reviewer{
		UserID: portAsInt,
		Keys:   keys,
	}

	model.CreateUser(&user)
	reviewerexists = true
}

func PaperBidHandler(c *gin.Context) {
	backend.InitLocalPC()
	var tpl = template.Must(template.ParseFiles("templates/reviewer/bidstage.html"))
	backend.InitLocalPCPaperList()
	papers := reviewer.GetPapersReviewer(backend.Pc.AllPapers)
	fmt.Println(len(papers))
	tpl.Execute(c.Writer, papers)
}
