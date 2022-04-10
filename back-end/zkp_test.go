package backend

import (
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"testing"
)

func TestNewEqProofK256(t *testing.T) {
	keys := newKeys()
	p := Paper{
		1,
		&CommitStruct{},
		true,
		nil,
	}
	submitter := Submitter{
		keys,
		"1", //userID
		&CommitStruct{},
		&p,
		&Receiver{},
		nil,
		nil,
	}
	reviewer := Reviewer{
		"reviewer1",
		newKeys(),
		nil,
		map[int][]byte{},
		nil,
		&p,
		nil,
		nil,
	}

	curve1 := elliptic.P256()
	curve := curve1.Params()
	r1, _ := rand.Int(rand.Reader, curve.N)
	r2, _ := rand.Int(rand.Reader, curve.N)
	nonce, _ := rand.Int(rand.Reader, curve.N)
	
	// q1x, q1y, q2x, q2y := getGenerators()
	// b := NewCommitment(x, r1, q1x, q1y)
	// c := NewCommitment(x, r2, q2x, q2y)

	//x, _ := rand.Int(rand.Reader, curve.N)

	msg := GetRandomInt(submitter.keys.D)

	//msg := MsgToBigInt(p)
	
	commit1, err := submitter.GetCommitMessagePaper(msg)
	fmt.Println(commit1)

	if err != nil {
		t.Errorf("Error in GetCommitMsgPaper: %v", err)
	}

	commit2, err := reviewer.GetCommitMessageReviewPaper(msg)
	if err != nil {
		t.Errorf("Error in GetCommitMsgPaperReviewer: %v", err)
	}
	fmt.Printf("%#v", commit2)

	c1 := &Commitment{
		submitter.paperCommittedValue.CommittedValue.CommittedValue.X,
		submitter.paperCommittedValue.CommittedValue.CommittedValue.Y,
	}
	c2 := &Commitment{
		reviewer.paperCommittedValue.CommittedValue.CommittedValue.X,
		reviewer.paperCommittedValue.CommittedValue.CommittedValue.Y,
	}
	
	r1 = submitter.paperCommittedValue.CommittedValue.r
	r2 = reviewer.paperCommittedValue.CommittedValue.r


	proof := NewEqProofP256(msg, r1, r2, nonce, &submitter.keys.PublicKey, &reviewer.keys.PublicKey)
	if !proof.OpenP256(c1, c2, nonce, &submitter.keys.PublicKey, &reviewer.keys.PublicKey) {
		t.Fail()
	}
}