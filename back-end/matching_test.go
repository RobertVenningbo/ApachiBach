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


func TestDistributeAndGetPapersForReviewers(t *testing.T) {
	pc.allPapers = append(pc.allPapers, &p)
	reviewers := []Reviewer{reviewer, reviewer2}

	//Papers have been put into the log encrypted with a shared secret key between the given reviewer and the pc.
	pc.distributePapers(reviewers, pc.allPapers)

	//reviewer2 now wants to retrieve his papers.
	//intended to be called with pc.allPapers which is a general lookup table for paper.Ids etc.
	retrievedPapers := reviewer2.GetPapersReviewer(pc.allPapers)

	assert.Equal(t, pc.allPapers, retrievedPapers, "TestDistributeAndGetPapersForReviewers failed")
}

func TestGetBiddedPaper(t *testing.T) {
	commitStructPaper := &CommitStructPaper{
		nil,
		nil,
		nil,
		&Paper{},
	}
	reviewerScope := &Reviewer{
		123123,
		newKeys(),
		commitStructPaper,
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
		3,
		newKeys(),
		nil,
	}
	reviewer4 := Reviewer{
		4,
		newKeys(),
		nil,
	}
	p1 := Paper{
		1,
		false,
		nil,
	}
	p2 := Paper{
		2,
		false,
		nil,
	}
	p3 := Paper{
		3,
		false,
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
	keys := newKeys()
	submitter1 := Submitter{
		keys,
		2, //userID
		&CommitStruct{},
		&CommitStructPaper{},
		&Receiver{},
	}
	curve1 := elliptic.P256()
	curve := curve1.Params()

	submitter1.Submit(&p)
	submitStruct := pc.GetPaperAndRandomness(p.Id)
	rr := submitStruct.Rr

	PaperBigInt := MsgToBigInt(EncodeToBytes(p.Id))
	nonce, _ := rand.Int(rand.Reader, curve.N)
	ReviewCommit, _ := pc.GetCommitMessagePaperPC(PaperBigInt, rr)

	reviewStruct := ReviewSignedStruct{
		ReviewCommit,
		[]ecdsa.PublicKey{reviewer.Keys.PublicKey}, //reviewer doesnt have a key, might propegate this error: "gob: cannot encode nil pointer of type *ecdsa.PublicKey inside interface"
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
	PaperBigInt := MsgToBigInt(EncodeToBytes(p.Id))

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

func TestMatchPapers(t *testing.T) {
	submitter.Submit(&p)
	reviewerSlice := []*Reviewer{&reviewer}
	reviewer.SignBidAndEncrypt(&p)
	pc.assignPaper(reviewerSlice)
	pc.MatchPapers()
}

