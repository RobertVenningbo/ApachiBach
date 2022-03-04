package backend

import (
	"fmt"
	"testing"
)

func TestGenerateSharedSecret(t *testing.T) {
	pc := PC{
		newKeys(),
		RandomNumber{nil, nil, nil, nil},
	}
	submitter := Submitter{
		newKeys(),
		RandomNumber{nil, nil, nil, nil},
		"1", //userID
		nil,
		Paper{},
		nil,
		nil,
	}

	got := generateSharedSecret(&pc, &submitter)
	fmt.Printf(got)

}

func TestGetCommitMessageVerifyTrapdoorSubmitter(t *testing.T) {
	submitter := Submitter{
		newKeys(),
		RandomNumber{nil, nil, nil, nil},
		"1", //userID
		nil,
		Paper{},
		nil,
		nil,
	}

	val := GetRandomInt(submitter.keys.D)
	fmt.Println(submitter.random.Ri)

	c, _ := submitter.GetCommitMessage(val)

	submitter.SubmitterCommittedValue = c

	got := submitter.VerifyTrapdoorSubmitter(submitter.random.Ri)
	fmt.Printf("%t", got)

	want := true

	if got != want {
		t.Errorf("TestGetCommitMessageVerifyTrapdoorSubmitter failed")
	}
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
