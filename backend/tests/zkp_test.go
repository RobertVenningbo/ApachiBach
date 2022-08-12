package backend_test

import (
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
	ls "math/rand"
	. "swag/backend"
	_ "swag/ec"
	"testing"
	"time"
	"github.com/0xdecaf/zkrp/ccs08"
	"github.com/stretchr/testify/assert"
)

func TestNewEqProofP256(t *testing.T) {

	InitGobs()
	submitterKey := submitter.Keys.PublicKey
	Pc.Keys = NewKeys()

	curve1 := elliptic.P256()
	curve := curve1.Params()

	r1, _ := rand.Int(rand.Reader, curve.N)
	r2, _ := rand.Int(rand.Reader, curve.N)
	nonce, _ := rand.Int(rand.Reader, curve.N)

	PaperBigInt := MsgToBigInt(EncodeToBytes(p))

	commit1, err := submitter.GetPaperSubmissionCommit(PaperBigInt, r1)
	if err != nil {
		t.Errorf("Error in GetCommitMsgPaper: %v", err)
	}

	commit2 := Pc.GetPaperReviewCommitPC(PaperBigInt, r2)
	if err != nil {
		t.Errorf("Error in GetCommitMsgPaperPC: %v", err)
	}

	c1 := &Commitment{
		X: commit1.X,
		Y: commit1.Y,
	}
	c2 := &Commitment{
		X: commit2.X,
		Y: commit2.Y,
	}

	proof := NewEqProofP256(PaperBigInt, r1, r2, nonce, &submitterKey, &Pc.Keys.PublicKey) //Commitment equality proof

	got := proof.OpenP256(c1, c2, nonce, &submitterKey, &Pc.Keys.PublicKey) //Verify commitment equality proof
	fmt.Printf("\n%s %v", "Commits hold same paper: ", got)
	want := true
	assert.Equal(t, want, got, "TestEqProof Failed")
}

func TestMsgToBigInt(t *testing.T) {
	msg := MsgToBigInt(EncodeToBytes(p))
	msg1 := MsgToBigInt(EncodeToBytes(p))
	assert.Equal(t, msg, msg1, "TestMsgToBigInt failed")
}

func BenchmarkTestCommitmentEqualityNizk(b *testing.B) {
	Pc.Keys = NewKeys()
	InitGobs()
	for n := 0; n < b.N; n++ {
		submitterKey := submitter.Keys.PublicKey

		curve1 := elliptic.P256()
		curve := curve1.Params()

		r1, _ := rand.Int(rand.Reader, curve.N)
		r2, _ := rand.Int(rand.Reader, curve.N)
		nonce, _ := rand.Int(rand.Reader, curve.N)

		PaperBigInt := MsgToBigInt(EncodeToBytes(p))

		commit1, err := submitter.GetPaperSubmissionCommit(PaperBigInt, r1)
		if err != nil {
			b.Errorf("Error in GetCommitMsgPaper: %v", err)
		}

		commit2 := Pc.GetPaperReviewCommitPC(PaperBigInt, r2)
		if err != nil {
			b.Errorf("Error in GetCommitMsgPaperPC: %v", err)
		}

		c1 := &Commitment{
			X: commit1.X,
			Y: commit1.Y,
		}
		c2 := &Commitment{
			X: commit2.X,
			Y: commit2.Y,
		}

		proof := NewEqProofP256(PaperBigInt, r1, r2, nonce, &submitterKey, &Pc.Keys.PublicKey) //Commitment equality proof

		got := proof.OpenP256(c1, c2, nonce, &submitterKey, &Pc.Keys.PublicKey) //Verify commitment equality proof
		want := true
		assert.Equal(b, want, got, "TestEqProof Failed")
	}

}

func TestZKSetMembership(t *testing.T) {
	var gradeSet []int64
	var i int64
	for i = 0; i < 500; i++ {
		gradeSet = append(gradeSet, i)
	}
	params, errSetup := ccs08.SetupSet(gradeSet)

	if errSetup != nil {
		log.Panicln(errSetup)
	}

	ls.Seed(time.Now().UnixNano())
	min := int64(1)
	max := int64(500)
	grade := ls.Int63n(max-min+1) + min
	r, _ := rand.Int(rand.Reader, elliptic.P256().Params().N)
	proof_out, _ := ccs08.ProveSet(grade, r, params)
	result, _ := ccs08.VerifySet(&proof_out, &params)
	assert.Equal(t, true, result, "TestZKMembership failed")
}

func BenchmarkTestZKSetMembership(b *testing.B) {
	var gradeSet []int64
	var i int64
	for i = 0; i < 500; i++ {
		x := ls.Int63n(1844674407370955161) //some random large number to generate from, 1 bit smaller than int64 max cap.
		gradeSet = append(gradeSet, x)
	}
	params, errSetup := ccs08.SetupSet(gradeSet)

	if errSetup != nil {
		log.Panicln(errSetup)
	}

	for n := 0; n < b.N; n++ {
		r, _ := rand.Int(rand.Reader, elliptic.P256().Params().N)
		proof_out, _ := ccs08.ProveSet(gradeSet[n], r, params)
		result, _ := ccs08.VerifySet(&proof_out, &params)
		assert.Equal(b, true, result, "TestZKMembership failed")
	}
	
}
