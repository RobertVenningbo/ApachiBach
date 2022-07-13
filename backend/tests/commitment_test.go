package backend_test

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	. "swag/backend"
	ec "swag/ec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyTrapdoorSubmitter(t *testing.T) {
	rec := NewReceiver(submitter.Keys)
	submitter.Receiver = rec
	got := submitter.VerifyTrapdoorSubmitter(GetTrapdoor(submitter.Receiver))
	want := true

	if got != want {
		t.Errorf("TestGetCommitMessageVerifyTrapdoorSubmitter failed")
	}
}


func TestPedersenCommitment(t *testing.T) {

	submitter.Receiver = NewReceiver(submitter.Keys)

	a := ec.GetRandomInt(submitter.Keys.D) //secret value
	b := ec.GetRandomInt(submitter.Keys.D) //random
	c, err := submitter.GetPaperSubmissionCommit(a, b)

	if err != nil {
		t.Errorf("Error in GetCommitMsg: %v", err)
	}

	SetCommitment(submitter.Receiver, c)
	submittedVal, r := submitter.GetDecommitMsgPaper()

	success := submitter.Receiver.CheckDecommitment(r, submittedVal)

	assert.Equal(t, true, success, "pedersen commitment failed")

}

func TestCommitSignatureAndVerify(t *testing.T) {

	a := ec.GetRandomInt(submitter.Keys.D)
	b := ec.GetRandomInt(submitter.Keys.D)
	c, _ := submitter.GetCommitMessage(a, b)

	hashedMsgSubmit1, _ := GetMessageHash([]byte(fmt.Sprintf("%v", c)))

	signatureSubmit, _ := ecdsa.SignASN1(rand.Reader, submitter.Keys, hashedMsgSubmit1) //rand.Reader idk??

	got := ecdsa.VerifyASN1(&submitter.Keys.PublicKey, hashedMsgSubmit1, signatureSubmit) //testing

	assert.Equal(t, true, got, "Sign and Verify failed")
}
