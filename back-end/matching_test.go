package backend

import (
	"fmt"
	"swag/ec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchPapers(t *testing.T) {
	p := Paper{
		1,
		true,
		nil,
		nil,
	}
	submitter.PaperCommittedValue.Paper = p
	allPapers := append(pc.allPapers, p)
	fmt.Printf("Paper: %v in allPapers", p.Id)
	submitters := []Submitter{submitter}
	reviewers := []Reviewer{reviewer}

	r := ec.GetRandomInt(submitter.Keys.D)
	PaperBigInt := MsgToBigInt(EncodeToBytes(p))
	SubmitterBigInt := MsgToBigInt(EncodeToBytes(submitter))
	PaperSubmissionCommit, _ := submitter.GetCommitMessagePaper(PaperBigInt, r)
	fmt.Printf("%s %v \n", "PaperSubmissionCommitT:", *PaperSubmissionCommit)
	IdentityCommit, _ := submitter.GetCommitMessage(SubmitterBigInt, r)

	commitMsg := CommitMsg{
		EncodeToBytes(IdentityCommit),
		EncodeToBytes(PaperSubmissionCommit),
	}

	signedCommitMsg := SignsPossiblyEncrypts(submitter.Keys, EncodeToBytes(commitMsg), "")
	tree.Put("signedCommitMsg"+submitter.UserID, signedCommitMsg)

	reviewer.SignBidAndEncrypt(&p)
	// got :=pc.assignPaper(reviewers)
	// fmt.Printf("%v  \n", got)
	pc.matchPapers(reviewers, submitters, allPapers)
}

func TestDistributeAndGetPapersForReviewers(t *testing.T) {

	reviewers := []Reviewer{reviewer, reviewer2}

	//Papers have been put into the log encrypted with a shared secret key between the given reviewer and the pc.
	pc.distributePapers(reviewers, paperListTest)

	//reviewer2 now wants to retrieve his papers.
	//intended to be called with pc.allPapers which is a general lookup table for paper.Ids etc.
	retrievedPapers := reviewer2.GetPapersReviewer(paperListTest)

	assert.Equal(t, paperListTest, retrievedPapers, "TestGetPaperSubmissionSignature failed")
}

func TestGetBiddedPaper(t *testing.T) {
	commitStructPaper := &CommitStructPaper{
		nil,
		nil,
		nil,
		Paper{},
	}
	reviewerScope := &Reviewer{
		"reviewer123123",
		newKeys(),
		commitStructPaper,
		nil,
		nil,
	}

	reviewerScope.SignBidAndEncrypt(&p)

	paperBid := reviewerScope.getBiddedPaper()

	fmt.Printf("%s %v \n", "reviewer1: ", reviewerScope)
	fmt.Printf("%s %v \n", "reviewer2: ", paperBid.Reviewer)

	assert.Equal(t, p, *paperBid.Paper, "TestGetBiddedPaper failed")
	assert.Equal(t, reviewerScope, paperBid.Reviewer, "TestGetBiddedPaper failed") //TODO: CURRENTLY THERE'S A BUG WHICH DOESN'T ALLOW NESTED STRUCTS
}
