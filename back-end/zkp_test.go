package backend

import (
	"crypto/elliptic"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	_ "swag/ec"
	"testing"

	"github.com/0xdecaf/zkrp/bulletproofs"
	"github.com/stretchr/testify/assert"
)

func TestNewEqProofP256(t *testing.T) {
	submitter.Submit(&p)
	submitterKey := pc.GetPaperSubmitterPK(p.Id)

	curve1 := elliptic.P256()
	curve := curve1.Params()

	r1, _ := rand.Int(rand.Reader, curve.N)
	r2, _ := rand.Int(rand.Reader, curve.N)
	nonce, _ := rand.Int(rand.Reader, curve.N)

	//msg := ec.GetRandomInt(submitter.Keys.D)
	msg := MsgToBigInt(EncodeToBytes(p))

	commit1, err := submitter.GetCommitMessagePaper(msg, r1)
	if err != nil {
		t.Errorf("Error in GetCommitMsgPaper: %v", err)
	}

	commit2, err := pc.GetCommitMessagePaperPC(msg, r2)
	if err != nil {
		t.Errorf("Error in GetCommitMsgPaperPC: %v", err)
	}

	c1 := &Commitment{
		commit1.X,
		commit1.Y,
	}
	c2 := &Commitment{
		commit2.X,
		commit2.Y,
	}

	proof := NewEqProofP256(msg, r1, r2, nonce, &submitterKey, &pc.Keys.PublicKey)

	got := proof.OpenP256(c1, c2, nonce, &submitterKey, &pc.Keys.PublicKey)
	fmt.Printf("\n%s %v", "Commits hold same paper: ", got)
	want := true
	assert.Equal(t, want, got, "TestEqProof Failed")
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

func TestMsgToBigInt(t *testing.T) {
	msg := MsgToBigInt(EncodeToBytes(p))
	msg1 := MsgToBigInt(EncodeToBytes(p))
	assert.Equal(t, msg, msg1, "failzzMsgToBigInt")
}
