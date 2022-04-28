package backend_test

import (
	"fmt"
	"testing"
	. "swag/backend"

	"github.com/stretchr/testify/assert"
)

func TestCalculateNearestGrade(t *testing.T) {

	avg := 5.6
	want := 7

	got := CalculateNearestGrade(avg)
	assert.Equal(t, want, got, "TestCalculateNearestGrade failed")
}

func TestGradePaperAndGetGrade(t *testing.T) {
	reviewerSlice := []Reviewer{reviewer}
	Pc.GenerateKeysForDiscussing(reviewerSlice) //Calling this to fill log with necessary data, has been tested in reviewing_test.go
	reviewer.PaperCommittedValue.Paper = &p
	want := 7
	reviewer.GradePaper(want)

	GradeStruct := reviewer.GetGradeForReviewer(reviewer.UserID)
	got := GradeStruct.Grade

	assert.Equal(t, want, got, "TestGradePaperAndGetGrade failed")
}

func TestAgreeOnGrade(t *testing.T) {
	reviewer.PaperCommittedValue.Paper = &paperListTest[0]
	reviewer2.PaperCommittedValue.Paper = &paperListTest[0]
	reviewer3.PaperCommittedValue.Paper = &paperListTest[0]
	reviewerSlice := []Reviewer{reviewer, reviewer2, reviewer3}
	Pc.GenerateKeysForDiscussing(reviewerSlice) //Calling this to fill log with necessary data, has been tested in reviewing_test.go

	paperListTest[0].ReviewerList = append(paperListTest[0].ReviewerList, reviewer, reviewer2, reviewer3)

	reviewer.GradePaper(4)
	reviewer2.GradePaper(7)
	reviewer3.GradePaper(12)

	got := AgreeOnGrade(&paperListTest[0])
	want := 7

	assert.Equal(t, want, got, "TestAgreeOnGrade")
}

func TestMakeGradeCommit(t *testing.T) {
	reviewerSlice := []Reviewer{reviewer, reviewer2}
	Pc.GenerateKeysForDiscussing(reviewerSlice) //Calling this to fill log with necessary data, has been tested in reviewing_test.go
	reviewer.PaperCommittedValue.Paper = &p
	reviewer2.PaperCommittedValue.Paper = &p
	gradeCommit := reviewer.MakeGradeCommit()
	gradeCommit2 := reviewer2.MakeGradeCommit()

	assert.Equal(t, gradeCommit, gradeCommit2, "TestMakeGradeCommit Failed")
}

func TestSignCommitsAndNonce(t *testing.T) { //TODO Test with Get functions
	Pc.AllPapers = append(Pc.AllPapers, &p)
	submitter.Submit(&p)
	reviewerSlice := []*Reviewer{&reviewer}
	reviewerSlice1 := []Reviewer{reviewer}
	reviewer.SignBidAndEncrypt(&p)
	Pc.AssignPaper(reviewerSlice)
	Pc.MatchPapers()

	Pc.GenerateKeysForDiscussing(reviewerSlice1) //Calling this to fill log with necessary data, has been tested in reviewing_test.go
	reviewer.PaperCommittedValue.Paper = &p
	reviewer.GradePaper(7)
	reviewer.SignReviewPaperCommit()
	reviewer.SignCommitsAndNonce()
	reviewer.SignAndEncryptGrade()

	gradeReviewCommit := GradeReviewCommits{
		reviewer.GetReviewCommitNonceStruct().Commit,
		reviewer.MakeGradeCommit(),
		reviewer.GetReviewCommitNonceStruct().Nonce,
	}

	fmt.Printf("%#v\n", gradeReviewCommit)

}
