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

var papers []*backend.Paper
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
	backend.InitLocalPC()
	reviewerexists = true
	Kpcr := backend.GenerateSharedSecret(&backend.Pc, nil, &reviewer)
	str := fmt.Sprintf("KPCR with PC and R%v", reviewer.UserID)
	msg2 := model.Log{
		State:      99,
		LogMsg:     str,
		FromUserID: reviewer.UserID,
		Value:      backend.EncodeToBytes(Kpcr),
	}
	model.CreateLogMsg(&msg2)
}

func PaperBidHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/reviewer/bidstage.html"))
	backend.InitLocalPCPaperList()
	papers = reviewer.GetPapersReviewer(backend.Pc.AllPapers)

	// papers := []backend.Paper{
	// 	{
	// 		Id:           1,
	// 		Selected:     false,
	// 		ReviewerList: nil,
	// 		Bytes:        nil,
	// 		Title:        "Fed Titel1",
	// 	},
	// 	{
	// 		Id:           2,
	// 		Selected:     false,
	// 		ReviewerList: nil,
	// 		Bytes:        nil,
	// 		Title:        "Fed Titel2",
	// 	},
	// 	{
	// 		Id:           3,
	// 		Selected:     false,
	// 		ReviewerList: nil,
	// 		Bytes:        nil,
	// 		Title:        "Fed Titel3",
	// 	},
	// }

	tpl.Execute(c.Writer, papers)
}

func MakeBidHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/reviewer/bidstage.html"))
	c.Request.ParseForm()
	var PaperIdBids []string
	for _, value := range c.Request.PostForm {
		PaperIdBids = append(PaperIdBids, value...)
	}

	for _, v := range papers {
		for _, id := range PaperIdBids {
			idInt, err := strconv.Atoi(id)
			if err != nil {
				log.Println("error converting id string to id int")
			}
			if idInt == v.Id {
				reviewer.SignBidAndEncrypt(v)
			}
		}
	}

	var papersMatched bool

	type Bools struct {
		BidsSent      bool
		PapersMatched bool
	}

	var logmsg model.Log
	model.GetLastLogMsg(&logmsg)
	papersMatched = logmsg.State > 6
	

	bools := Bools{
		BidsSent: true,
		PapersMatched: papersMatched,
	}


	tpl.Execute(c.Writer, bools)
}

func MakeReviewHandler(c *gin.Context) {
	tpl = template.Must(template.ParseFiles("templates/reviewer/makereview.html"))
	// get data
	tpl.Execute(c.Writer, nil)
}

func DiscussingHandler(c *gin.Context) {
	tpl = template.Must(template.ParseFiles("templates/reviewer/discussing.html"))
	/*
		TODO:
		- Hent relevante log beskeder
		- Sørg for håndtering af grade input og kommentar input
	*/
	tpl.Execute(c.Writer, nil)
}
