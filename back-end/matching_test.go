package backend

import (
	"fmt"
	"swag/ec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchPapers(t *testing.T) {
	submitter.PaperCommittedValue.Paper = p
	allPapers := append(pc.allPapers, p)
	fmt.Printf("Paper: %v in allPapers", p.Id)
	submitters := []Submitter{submitter}
	reviewers := []Reviewer{reviewer}

	r := ec.GetRandomInt(submitter.Keys.D)
	PaperBigInt := MsgToBigInt(EncodeToBytes(p))
	SubmitterBigInt := MsgToBigInt(EncodeToBytes(submitter))
	commit, _ := submitter.GetCommitMessagePaper(PaperBigInt, r)
	commit2, _ := submitter.GetCommitMessage(SubmitterBigInt, r)

	commitMsg := CommitMsg{
		commit2,
		commit,
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
	reviewerScope := Reviewer{
		"reviewer123123",
		newKeys(),
		&CommitStructPaper{},
		nil,
		nil,
	}

	reviewerScope.SignBidAndEncrypt(&p)

	paperBid := reviewerScope.getBiddedPaper()

	assert.Equal(t, p, paperBid.Paper, "TestGetBiddedPaper failed")
	assert.Equal(t, reviewerScope, paperBid.Reviewer, "TestGetBiddedPaper failed") //TODO: CURRENTLY THERE'S A BUG WHICH DOESN'T ALLOW NESTED STRUCTS
}
