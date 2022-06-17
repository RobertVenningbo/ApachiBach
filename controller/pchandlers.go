package controller

import (
	"swag/backend"
	"swag/model"
	"text/template"

	"github.com/gin-gonic/gin"
)

var ispctaken bool

func PCHomeHandler(c *gin.Context) {

	var tpl = template.Must(template.ParseFiles("templates/pc/pc.html"))
	tpl.Execute(c.Writer, nil)
	//var DBuser model.User
	//model.GetPC(&DBuser)
	if ispctaken {
		return
	}
	user := model.User{}
	ispctaken = true
	keys := backend.NewKeys()
	pubkeys := backend.EncodeToBytes(keys.PublicKey)
	backend.Pc.Keys = keys

	// if DBuser.Usertype == "pc" {
	// 	fmt.Println("PC already exist in DB.")
	// 	model.UpdatePCKeys(pubkeys)
	// 	return
	// }

	user = model.User{
		Username:   "Mr. Program Committee",
		Usertype:   "pc",
		PublicKeys: pubkeys,
	}
	model.CreateUser(&user)
}

// func DistributePapersHandler(c *gin.Context) {
// 	PCDistributePapers()
// 	c.Redirect(303, "/")
// }
func GetBidsHandler(c *gin.Context) {

	c.Redirect(303, "/")
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
		UsersLength	  int
	}

	if len(users) == len(unique) {
		str = "All reviewers have made bids"
		showBool = true

	} else {
		str = "Not all reviewers have made bids"
		showBool = true
	} 

	blabla := AllBids{
		PaperBidCount: len(unique),
		Status:        str,
		ShowBool:      showBool,
		UsersLength:   len(users),
	}

	tpl.Execute(c.Writer, blabla)

}

func DecisionHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/pc/decision.html"))

	type Paper struct {
		Title string
		Grade int
	}
	var paper []Paper
	paper = append(paper, Paper{
		Title: "test1",
		Grade: 4,
	},
		Paper{
			Title: "test2",
			Grade: 7})

	tpl.Execute(c.Writer, &paper)
}

func ShareReviewsHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/pc/share_reviews.html"))

	bidList := backend.Pc.GetAllBids()
	backend.Pc.AssignPaper(bidList)
	type Reviewer struct {
		User string
	}
	type Message struct {
		Title     string
		Reviewers []Reviewer
	}
	reviewers := []Reviewer{
		{"reviewer1"},
		{"reviewer2"},
	}
	msg := []Message{
		{
			Title:     "Title1",
			Reviewers: reviewers,
		},
		{
			Title:     "Title2",
			Reviewers: reviewers,
		},
	}
	tpl.Execute(c.Writer, &msg)
}
