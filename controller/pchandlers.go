package controller

import (
	"fmt"
	"log"
	"strconv"
	"swag/backend"
	"swag/model"
	"text/template"

	"github.com/gin-gonic/gin"
)

var ispctaken bool

//var AcceptedPapers []backend.Paper

func PCHomeHandler(c *gin.Context) {

	var tpl = template.Must(template.ParseFiles("templates/pc/pc.html"))
	tpl.Execute(c.Writer, nil)
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

func BidWaitHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/pc/match_papers.html"))
	users := []model.User{}
	model.GetReviewers(&users)
	var reviewerSlice []backend.Reviewer
	for _, user := range users {
		reviewerSlice = append(reviewerSlice, UserToReviewer(user))
	}
	backend.InitLocalPCPaperList()
	fmt.Printf("\n BidWaitHandler allpapers length: %v", len(backend.Pc.AllPapers))
	backend.Pc.DistributePapers(reviewerSlice, backend.Pc.AllPapers)

	tpl.Execute(c.Writer, nil)
}

func GetAllBidsHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/pc/match_papers.html"))
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

	type AllBids struct {
		PaperBidCount int
		Status        string
		ShowBool      bool
		UsersLength   int
	}

	if len(users) == len(unique) {
		str = "All reviewers have made bids"
		showBool = true

	} else {
		str = "Not all reviewers have made bids"
		showBool = true
	}

	data := AllBids{
		PaperBidCount: len(unique),
		Status:        str,
		ShowBool:      showBool,
		UsersLength:   len(users),
	}

	tpl.Execute(c.Writer, data)

}

func ShareReviewsHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/pc/share_reviews.html"))

	bidList := backend.Pc.GetAllBids()
	backend.Pc.AssignPaper(bidList)
	backend.Pc.MatchPapers()
	backend.Pc.DeliverAssignedPaper()

	messages := PaperRowHelper()

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
	msgs := PaperRowHelper()
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

func CheckConfirmedOwnerships() string{
	var users []model.User
	model.GetSubmitters(&users)
	userlength := len(users)
	confirmedLength := 0

	for _, p := range backend.AcceptedPapers {
		str := fmt.Sprintf("Submitter claims paper %v by revealing paper and ri.", p.PaperId)
		var logmsgs []model.Log
		model.GetAllLogMsgsByMsg(&logmsgs, str)
		confirmedLength = len(logmsgs)
	}

	str := fmt.Sprintf("%v/%v submitters have claimed ownership of their paper", confirmedLength, userlength)

	return str
}

func CheckConfirmedOwnershipHandler(c *gin.Context) {

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
		Papers	[]Paper
		Status  string
	}

	msg := Message{
		Papers: papers,
		Status: "",
	}

	tpl.Execute(c.Writer, msg)
}

func ConfirmOwnershipHandler(c *gin.Context) {
	paperid := c.Request.FormValue("paperid")
	paperidint, err := strconv.Atoi(paperid)
	if err != nil {
		log.Println("error converting id string to id int")
		return
	}

	users := []model.User{}
	model.GetSubmitters(&users)
	var submitterSlice []backend.Submitter
	for _, user := range users {
		submitterSlice = append(submitterSlice, UserToSubmitter(user))
	}

	backend.Pc.ConfirmOwnership(paperidint)
	c.Redirect(303, "/confirmowner")
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
		Papers	[]Paper
		Status  string
	}

	status := CheckConfirmedOwnerships()

	msg := Message{
		Papers: papers,
		Status: status,
	}
	tpl.Execute(c.Writer, msg)

}
