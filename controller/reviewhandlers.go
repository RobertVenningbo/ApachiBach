package controller

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"strconv"
	"strings"
	"swag/backend"
	"swag/model"
	"text/template"

	"github.com/gin-gonic/gin"
)

var papers []*backend.Paper //delete if possible, only used in the downloading. Maybe other way? Could use Pc.allpapers maybe idk.
var reviewer backend.Reviewer
var reviewerexists bool
var madeBid bool

func PrepStageHandler(c *gin.Context) { //this should be looked at
	var tpl = template.Must(template.ParseFiles("templates/reviewer/prepstage.html"))
	var logMsg model.Log
	model.GetLastLogMsg(&logMsg)
	type Message struct {
		Proceed bool
		Status  string
		WhereTo string
	}
	var msg Message
	shouldBid := false
	if (reviewer != backend.Reviewer{}) {
		msgs := []model.Log{}
		model.GetAllLogMsgs(&msgs)
		for _, msg := range msgs {
			if strings.Contains(msg.LogMsg, "SignedAndEncryptedPaper") {
				shouldBid = true
				break
			}
		}
	}

	if shouldBid {
		str := fmt.Sprintf("EncryptedSignedBids %v", reviewer.UserID)
		if !model.DoesLogMsgExist(str) {
			msg = Message{
				Proceed: true,
				Status:  "Continue to bid on papers.",
				WhereTo: "/paperbid",
			}
		} else {
			msg.Proceed = false
		}
	}

	if logMsg.State >= 6 {
		msg = Message{
			Proceed: true,
			Status:  "You have been assigned a paper, please continue.",
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
		UserID:              portAsInt,
		Keys:                keys,
		PaperCommittedValue: &backend.CommitStructPaper{},
	}
	model.CreateUser(&user)
	backend.InitLocalPC()
	reviewerexists = true
	Kpcr := backend.GenerateSharedSecret(&backend.Pc, nil, &reviewer)
	str := fmt.Sprintf("KPCR with PC and R%v", reviewer.UserID)
	msg2 := model.Log{
		State:      0,
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
	tpl.Execute(c.Writer, papers)
}

func WriteToFileHandler(c *gin.Context) {
	c.Request.ParseForm()
	var title string
	var downloads string
	var paperbytes []byte
	var PaperIds []string
	for _, v := range c.Request.Form {
		PaperIds = append(PaperIds, v...)
	}
	for _, p := range papers {
		for _, id := range PaperIds {
			idInt, err := strconv.Atoi(id)
			if err != nil {
				log.Println("error converting id string to id int")
			}
			if idInt == p.Id {
				paperbytes = p.Bytes

				if err != nil {
					panic(err)
				}
				currentuser, err := user.Current()
				if err != nil {
					panic(err)
				}
				downloads = currentuser.HomeDir + "/Downloads/"
				downloadsWin := currentuser.HomeDir + "\\Downloads\\"
				title = p.Title

				err = os.WriteFile(downloads+title+".pdf", paperbytes, 0644)
				if err != nil {
					log.Println("error writing file")
				}
				err = os.WriteFile(downloadsWin+title+".pdf", paperbytes, 0644)
				if err != nil {
					log.Println("error writing file")
				}
			}
		}
	}
	if madeBid {
		c.Redirect(303, "/makereview")
	} else {
		c.Redirect(303, "/paperbid")
	}

}

func DownloadRedirect(c *gin.Context) {
	if madeBid {
		c.Redirect(303, "/makereview")
	} else {
		c.Redirect(303, "/paperbid")
	}
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
				madeBid = true
			}
		}
	}

	type Bools struct {
		BidsSent      bool
		PapersMatched bool
	}

	var logmsg model.Log
	model.GetLastLogMsg(&logmsg)

	bools := Bools{
		BidsSent:      true,
		PapersMatched: logmsg.State > 6, //if the current state is gt 6 we assume papers have been matched.
	}
	c.Redirect(303, "/") //fix this pls
	tpl.Execute(c.Writer, bools)
}

func MakeReviewHandler(c *gin.Context) {
	tpl = template.Must(template.ParseFiles("templates/reviewer/makereview.html"))

	paper := reviewer.GetAssignedPaperFromPCLog()
	reviewer.PaperCommittedValue.Paper = paper

	tpl.Execute(c.Writer, paper)
}

