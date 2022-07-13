package backend_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	_ "swag/ec"
	ec "swag/ec"
	"testing"
	. "swag/backend"

	"github.com/stretchr/testify/assert"
)

func TestDistributeAndGetPapersForReviewers(t *testing.T) {
	Pc.AllPapers = append(Pc.AllPapers, &p)
	reviewers := []Reviewer{reviewer, reviewer2}

	//Papers have been put into the log encrypted with a shared secret key between the given reviewer and the pc.
	Pc.DistributePapers(reviewers, Pc.AllPapers)

	//reviewer2 now wants to retrieve his papers.
	//intended to be called with pc.allPapers which is a general lookup table for paper.Ids etc.
	retrievedPapers := reviewer2.GetPapersReviewer(Pc.AllPapers)
	fmt.Print(Pc.Keys)
	assert.Equal(t, Pc.AllPapers, retrievedPapers, "TestDistributeAndGetPapersForReviewers failed")
}

func TestGetBiddedPaper(t *testing.T) {
	commitStructPaper := &CommitStructPaper{
		CommittedValue: nil,
		R: nil,
		Val: nil,
		Paper: &Paper{},
	}
	reviewerScope := &Reviewer{
		UserID: 123123,
		Keys: NewKeys(),
		PaperCommittedValue: commitStructPaper,
	}

	reviewerScope.SignBidAndEncrypt(&p)

	paperBid := Pc.GetBiddedPaper(reviewerScope.UserID)

	fmt.Printf("%s %v \n", "reviewer1: ", reviewerScope)
	fmt.Printf("%s %v \n", "reviewer2: ", paperBid.Reviewer)

	assert.Equal(t, p, *paperBid.Paper, "TestGetBiddedPaper failed")
	assert.Equal(t, reviewerScope, paperBid.Reviewer, "TestGetBiddedPaper failed")
}

func TestAssignPapers(t *testing.T) {

	p1 := Paper{
		1,
		false,
		nil,
		nil,
		"",
	}
	p2 := Paper{
		2,
		false,
		nil,
		nil,
		"",
	}

	Pc.AllPapers = append(Pc.AllPapers, &p1, &p2)

	bid1 := PaperBid{
		Paper: &p1,
		Reviewer: &reviewer,
	}
	bid2 := PaperBid{
		Paper: &p2,
		Reviewer: &reviewer,
	}
	bid3 := PaperBid{
		Paper: &p1,
		Reviewer: &reviewer2,
	}
	bid4 := PaperBid{
		Paper: &p2,
		Reviewer: &reviewer2,
	}
	var bidSlice []*PaperBid
	bidSlice = append(bidSlice, &bid1, &bid2, &bid3, &bid4)
	

	Pc.AssignPaper(bidSlice)

	//TODO insert assert
}

func TestSupplyNizk(t *testing.T) {
	keys := NewKeys()
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
	submitStruct := Pc.GetPaperAndRandomness(p.Id)
	rr := submitStruct.Rr

	PaperBigInt := MsgToBigInt(EncodeToBytes(p.Id))
	nonce, _ := rand.Int(rand.Reader, curve.N)
	ReviewCommit, _ := Pc.GetCommitMessagePaperPC(PaperBigInt, rr)

	reviewStruct := ReviewSignedStruct{
		ReviewCommit,
		[]ecdsa.PublicKey{reviewer.Keys.PublicKey}, //reviewer doesnt have a key, might propegate this error: "gob: cannot encode nil pointer of type *ecdsa.PublicKey inside interface"
		nonce,
	}

	signature := SignsPossiblyEncrypts(Pc.Keys, EncodeToBytes(reviewStruct), "")

	msg := fmt.Sprintf("ReviewSignedStruct with P%v", p.Id)
	Trae.Put(msg, signature)

	got := Pc.SupplyNIZK(&p)
	want := true

	assert.Equal(t, want, got, "Nizk failed")
}

func TestGetPapersReviewer(t *testing.T) {
	Pc.Keys = NewKeys()
	reviewer.Keys = NewKeys()
	Pc.AllPapers = append(Pc.AllPapers, &p)
	reviewerSlice := []Reviewer{reviewer}
	Pc.DistributePapers(reviewerSlice, Pc.AllPapers)
	reviewer.GetPapersReviewer(Pc.AllPapers)
}

func TestGetReviewSignedStruct(t *testing.T) {
	Pc.AllPapers = append(Pc.AllPapers, &p)
	rr := ec.GetRandomInt(Pc.Keys.D)
	PaperBigInt := MsgToBigInt(EncodeToBytes(p.Id))

	commit, _ := Pc.GetCommitMessagePaperPC(PaperBigInt, rr)
	nonce_r := ec.GetRandomInt(Pc.Keys.D)

	//reviewerKeyList := []ecdsa.PublicKey{}

	reviewStruct := ReviewSignedStruct{
		commit,
		nil,
		nonce_r,
	}

	signature := SignsPossiblyEncrypts(Pc.Keys, EncodeToBytes(reviewStruct), "")

	msg := fmt.Sprintf("ReviewSignedStruct with P%v", p.Id)
	Trae.Put(msg, signature)

	r_struct := Pc.GetReviewSignedStruct(p.Id)

	assert.Equal(t, reviewStruct, r_struct, "TestGetReviewStruct Failed")

}

func TestPCGetBiddedPaper(t *testing.T) {
	bid := &PaperBid{}
	submitter.Submit(&p)
	reviewer.MakeBid(&p)
	reviewer.SignBidAndEncrypt(&p)
	bid = Pc.GetBiddedPaper(reviewer.UserID)
	fmt.Println(bid.Paper.Id)

}

// func TestMatchPapers(t *testing.T) {
// 	submitter.Submit(&p)
// 	reviewerSlice := []*Reviewer{&reviewer}
// 	reviewer.SignBidAndEncrypt(&p)
// 	Pc.AssignPaper(reviewerSlice)
// 	Pc.MatchPapers()
// }
