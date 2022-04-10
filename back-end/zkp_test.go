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
	//x := GetRandomInt(reviewer.keys.D)
	
	r1 := GetRandomInt(curve.N)
	r2 := GetRandomInt(curve.N)
	nonce, _ := rand.Int(rand.Reader, curve.N)
	
	// q1x, q1y, q2x, q2y := getGenerators()
	// b := NewCommitment(x, r1, q1x, q1y)
	// c := NewCommitment(x, r2, q2x, q2y)

	//x, _ := rand.Int(rand.Reader, curve.N)

	//msg := GetRandomInt(submitter.keys.D)

	msg := MsgToBigInt(EncodeToBytes(p))
	
	commit1, err := submitter.GetCommitMessagePaperTest(msg, r1)
	fmt.Println(commit1)
	
	if err != nil {
		t.Errorf("Error in GetCommitMsgPaper: %v", err)
	}

	commit2, err := reviewer.GetCommitMessageReviewPaperTest(msg, r2)
	if err != nil {
		t.Errorf("Error in GetCommitMsgPaperReviewer: %v", err)
	}
	fmt.Printf("%#v", commit2)

	c1 := &Commitment{
		commit1.X,
		commit1.Y,
	}
	c2 := &Commitment{
		commit2.X,
		commit2.Y,
	}
	
	//r1 = submitter.paperCommittedValue.CommittedValue.r
	//r2 = reviewer.paperCommittedValue.CommittedValue.r


	proof := NewEqProofP256(msg, r1, r2, nonce, &submitter.keys.PublicKey, &reviewer.keys.PublicKey)
	if !proof.OpenP256(c1, c2, nonce, &submitter.keys.PublicKey, &reviewer.keys.PublicKey) {
		t.Fail()
	}
}