package controller

import (
	"crypto/ecdsa"
	"math/big"
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

func DistributePapersHandler(c *gin.Context) {
	PCDistributePapers()
	c.Redirect(303, "/")
}

func PCDistributePapers() {
	users := []model.User{}
	model.GetReviewers(&users)
	var reviewerSlice []backend.Reviewer
	for _, user := range users {
		reviewerSlice = append(reviewerSlice, UserToReviewer(user))
	}
	backend.InitLocalPCPaperList()
	backend.Pc.DistributePapers(reviewerSlice, backend.Pc.AllPapers)
}

func UserToReviewer(user model.User) backend.Reviewer {
	keys := backend.DecodeToStruct(user.PublicKeys).(ecdsa.PublicKey)
	return backend.Reviewer{
		UserID: user.Id,
		Keys: &ecdsa.PrivateKey{
			PublicKey: keys,
			D:         big.NewInt(0),
		},
		PaperCommittedValue: &backend.CommitStructPaper{},
	}
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
