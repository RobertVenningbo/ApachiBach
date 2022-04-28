package backend_test

import (
	"testing"
	. "swag/backend"

	"github.com/stretchr/testify/assert"
)

func TestSendGrades_And_GetGrade(t *testing.T) {
	submitter.Submit(&p)
	reviewerSlice := []*Reviewer{&reviewer}
	reviewer.SignBidAndEncrypt(&p)
	Pc.AssignPaper(reviewerSlice)
	Pc.MatchPapers()                           //step 7
	reviewer.FinishReview("Pretty rad paper!") //step 8
	reviewer.SignReviewPaperCommit()           //step 9
	Pc.GenerateKeysForDiscussing([]Reviewer{reviewer})
	Pc.CollectReviews(p.Id) //step 11
	reviewer.GradePaper(7)
	reviewer.SignReviewPaperCommit()
	reviewer.SignCommitsAndNonce()
	reviewer.SignAndEncryptGrade()

	Pc.SendGrades(&submitter)
	sendGradeStructActual := submitter.RetrieveGrades()
	sendGradeStructExpected := SendGradeStruct{
		[]string{"Pretty rad paper!"},
		7,
	}

	assert.Equal(t, sendGradeStructExpected, sendGradeStructActual, "failz")
}

func TestRejectPaper(t *testing.T) {
	submitter.Submit(&p)
	reviewerSlice := []*Reviewer{&reviewer}
	reviewer.SignBidAndEncrypt(&p)
	Pc.AssignPaper(reviewerSlice)
	Pc.MatchPapers()                           //step 7
	reviewer.FinishReview("Pretty rad paper!") //step 8
	reviewer.SignReviewPaperCommit()           //step 9
	Pc.GenerateKeysForDiscussing([]Reviewer{reviewer})
	Pc.CollectReviews(p.Id) //step 11
	reviewer.GradePaper(2)
	reviewer.SignReviewPaperCommit()
	reviewer.SignCommitsAndNonce()
	reviewer.SignAndEncryptGrade()

	Pc.SendGrades(&submitter)
	actual := Pc.RejectPaper(p.Id)
	KpAndRg := Pc.GetKpAndRgPC(p.Id)
	expected := RejectMessage{
		Pc.GetReviewSignedStruct(p.Id).Commit,
		Pc.GetGrade(p.Id),
		KpAndRg.Rg,
	}

	assert.Equal(t, expected, actual, "TestRejectPaper failed")
}

func TestGetCompiledGrades_And_Get(t *testing.T) {
	submitter.Submit(&p)
	reviewerSlice := []*Reviewer{&reviewer}
	reviewer.SignBidAndEncrypt(&p)
	Pc.AssignPaper(reviewerSlice)
	Pc.MatchPapers()                           //step 7
	reviewer.FinishReview("Pretty rad paper!") //step 8
	reviewer.SignReviewPaperCommit()           //step 9
	Pc.GenerateKeysForDiscussing([]Reviewer{reviewer})
	Pc.CollectReviews(p.Id) //step 11
	reviewer.GradePaper(2)
	reviewer.SignReviewPaperCommit()
	reviewer.SignCommitsAndNonce()
	reviewer.SignAndEncryptGrade()
	
	Pc.AcceptPaper(p.Id)
	Pc.CompileGrades()
	actual := Pc.GetCompiledGrades()
	Pc.RevealAcceptedPaperInfo(p.Id)
	expected := []int64{2}

	assert.Equal(t, expected, actual, "failz")
}

func TestRevealAcceptedPaperInfo(t *testing.T) {
	submitter.Submit(&p)
	reviewerSlice := []*Reviewer{&reviewer}
	reviewer.SignBidAndEncrypt(&p)
	Pc.AssignPaper(reviewerSlice)
	Pc.MatchPapers()                           //step 7
	reviewer.FinishReview("Pretty rad paper!") //step 8
	reviewer.SignReviewPaperCommit()           //step 9
	Pc.GenerateKeysForDiscussing([]Reviewer{reviewer})
	Pc.CollectReviews(p.Id) //step 11
	reviewer.GradePaper(2)
	reviewer.SignReviewPaperCommit()
	reviewer.SignCommitsAndNonce()
	reviewer.SignAndEncryptGrade()
	
	Pc.AcceptPaper(p.Id)
	Pc.CompileGrades()
	actual := Pc.RevealAcceptedPaperInfo(p.Id)
	p := Pc.GetPaperAndRandomness(p.Id)
	expected := RevealPaper{
		*p.Paper,
		p.Rs,
	}
	assert.Equal(t, expected, actual, "failz")
}

func TestXxx4(t *testing.T) {
	//it's here if u wanna add a test
	assert.Equal(t, true, true, "failz")
}