func FinishedReviewHandler(c *gin.Context) {
	tpl = template.Must(template.ParseFiles("templates/reviewer/prepstage.html"))

	review := c.Request.FormValue("textarea_name")
	reviewer.FinishReview(review)
	reviewer.SignReviewPaperCommit()

	putStr := fmt.Sprintf("Sharing reviews with Reviewers matched to paper: %v", reviewer.PaperCommittedValue.Paper.Id)
	proceed := model.DoesLogMsgExist(putStr)
	type Message struct {
		Proceed bool
		Status  string
		WhereTo string
	}
	msg := Message{
		Proceed: proceed,
		Status:  "All reviews are now finished. Please continue to discussing.",
		WhereTo: "/discussing",
	}

	tpl.Execute(c.Writer, msg)
}

func GetFinishedReviewHandler(c *gin.Context) {
	tpl = template.Must(template.ParseFiles("templates/reviewer/prepstage.html"))

	putStr := fmt.Sprintf("Sharing reviews with Reviewers matched to paper: %v", reviewer.PaperCommittedValue.Paper.Id)
	proceed := model.DoesLogMsgExist(putStr)
	type Message struct {
		Proceed bool
		Status  string
		WhereTo string
	}
	msg := Message{
		Proceed: proceed,
		Status:  "All reviews are now finished. Please continue to discussing.",
		WhereTo: "/discussing",
	}

	tpl.Execute(c.Writer, msg)
}

func DiscussingHandler(c *gin.Context) {
	tpl = template.Must(template.ParseFiles("templates/reviewer/discussing.html"))
	data := reviewer.GetSecretMsgsFromReviewers() //retrieves messages for same paper as the reviewer itself and decrypts with Kp etc etc
	tpl.Execute(c.Writer, data)
}

func PostMessageDiscussingHandler(c *gin.Context) {
	discussingMessage := c.Request.FormValue("textarea_name")
	str := fmt.Sprintf("Reviewer %v: %s", reviewer.UserID, discussingMessage)
	reviewer.SendSecretMsgToReviewers(str)
	c.Redirect(303, "/discussing") //cheesy way of refreshing gui
}

func GetGradeDiscussingHandler(c *gin.Context) {
	tpl = template.Must(template.ParseFiles("templates/reviewer/prepstage.html"))

	type Message struct {
		Proceed bool
		Status  string
		WhereTo string
	}
	str := fmt.Sprintf("All grades have been submitted for Paper: %v", reviewer.PaperCommittedValue.Paper.Id)
	if !model.DoesLogMsgExist(str) {
		reviewer.PublishAgreedGrade()
	}
	willProceed := false
	if model.DoesLogMsgExist(str) {
		willProceed = true
	}
	msg := Message{
		Proceed: willProceed,
		Status:  "All grades are now submitted. Please continue.",
		WhereTo: "/signgradecommit",
	}

	tpl.Execute(c.Writer, msg)
}

func PostGradeDiscussingHandler(c *gin.Context) {
	tpl = template.Must(template.ParseFiles("templates/reviewer/prepstage.html"))

	c.Request.ParseForm()
	individualgrade := c.Request.FormValue("grade_name")
	gradeAsInt, err := strconv.Atoi(individualgrade)
	if err != nil {
		log.Println("Error converting grade to int")
	}
	reviewer.GradePaper(gradeAsInt)
	type Message struct {
		Proceed bool
		Status  string
		WhereTo string
	}
	str := fmt.Sprintf("All grades have been submitted for Paper: %v", reviewer.PaperCommittedValue.Paper.Id)

	exists := model.DoesLogMsgExist(str)
	if !exists {
		reviewer.PublishAgreedGrade()
	}

	c.Redirect(303, "/submitgrade")
}

func GetAgreedGradeHandler(c *gin.Context) {
	tpl = template.Must(template.ParseFiles("templates/reviewer/avggrade.html"))

	gradeStruct := reviewer.GetAgreedGroupGrade()

	type Msg struct {
		Title string
		Grade int
	}
	msg := Msg{
		Title: reviewer.PaperCommittedValue.Paper.Title,
		Grade: int(gradeStruct.GradeBefore),
	}

	if (gradeStruct == backend.RandomizeGradesForProofStruct{}) {
		c.Redirect(303, "/signgradecommit")
	}

	tpl.Execute(c.Writer, msg)
}

func SignGradeHandler(c *gin.Context) {
	tpl = template.Must(template.ParseFiles("templates/public/finished.html"))

	type Message struct {
		Status string
	}

	msg := Message{
		Status: "Thank you for participating as a reviewer! 📝😉 ",
	}

	reviewer.SignCommitsAndNonce()
	reviewer.SignAndEncryptGrade()

	tpl.Execute(c.Writer, msg)
}
