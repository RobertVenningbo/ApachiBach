package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSendGrades_And_GetGrade(t *testing.T) {
	submitter.Submit(&p)
	reviewerSlice := []*Reviewer{&reviewer}
	reviewer.SignBidAndEncrypt(&p)
	pc.assignPaper(reviewerSlice)
	pc.MatchPapers()                           //step 7
	reviewer.FinishReview("Pretty rad paper!") //step 8
	reviewer.SignReviewPaperCommit()           //step 9
	pc.GenerateKeysForDiscussing([]Reviewer{reviewer})
	pc.CollectReviews(p.Id) //step 11
	reviewer.GradePaper(7)
	reviewer.SignReviewPaperCommit()
	reviewer.SignCommitsAndNonce()
	reviewer.SignAndEncryptGrade()

	pc.SendGrades(&submitter)
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
	pc.assignPaper(reviewerSlice)
	pc.MatchPapers()                           //step 7
	reviewer.FinishReview("Pretty rad paper!") //step 8
	reviewer.SignReviewPaperCommit()           //step 9
	pc.GenerateKeysForDiscussing([]Reviewer{reviewer})
	pc.CollectReviews(p.Id) //step 11
	reviewer.GradePaper(2)
	reviewer.SignReviewPaperCommit()
	reviewer.SignCommitsAndNonce()
	reviewer.SignAndEncryptGrade()

	pc.SendGrades(&submitter)
	actual := pc.RejectPaper(p.Id)
	KpAndRg := pc.GetKpAndRgPC(p.Id)
	expected := RejectMessage{
		pc.GetReviewSignedStruct(p.Id).Commit,
		pc.GetGrade(p.Id),
		KpAndRg.Rg,
	}

	assert.Equal(t, expected, actual, "TestRejectPaper failed")
}

func TestGetCompiledGrades_And_Get(t *testing.T) {
	submitter.Submit(&p)
	reviewerSlice := []*Reviewer{&reviewer}
	reviewer.SignBidAndEncrypt(&p)
	pc.assignPaper(reviewerSlice)
	pc.MatchPapers()                           //step 7
	reviewer.FinishReview("Pretty rad paper!") //step 8
	reviewer.SignReviewPaperCommit()           //step 9
	pc.GenerateKeysForDiscussing([]Reviewer{reviewer})
	pc.CollectReviews(p.Id) //step 11
	reviewer.GradePaper(2)
	reviewer.SignReviewPaperCommit()
	reviewer.SignCommitsAndNonce()
	reviewer.SignAndEncryptGrade()
	
	pc.AcceptPaper(p.Id)
	pc.CompileGrades()
	actual := pc.GetCompiledGrades()
	pc.RevealAcceptedPaperInfo(p.Id)
	expected := []int64{2}

	assert.Equal(t, expected, actual, "failz")
}

func TestRevealAcceptedPaperInfo(t *testing.T) {
	submitter.Submit(&p)
	reviewerSlice := []*Reviewer{&reviewer}
	reviewer.SignBidAndEncrypt(&p)
	pc.assignPaper(reviewerSlice)
	pc.MatchPapers()                           //step 7
	reviewer.FinishReview("Pretty rad paper!") //step 8
	reviewer.SignReviewPaperCommit()           //step 9
	pc.GenerateKeysForDiscussing([]Reviewer{reviewer})
	pc.CollectReviews(p.Id) //step 11
	reviewer.GradePaper(2)
	reviewer.SignReviewPaperCommit()
	reviewer.SignCommitsAndNonce()
	reviewer.SignAndEncryptGrade()
	
	pc.AcceptPaper(p.Id)
	pc.CompileGrades()
	actual := pc.RevealAcceptedPaperInfo(p.Id)
	p := pc.GetPaperAndRandomness(p.Id)
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

