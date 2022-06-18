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
var paper *backend.Paper
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
	proceedToMatch := false
	proceedToBid := false
	type Message struct {
		Proceed bool
		Status  string
		WhereTo string
	}
	var msg Message
	if logMsg.State == 4  {
		proceedToBid = true
		msg = Message{
			Proceed: proceedToBid,
			Status: "Continue to bid on papers.",
			WhereTo: "/paperbid",

		}
	}
	if logMsg.State > 6 {
		proceedToBid = false
		proceedToMatch = true
		msg = Message {
			Proceed: proceedToMatch,
			Status: "You have been assigned a paper, please continue.",
			WhereTo: "/makereview",
		}
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
		PaperCommittedValue: &backend.CommitStructPaper{},
	}
	model.CreateUser(&user)
	backend.InitLocalPC()
	reviewerexists = true
	Kpcr := backend.GenerateSharedSecret(&backend.Pc, nil, &reviewer)
	str := fmt.Sprintf("KPCR with PC and R%v", reviewer.UserID)
	msg2 := model.Log{
		State:      3, //This needs to be something else
		LogMsg:     str,
		FromUserID: reviewer.UserID,
		Value:      backend.EncodeToBytes(Kpcr),
	}
	model.CreateLogMsg(&msg2)
}

func PaperBidHandler(c *gin.Context) { //TODO: Implement a way to refresh without adding the same paper to the paper list
	var tpl = template.Must(template.ParseFiles("templates/reviewer/bidstage.html"))

	backend.InitLocalPCPaperList()
	papers = reviewer.GetPapersReviewer(backend.Pc.AllPapers)
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
		BidsSent:      true,
		PapersMatched: papersMatched,
	}
	c.Redirect(303, "/")
	tpl.Execute(c.Writer, bools)
}

func MakeReviewHandler(c *gin.Context) {
	tpl = template.Must(template.ParseFiles("templates/reviewer/makereview.html"))

	paper = reviewer.GetAssignedPaperFromPCLog()
	fmt.Println(paper.Id)
	reviewer.PaperCommittedValue.Paper = paper

	tpl.Execute(c.Writer, paper)
}

func FinishedReviewHandler(c *gin.Context) {
	tpl = template.Must(template.ParseFiles("templates/reviewer/prepstage.html"))

	review := c.Request.FormValue("textarea_name")
	reviewer.FinishReview(review)
	reviewer.SignReviewPaperCommit()
	fmt.Println(review)

	logmsg := model.Log{}
	model.GetLastLogMsg(&logmsg)
	
	type Message struct {
		Proceed bool
		Status  string
		WhereTo string
	}
	msg := Message{
		Proceed: logmsg.State == 12, //change later
		Status:	 "All reviews are now finished. Please continue to discussing.",
		WhereTo: "/discussing",
	}

	tpl.Execute(c.Writer, msg)

}

func GetFinishedReviewHandler(c *gin.Context) {
	tpl = template.Must(template.ParseFiles("templates/reviewer/prepstage.html"))

	logmsg := model.Log{}
	model.GetLastLogMsg(&logmsg)
	
	type Message struct {
		Proceed bool
		Status  string
		WhereTo string
	}
	msg := Message{
		Proceed: logmsg.State == 12, //change later
		Status:	 "All reviews are now finished. Please continue to discussing.",
		WhereTo: "/discussing",
	}

	tpl.Execute(c.Writer, msg)

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
