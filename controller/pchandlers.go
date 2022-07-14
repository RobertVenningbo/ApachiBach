package controller

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"swag/backend"
	"swag/model"
	"text/template"

	"github.com/gin-gonic/gin"
)

var ispctaken bool
var initpaperlist bool
var sharedreviews bool

func PCHomeHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/pc/pc.html"))
	msg := CheckSubmissions()
	tpl.Execute(c.Writer, msg)
	var DBuser model.User
	model.GetPC(&DBuser)
	if ispctaken {
		return
	}
	user := model.User{}
	ispctaken = true
	keys := backend.NewKeys()
	pubkeys := backend.EncodeToBytes(keys.PublicKey)
	backend.Pc.Keys = keys

	if DBuser.Usertype == "pc" {
		fmt.Println("PC already exist in DB.")
		model.UpdatePCKeys(pubkeys)
		return
	}

	user = model.User{
		Username:   "Mr. Program Committee",
		Usertype:   "pc",
		PublicKeys: pubkeys,
	}
	model.CreateUser(&user)

}

//Helper function for PC home handler checks if all submitters have submitted
func CheckSubmissions() backend.CheckSubmissionsMessage {
	var submitters []model.User
	var logmsgs []model.Log
	var str []string
	model.GetSubmitters(&submitters)
	model.GetAllLogMsgsByState(&logmsgs, 1)
	for _, l := range logmsgs {
		paperid := strings.Split(l.LogMsg, " ")
		submitted := "Submitter " + strconv.Itoa(l.FromUserID) + " submitted paper " + paperid[1]
		str = append(str, submitted)
	}
	msg := backend.CheckSubmissionsMessage{
		SubmittersLength: len(submitters),
		Submissions:      len(logmsgs),
		Submitters:       str,
	}
	return msg

}

func BidWaitHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/pc/match_papers.html"))
	users := []model.User{}
	model.GetReviewers(&users)
	var reviewerSlice []backend.Reviewer
	for _, user := range users {
		reviewerSlice = append(reviewerSlice, backend.UserToReviewer(user))
	}
	if !initpaperlist {
		backend.InitLocalPCPaperList()
		initpaperlist = true
	}
	fmt.Printf("\n BidWaitHandler allpapers length: %v", len(backend.Pc.AllPapers))
	backend.Pc.DistributePapers(reviewerSlice, backend.Pc.AllPapers)

	data := GetAllBids()

	tpl.Execute(c.Writer, data)
}

func GetAllBids() backend.AllBids { //Helper function to check if reviewers have bidded on papers
	bidList := backend.Pc.GetAllBids()
	var users []model.User
	model.GetReviewers(&users)
	var unique []int
	m := map[int]bool{}
	for _, v := range bidList {
		if !m[v.Reviewer.UserID] {
			if v.Reviewer.UserID == -1 {
				break
			}
			m[v.Reviewer.UserID] = true
			unique = append(unique, v.Reviewer.UserID)
		}
	}
	str := ""
	showBool := false
	if len(users) == len(unique) {
		str = "All reviewers have made bids"
		showBool = true

	} else {
		str = "Not all reviewers have made bids"
		showBool = true
	}

	data := backend.AllBids{
		PaperBidCount: len(unique),
		Status:        str,
		ShowBool:      showBool,
		UsersLength:   len(users),
	}
	return data
}

func GetAllBidsHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/pc/match_papers.html"))
	data := GetAllBids()
	tpl.Execute(c.Writer, data)

}

func ShareReviewsHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/pc/share_reviews.html"))

	bidList := backend.Pc.GetAllBids()
	backend.Pc.AssignPaper(bidList)
	backend.Pc.MatchPapers()
	backend.Pc.DeliverAssignedPaper()
	messages := backend.ShareReviewsMessage{}
	if !sharedreviews {
		messages = PaperRowHelper()
		sharedreviews = true
	}

	tpl.Execute(c.Writer, &messages)
}

func ShareReviewsButtonHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/pc/share_reviews.html"))
	backend.Pc.GenerateKeysForDiscussing() // This creates all Kp for each reviewerlist, i.e. each paper.
	backend.Pc.CollectReviews()

	messages := PaperRowHelper()

	tpl.Execute(c.Writer, &messages)
}

func PaperRowHelper() backend.ShareReviewsMessage {
	messages := []backend.Message{}

	for _, p := range backend.Pc.AllPapers {
		ids := []int{}
		for _, r := range p.ReviewerList {
			ids = append(ids, r.UserID)
		}
		message := backend.Message{
			Title:       p.Title,
			ReviewerIds: ids,
		}
		messages = append(messages, message)
	}
	return backend.ShareReviewsMessage{
		Reviews: "",
		Msgs:    messages,
	}
}

func CheckReviewsHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/pc/share_reviews.html"))
	msgs := backend.ShareReviewsMessage{}
	sharedreviews = true
	var users []model.User
	model.GetReviewers(&users)
	size := len(users)
	counter := 0
	for _, p := range backend.Pc.AllPapers {
		for _, r := range p.ReviewerList {
			str := fmt.Sprintf("Reviewer, %v, finish review on paper", r.UserID)
			backend.CheckStringAgainstDB(str)
			item := backend.Trae.Find(str)
			if item != nil {
				counter++
			}
		}
	}
	msgs.Reviews = fmt.Sprintf("%v/%v reviewers have made their review.", counter, size)
	tpl.Execute(c.Writer, msgs)
}

//check up on this
func CheckConfirmedOwnerships() string {
	userlength := len(backend.AcceptedPapers)
	confirmedLength := 0

	for _, p := range backend.AcceptedPapers {
		str := fmt.Sprintf("Submitter claims paper %v by revealing paper and ri.", p.PaperId)
		var logmsgs []model.Log
		model.GetAllLogMsgsByMsg(&logmsgs, str)
		confirmedLength = len(logmsgs)
	}

	str := fmt.Sprintf("%v/%v submitters have claimed ownership of their accepted paper", confirmedLength, userlength)

	return str
}

func DecisionHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/pc/decision.html"))
	type Paper struct {
		Title string
		Grade int
		ID    int
	}
	var papers []Paper

	for _, p := range backend.Pc.AllPapers {
		GradeAndPaper := backend.Pc.GetGradeAndPaper(p.Id)
		msg := Paper{
			Title: p.Title,
			Grade: int(GradeAndPaper.GradeBefore),
			ID:    p.Id,
		}
		papers = append(papers, msg)
	}
	tpl.Execute(c.Writer, &papers)
}

func AcceptPaperHandler(c *gin.Context) {
	paperid := c.Request.FormValue("paperid")
	paperidint, err := strconv.Atoi(paperid)
	if err != nil {
		log.Println("error converting id string to id int")
		return
	}
	backend.Pc.SendGrades2(paperidint)
	backend.Pc.AcceptPaper(paperidint)

	c.Redirect(303, "/decision")
}

func RejectPaperHandler(c *gin.Context) {
	paperid := c.Request.FormValue("paperid")
	paperidint, err := strconv.Atoi(paperid)
	if err != nil {
		log.Println("error converting id string to id int")
		return
	}
	backend.Pc.SendGrades2(paperidint)
	backend.Pc.RejectPaper(paperidint)

	c.Redirect(303, "/decision")
}

func CompileGradesHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/pc/confirm_owner.html"))

	backend.Pc.CompileGrades()
	backend.Pc.RevealAllAcceptedPapers()


	type Paper struct {
		Title string
		Grade int
		ID    int
	}
	var papers []Paper

	for _, p := range backend.Pc.AllPapers {
		if backend.Pc.CheckAcceptedPapers(p.Id) {
			GradeAndPaper := backend.Pc.GetGradeAndPaper(p.Id)
			msg := Paper{
				Title: p.Title,
				Grade: int(GradeAndPaper.GradeBefore),
				ID:    p.Id,
			}
			papers = append(papers, msg)
		}
	}

	type Message struct {
		Papers []Paper
		Status string
	}

	msg := Message{
		Papers: papers,
		Status: "",
	}

	tpl.Execute(c.Writer, msg)
}
func FinishedProtocolHandler(c *gin.Context) {

	c.Redirect(303, "/postconfirmowner")
}

func ConfirmOwnershipHandler(c *gin.Context) {
	paperid := c.Request.FormValue("paperid")
	fmt.Printf("\nPaperID: %v", paperid)
	paperidint, err := strconv.Atoi(paperid)
	if err != nil {
		log.Println("\nerror converting string to id int")
		return
	}
	backend.Pc.ConfirmOwnership(paperidint)
	var tpl = template.Must(template.ParseFiles("templates/public/finished.html"))

	type Message struct {
		Status string
	}

	msg := Message{
		Status: "Confirmed ownership of all submitted papers",
	}

	tpl.Execute(c.Writer, msg)
}

func GetConfirmOwnershipHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/pc/confirm_owner.html"))
	type Paper struct {
		Title string
		Grade int
		ID    int
	}
	var papers []Paper
	for _, p := range backend.Pc.AllPapers {
		if backend.Pc.CheckAcceptedPapers(p.Id) { // this check doesn't make sense? at least not how its used. No time currently, will check up on this later
			GradeAndPaper := backend.Pc.GetGradeAndPaper(p.Id)
			msg := Paper{
				Title: p.Title,
				Grade: int(GradeAndPaper.GradeBefore),
				ID:    p.Id,
			}
			papers = append(papers, msg)
		}
	}

	type Message struct {
		Papers []Paper
		Status string
	}

	status := CheckConfirmedOwnerships()

	msg := Message{
		Papers: papers,
		Status: status,
	}
	tpl.Execute(c.Writer, msg)
}
