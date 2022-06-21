package controller

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"swag/backend"
	"swag/model"
	"text/template"

	"github.com/gin-gonic/gin"
)

var tpl = template.Must(template.ParseFiles("templates/submitter/submission.html"))
var submitter backend.Submitter

func SubmissionHandler(c *gin.Context) {
	tpl.Execute(c.Writer, nil)
}

func WaitHandler(c *gin.Context) {
	//TODO get data
	//retrieve latest message from the log, check its state and depending on
	//the state you change a string saying pending, ok, error or something along those lines
	type Message struct {
		Msg  string
		Cont bool
	}
	msg := Message{
		Msg:  "pending...",
		Cont: false, //true for button, just trying some frontend logic
	}
	tpl = template.Must(template.ParseFiles("templates/submitter/you_have_submitted.html"))
	tpl.Execute(c.Writer, &msg)
}

func GradedPaperHandler(c *gin.Context) {

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
	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	c.Request.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file

	name := c.Request.FormValue("name")
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



	url := strings.Split(c.Request.Host, ":")
	portAsInt, _ := strconv.Atoi(url[1])

	keys := backend.NewKeys()
	pubkeys := backend.EncodeToBytes(keys.PublicKey)

	user := model.User{
		Id:         portAsInt,
		Username:   name,
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

	paper := backend.Paper{
		Id:    portAsInt,
		Bytes: fileBytes,
		Title: title,
	}

	submitter.Submit(&paper)

	c.Redirect(303, "/wait")

}
