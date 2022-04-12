package backend

import (
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	ec "swag/ec"
	"testing"

	"github.com/0xdecaf/zkrp/bulletproofs"
	"github.com/stretchr/testify/assert"
)

func TestNewEqProofK256(t *testing.T) {
	keys := newKeys()
	p := Paper{
		1,
		true,
		nil,
		nil,
	}
	submitter := Submitter{
		keys,
		"1", //userID
		&CommitStruct{},
		&CommitStructPaper{},
		&Receiver{},
	}
	reviewer := Reviewer{
		"reviewer1",
		newKeys(),
		map[int][]byte{},
		&CommitStructPaper{},
		nil,
		nil,
	}

	curve1 := elliptic.P256()
	curve := curve1.Params()

	r1 := ec.GetRandomInt(curve.N)
	r2 := ec.GetRandomInt(curve.N)
	nonce, _ := rand.Int(rand.Reader, curve.N)

	msg := MsgToBigInt(EncodeToBytes(p))
	fmt.Printf("%s %v \n", "msg : ", msg)
	fmt.Printf("%s %v \n", "key : ", submitter.Keys.D)
	//msg := ec.GetRandomInt(submitter.keys.D)

	commit1, err := submitter.GetCommitMessagePaper(msg, r1)
	fmt.Println(commit1)

	if err != nil {
		t.Errorf("Error in GetCommitMsgPaper: %v", err)
	}

	commit2, err := reviewer.GetCommitMessageReviewPaper(msg, r2)
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

	proof := NewEqProofP256(msg, r1, r2, nonce, &submitter.Keys.PublicKey, &reviewer.Keys.PublicKey)
	if !proof.OpenP256(c1, c2, nonce, &submitter.Keys.PublicKey, &reviewer.Keys.PublicKey) {
		t.Fail()
	}
}

func TestZKSetMembership(t *testing.T) {
	// Set up the range, [18, 200) in this case.
	// We want to prove that we are over 18, and less than 200 years old.
	params, errSetup := bulletproofs.SetupGeneric(18, 200)
	if errSetup != nil {
		t.Errorf(errSetup.Error())
		t.FailNow()
	}

	// Create the proof
	bigSecret := new(big.Int).SetInt64(int64(40))
	proof, errProve := bulletproofs.ProveGeneric(bigSecret, params)
	if errProve != nil {
		t.Errorf(errProve.Error())
		t.FailNow()
	}

	// Encode the proof to JSON
	jsonEncoded, err := json.Marshal(proof)
	if err != nil {
		t.Fatal("encode error:", err)
	}

	// Here the proof is passed to the verifier, possibly over a network.

	// Decode the proof from JSON
	var decodedProof bulletproofs.ProofBPRP
	err = json.Unmarshal(jsonEncoded, &decodedProof)
	if err != nil {
		t.Fatal("decode error:", err)
	}

	assert.Equal(t, proof, decodedProof, "should be equal")

	// Verify the proof
	ok, errVerify := decodedProof.Verify()
	if errVerify != nil {
		t.Errorf(errVerify.Error())
		t.FailNow()
	}
	assert.True(t, ok, "should verify")
}
