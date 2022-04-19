package backend

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	_ "swag/ec"
	ec "swag/ec"
	"testing"

	"github.com/stretchr/testify/assert"
)

//  func TestMatchPapers(t *testing.T) {
//  	p := Paper{
//  		1,
//  		true,
//  		nil,
//  		nil,
//  	}
//  	submitter.PaperCommittedValue.Paper = &p
//  	allPapers := append(pc.allPapers, &p)
//  	fmt.Printf("Paper: %v in allPapers", p.Id)
//  	submitters := []Submitter{submitter}
//  	reviewers := []Reviewer{reviewer}
//  	r := ec.GetRandomInt(submitter.Keys.D)
//  	PaperBigInt := MsgToBigInt(EncodeToBytes(p))
//  	SubmitterBigInt := MsgToBigInt(EncodeToBytes(submitter))
//  	PaperSubmissionCommit, _ := submitter.GetCommitMessagePaper(PaperBigInt, r)
//  	fmt.Printf("%s %v \n","PaperSubmissionCommitT:", *PaperSubmissionCommit)
//  	IdentityCommit, _ := submitter.GetCommitMessage(SubmitterBigInt, r)
//  	commitMsg := CommitMsg{
//  		EncodeToBytes(IdentityCommit),
//  		EncodeToBytes(PaperSubmissionCommit),
//  	}
//  	signedCommitMsg := SignsPossiblyEncrypts(submitter.Keys, EncodeToBytes(commitMsg), "")
//  	tree.Put("signedCommitMsg"+submitter.UserID, signedCommitMsg)
//  	reviewer.SignBidAndEncrypt(&p)
//  	// got :=pc.assignPaper(reviewers)
//  	// fmt.Printf("%v  \n", got)
//  	pc.matchPapers(nil,nil,nil)
//  }
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
		&Paper{},
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
 	assert.Equal(t, reviewerScope, paperBid.Reviewer, "TestGetBiddedPaper failed")
}


func TestAssignPapers(t *testing.T) {
	reviewer3 := Reviewer{
		"reviewer3",
		newKeys(),
		nil,
		nil,
		nil,
	}
	reviewer4 := Reviewer{
		"reviewer4",
		newKeys(),
		nil,
		nil,
		nil,
	}
	p1 := Paper{
		1,
		false,
		nil,
		nil,
	}
	p2 := Paper{
		2,
		false,
		nil,
		nil,
	}
	p3 := Paper{
		3,
		false,
		nil,
		nil,
	}

	pc.allPapers = append(pc.allPapers, &p1, &p2, &p3)
	reviewerSlice := []*Reviewer{&reviewer, &reviewer2, &reviewer3, &reviewer4}

	reviewer.SignBidAndEncrypt(&p1)
	reviewer2.SignBidAndEncrypt(&p1)
	reviewer3.SignBidAndEncrypt(&p1)
	reviewer4.SignBidAndEncrypt(&p1)

	pc.assignPaper(reviewerSlice)

	//TODO insert assert
}

func TestSupplyNizk(t *testing.T) {
	curve1 := elliptic.P256()
	curve := curve1.Params()

	submitter.Submit(&p)

	submitStruct := pc.GetPaperAndRandomness(1)
	rr := submitStruct.Rr
	
	PaperBigInt := MsgToBigInt(EncodeToBytes(p))

	nonce, _ := rand.Int(rand.Reader, curve.N)
	ReviewCommit, _ := pc.GetCommitMessagePaperPC(PaperBigInt, rr)

	reviewStruct := ReviewSignedStruct{
		ReviewCommit,
		&[]ecdsa.PublicKey{reviewer.Keys.PublicKey},
		nonce,
	}

	signature := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(reviewStruct), "")

	msg := fmt.Sprintf("ReviewSignedStruct with P%v", p.Id)
	tree.Put(msg, signature)
	

	got := pc.supplyNIZK(&p)
	want := true

	assert.Equal(t, want, got, "Nizk failed")
}

func TestGetReviewSignedStruct(t *testing.T) {
	pc.allPapers = append(pc.allPapers, &p)
	rr := ec.GetRandomInt(pc.Keys.D)
	PaperBigInt := MsgToBigInt(EncodeToBytes(p))

	commit, _ := pc.GetCommitMessagePaperPC(PaperBigInt, rr)
	nonce_r := ec.GetRandomInt(pc.Keys.D)

	//reviewerKeyList := []ecdsa.PublicKey{}

	reviewStruct := ReviewSignedStruct{
		commit,
		nil,
		nonce_r,
	}

	signature := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(reviewStruct), "")

	msg := fmt.Sprintf("ReviewSignedStruct with P%v", p.Id)
	tree.Put(msg, signature)

	r_struct := pc.GetReviewSignedStruct(p.Id)

	assert.Equal(t, reviewStruct, r_struct, "TestGetReviewStruct Failed")

}