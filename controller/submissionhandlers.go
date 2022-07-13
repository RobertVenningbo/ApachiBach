package controller

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"strconv"
	"strings"
	"swag/backend"
	"swag/model"
	"text/template"
	"time"
	"github.com/gin-gonic/gin"
)

var tpl = template.Must(template.ParseFiles("templates/submitter/submission.html"))
var submitter backend.Submitter

func SubmissionHandler(c *gin.Context) {
	url := strings.Split(c.Request.Host, ":")
	portAsInt, _ := strconv.Atoi(url[1])

	keys := backend.NewKeys()
	pubkeys := backend.EncodeToBytes(keys.PublicKey)

	user := model.User{
		Id:         portAsInt,
		Username:   "",
		Usertype:   "submitter",
		PublicKeys: pubkeys,
	}

	model.CreateUser(&user)

	submitter = backend.Submitter{
		Keys:                    keys,
		UserID:                  portAsInt, //userID
		SubmitterCommittedValue: &backend.CommitStruct{},
		PaperCommittedValue:     &backend.CommitStructPaper{},
	}

	tpl.Execute(c.Writer, nil)
}

func GetGradesAndReviewsHandler(c *gin.Context) {
	tpl = template.Must(template.ParseFiles("templates/submitter/paper_graded.html"))
	gradeandreviews := submitter.RetrieveGradeAndReviews()
	type Message struct {
		Status string
		Grade int
		Reviews []backend.ReviewStruct
	}

	str_rejected := fmt.Sprintf("PC rejects Paper: %v", submitter.PaperCommittedValue.Paper.Id)
	str_accepted := fmt.Sprintf("PC reveals accepted paper: %v", submitter.PaperCommittedValue.Paper.Id)

	var msg Message
	if model.DoesLogMsgExist(str_rejected) {
		msg = Message{
			Status: "Rejected",
			Grade: gradeandreviews.Grade,
			Reviews: gradeandreviews.Reviews,
		}
	} else if model.DoesLogMsgExist(str_accepted) {
		msg = Message{
			Status: "Accepted",
			Grade: gradeandreviews.Grade,
			Reviews: gradeandreviews.Reviews,
		}
	}
	tpl.Execute(c.Writer, &msg)
}

func WaitHandler(c *gin.Context) {
	type Message struct {
		Status  string
		Cont    bool
	}
	var msg Message

	msg = Message{
		Status: "Pending.",
		Cont: false,
	}
	
	str_rejected := fmt.Sprintf("PC rejects Paper: %v", submitter.PaperCommittedValue.Paper.Id)
	str_accepted := fmt.Sprintf("PC reveals accepted paper: %v", submitter.PaperCommittedValue.Paper.Id)
	if model.DoesLogMsgExist(str_rejected) || model.DoesLogMsgExist(str_accepted) {
		msg = Message{
			Status: "Your paper has been graded.",
			Cont: true,
		}
	}
	
	tpl = template.Must(template.ParseFiles("templates/submitter/you_have_submitted.html"))
	tpl.Execute(c.Writer, &msg)
}

func GradedPaperHandler(c *gin.Context) { //delete

	type Reviewer struct {
	}
	type Message struct {
		Status string
		Grade  int
		Count  []Reviewer
	}
	reviewers := []Reviewer{
		{},
		{},
	}
	msg := Message{
		Status: "pending...",
		Grade:  4,
		Count:  reviewers,
	}

	tpl = template.Must(template.ParseFiles("templates/submitter/paper_graded.html"))
	tpl.Execute(c.Writer, &msg)
}

func UploadFile(c *gin.Context) {
	fmt.Println("File Upload Endpoint Hit")
	// Parse our multipart form, 10 << 20 specifies a maximum upload of 10 MB files.
	c.Request.ParseMultipartForm(10 << 20)

	//name := c.Request.FormValue("name")
	title := c.Request.FormValue("title")
	file, handler, err := c.Request.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}


	source := rand.NewSource(time.Now().UnixNano())
	rand := rand.New(source)
	paperid := rand.Intn(99999-10000)

	paper := backend.Paper{
		Id:    paperid, 
		Bytes: fileBytes,
		Title: title,
	}
	


	submitter.Submit(&paper)

	c.Redirect(303, "/wait")

}

func ClaimPaperHandler(c *gin.Context) {
	tpl = template.Must(template.ParseFiles("templates/public/finished.html")) 

	type Message struct {
		Status  string
	}

	msg := Message{
		Status:  "Thank you for submitting a paper! ",
	}

	
	submitter.ClaimPaper(submitter.PaperCommittedValue.Paper.Id)

	tpl.Execute(c.Writer, msg)
	
}
