package controller

import (
	"text/template"

	"github.com/gin-gonic/gin"
)

func PCHomeHandler(c *gin.Context) {
	var tpl = template.Must(template.ParseFiles("templates/pc/pc.html"))
	tpl.Execute(c.Writer, nil)
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
