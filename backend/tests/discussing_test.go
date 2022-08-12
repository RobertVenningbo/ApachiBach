package backend_test

import (
	_ "fmt"
	. "swag/backend"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateNearestGrade(t *testing.T) {

	avg := 5.6
	want := 7

	got := CalculateNearestGrade(avg)
	assert.Equal(t, want, got, "TestCalculateNearestGrade failed")
}

func TestGradePaperAndGetGrade(t *testing.T) {
	reviewer.PaperCommittedValue.Paper = &p
	Pc.AllPapers[0] = &p
	Pc.AllPapers[0].ReviewerList = append(Pc.AllPapers[0].ReviewerList, reviewer)
	Pc.GenerateKeysForDiscussing() //Calling this to fill log with necessary data, has been tested in reviewing_test.go
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
	Pc.AllPapers[0] = &p
	Pc.AllPapers[0].ReviewerList = append(Pc.AllPapers[0].ReviewerList, reviewerSlice...)
	Pc.GenerateKeysForDiscussing() //Calling this to fill log with necessary data, has been tested in reviewing_test.go

	reviewer.GradePaper(4)
	reviewer2.GradePaper(7)
	reviewer3.GradePaper(12)

	got := reviewer.GetAgreedGroupGrade()
	want := 7

	assert.Equal(t, want, got, "TestAgreeOnGrade")
}

func TestMakeGradeCommit(t *testing.T) {
	reviewer.PaperCommittedValue.Paper = &p
	reviewer2.PaperCommittedValue.Paper = &p
	reviewerSlice := []Reviewer{reviewer, reviewer2}
	Pc.AllPapers[0] = &p
	paperListTest[0].ReviewerList = append(paperListTest[0].ReviewerList, reviewerSlice...)

	Pc.GenerateKeysForDiscussing()
	gradeCommit := reviewer.MakeGradeCommit()
	gradeCommit2 := reviewer2.MakeGradeCommit()

	assert.Equal(t, gradeCommit, gradeCommit2, "TestMakeGradeCommit Failed")
}

// func TestSignCommitsAndNonce(t *testing.T) { 
// 	Pc.AllPapers = append(Pc.AllPapers, &p)
// 	submitter.Submit(&p)
// 	reviewerSlice := []*Reviewer{&reviewer}
// 	reviewerSlice1 := []Reviewer{reviewer}
// 	reviewer.SignBidAndEncrypt(&p)
// 	Pc.AssignPaper(reviewerSlice)
// 	Pc.MatchPapers()

// 	Pc.GenerateKeysForDiscussing(reviewerSlice1) //Calling this to fill log with necessary data, has been tested in reviewing_test.go
// 	reviewer.PaperCommittedValue.Paper = &p
// 	reviewer.GradePaper(7)
// 	reviewer.SignReviewPaperCommit()
// 	reviewer.SignCommitsAndNonce()
// 	reviewer.SignAndEncryptGrade()

// 	gradeReviewCommit := GradeReviewCommits{
// 		reviewer.GetReviewCommitNonceStruct().Commit,
// 		reviewer.MakeGradeCommit(),
// 		reviewer.GetReviewCommitNonceStruct().Nonce,
// 	}

// 	fmt.Printf("%#v\n", gradeReviewCommit)

// }
