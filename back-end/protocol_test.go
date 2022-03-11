package backend

import (
	"fmt"
	"swag/ec"
	"testing"
	"github.com/stretchr/testify/assert"
	"crypto/ecdsa"
	"crypto/rand"
)

func TestGenerateSharedSecret(t *testing.T) {
	// pc := PC{
	// 	newKeys(),
	// 	RandomNumber{nil, nil, nil, nil},
	// }
	// submitter := Submitter{
	// 	newKeys(),
	// 	RandomNumber{nil, nil, nil, nil},
	// 	"1", //userID
	// 	nil,
	// 	Paper{},
	// 	nil,
	// 	nil,
	// }

	// got := generateSharedSecret(&pc, &submitter)
	// fmt.Printf(got)

}

func TestVerifyTrapdoorSubmitter(t *testing.T) {
	keys := newKeys()
	submitter := Submitter{
		keys,
		"1", //userID
		&CommitStruct{},
		&Paper{},
		&Receiver{keys, nil},
		nil,
		nil,
	}
	//fmt.Println(submitter.random.Rr)
	got := submitter.VerifyTrapdoorSubmitter(GetTrapdoor(submitter.receiver))

	fmt.Printf("%t", got)

	want := true

	if got != want {
		t.Errorf("TestGetCommitMessageVerifyTrapdoorSubmitter failed")
	}
}

func TestPedersenCommitment(t *testing.T) {
	keys := newKeys()
	submitter := Submitter{
		keys,
		"1", //userID
		&CommitStruct{},
		&Paper{},
		&Receiver{},
		nil,
		nil,
	}

	submitter.receiver = NewReceiver(submitter.keys)

	a := ec.GetRandomInt(submitter.keys.D)

	c, err := submitter.GetCommitMessage(a)

	if err != nil {
		t.Errorf("Error in GetCommitMsg: %v", err)
	}

	SetCommitment(submitter.receiver, c)
	submittedVal, r := submitter.GetDecommitMsg()
	
	success := submitter.receiver.CheckDecommitment(r, submittedVal)

	assert.Equal(t, true, success, "pedersen failed")

}

func TestCommitSignatureAndVerify(t *testing.T) {
	keys := newKeys()
	s := Submitter{
		keys,
		"1", //userID
		&CommitStruct{},
		&Paper{},
		&Receiver{},
		nil,
		nil,
	}
	
	a := ec.GetRandomInt(s.keys.D)
	fmt.Println(a)
	
	c, _ := s.GetCommitMessage(a)

	hashedMsgSubmit, _ := GetMessageHash([]byte(fmt.Sprintf("%v", c)))

	signatureSubmit, _ := ecdsa.SignASN1(rand.Reader, s.keys, hashedMsgSubmit) //rand.Reader idk??

	got := ecdsa.VerifyASN1(&s.keys.PublicKey, hashedMsgSubmit, signatureSubmit) //testing

	assert.Equal(t, true, got, "Sign and Verify failed")
}

/*


func TestSubmit() {

}

func TestGetMessageHash() {

}


func TestVerifyTrapdoor() {

}

func TestEquals() {

}

func TestGenCommitmentKey() {

}

func TestCommit() {

}

*/

func TestEncryptAndDecrypt(t *testing.T) {
	passphrase := "password"

	stuffToEncrypt := []byte("wauw123")

	got := Decrypt(Encrypt(stuffToEncrypt, passphrase), passphrase)

	want := stuffToEncrypt

	if string(got) != string(want) {
		t.Errorf("TestEncryptAndDecrypt Failed")
		t.Fail()
	} else {

	}

}
