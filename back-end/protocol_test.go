package backend

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"swag/ec"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	paperListTest = []Paper{
		Paper{1, false, nil, nil},
		Paper{2, false, nil, nil},
	}
	p = Paper{
		1,
		true,
		nil,
		nil,
	}
	reviewer = Reviewer{
		"reviewer",
		newKeys(),
		map[int][]byte{},
		&CommitStructPaper{},
		nil,
		nil,
	}
	reviewer2 = Reviewer{
		"reviewer2",
		newKeys(),
		map[int][]byte{},
		nil,
		nil,
		nil,
	}
	submitter = Submitter{
		newKeys(),
		"submitter", //userID
		&CommitStruct{},
		&CommitStructPaper{},
		&Receiver{},
	}
)

func TestGenerateSharedSecret(t *testing.T) {
	got := generateSharedSecret(&pc, &submitter, nil)
	want := generateSharedSecret(&pc, &submitter, nil)
	assert.Equal(t, got, want, "Test failed")

}

func TestVerifyTrapdoorSubmitter(t *testing.T) {

	got := submitter.VerifyTrapdoorSubmitter(GetTrapdoor(submitter.Receiver))

	want := true

	if got != want {
		t.Errorf("TestGetCommitMessageVerifyTrapdoorSubmitter failed")
	}
}

func TestPedersenCommitment(t *testing.T) {

	submitter.Receiver = NewReceiver(submitter.Keys)

	a := ec.GetRandomInt(submitter.Keys.D)
	b := ec.GetRandomInt(submitter.Keys.D)
	c, err := submitter.GetCommitMessage(a, b)

	if err != nil {
		t.Errorf("Error in GetCommitMsg: %v", err)
	}

	SetCommitment(submitter.Receiver, c)
	submittedVal, r := submitter.GetDecommitMsg()

	success := submitter.Receiver.CheckDecommitment(r, submittedVal)

	assert.Equal(t, true, success, "pedersen failed")

}

func TestPedersenCommitmentPaper(t *testing.T) {

	submitter.Receiver = NewReceiver(submitter.Keys)

	a := ec.GetRandomInt(submitter.Keys.D)
	b := ec.GetRandomInt(submitter.Keys.D)
	c, err := submitter.GetCommitMessagePaper(a, b)

	if err != nil {
		t.Errorf("Error in GetCommitMsg: %v", err)
	}

	SetCommitment(submitter.Receiver, c)
	submittedVal, r := submitter.GetDecommitMsgPaper()

	success := submitter.Receiver.CheckDecommitment(r, submittedVal)

	assert.Equal(t, true, success, "pedersen paper commitment failed")

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

func TestEncryptAndDecrypt(t *testing.T) {
	passphrase := "password"

	stuffToEncrypt := []byte{6, 189, 116, 133, 88, 195, 101, 218, 69, 205, 49, 94, 107, 156, 84, 78, 157, 178, 189, 211, 132, 69, 199, 190, 147, 60, 231, 10, 14, 71, 92, 168, 121, 157, 21, 128, 145, 192, 40, 78, 189, 231, 197}

	got := Decrypt(Encrypt(stuffToEncrypt, passphrase), passphrase)

	want := stuffToEncrypt

	if string(got) != string(want) {
		t.Errorf("TestEncryptAndDecrypt Failed")
		t.Fail()
	} else {

	}

}

func TestDecodeToStruct(t *testing.T) {
	EncodedStruct := EncodeToBytes(p)
	DecodedStruct := DecodeToStruct1(EncodedStruct, Paper{})
	assert.Equal(t, DecodedStruct, p, "Test failed")

}

func TestVerifyMethod(t *testing.T) {
	

	a := ec.GetRandomInt(submitter.Keys.D)
	b := ec.GetRandomInt(submitter.Keys.D)
	c, _ := submitter.GetCommitMessage(a, b)

	signatureAndPlaintext := Sign(submitter.Keys, c) //TODO; current bug is that this hash within this function is not the same hash as when taking the hash of the returned plaintext
	fmt.Println(signatureAndPlaintext)

	signature, text := SplitSignz(signatureAndPlaintext)
	fmt.Println(signature)

	hashedText, _ := GetMessageHash(EncodeToBytes(text))
	got := Verify(&submitter.Keys.PublicKey, signature, hashedText)

	assert.Equal(t, true, got, "Sign and Verify failed")
}

/*
func TestSignAndVerify(t *testing.T) {
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
	str := "hello"
	//signed, _ := ecdsa.SignASN1(rand.Reader, s.keys, []byte(str))

	//sig, msg := SplitSignz(signed)
//	fmt.Println("'" + msg + "'")
	//hash, _ := GetMessageHash([]byte(msg))

	signed2 := Sign(s.keys, str)
	sig, msg := SplitSignz(signed2)
	fmt.Println(msg)
	fmt.Printf("%s %v \n", "Sig from SignASN1 test: ", signed2)
	fmt.Printf("%s %v", "Sig from Sign Test: ", sig)


	//fmt.Printf("%s%v\n", "Hash from test1:", hash)


	got := ecdsa.VerifyASN1(&s.keys.PublicKey, []byte(str), []byte(signed2))

//	fmt.Println("Sig from test:" + sig)
	assert.Equal(t, true, got, "TestSignAndVerify Failed")

} */

func TestLogging(t *testing.T) {
	
	number := ec.GetRandomInt(submitter.Keys.D)
	bytes := EncodeToBytes(number)
	Kpcs := generateSharedSecret(&pc, &submitter, nil)
	encryptedNumber := Encrypt(bytes, Kpcs)
	tree.Put("encryptedNumber", encryptedNumber)
	tree.Put("Kpcs", Kpcs)

	encryptedNumberFromTree := tree.Find("encryptedNumber")
	KpcsFromTree := tree.Find("Kpcs")

	decryptedNumber := Decrypt(encryptedNumberFromTree.value.([]byte), KpcsFromTree.value.(string))
	decodedNumber := DecodeToStruct(decryptedNumber)
	got := decodedNumber.(big.Int)

	assert.Equal(t, number, &got, "logging test failed")

}

func TestFinalMatching(t *testing.T) {

	rs := ec.GetRandomInt(submitter.Keys.D)
	rr := ec.GetRandomInt(reviewer.Keys.D)

	PaperBigInt := MsgToBigInt(EncodeToBytes(p))


	_, err := submitter.GetCommitMessagePaper(PaperBigInt, rs)
	if err != nil {
		t.Errorf("Error in GetCommitMsgPaper: %v", err)
	}

	_, err = reviewer.GetCommitMessageReviewPaper(PaperBigInt, rr)
	if err != nil {
		t.Errorf("Error in GetCommitMsgPaperR: %v", err)
	}

	fmt.Printf("%s %v \n", "Rs: ", rs)
	fmt.Printf("%s %v \n", "Rr: ", rr)
	
	fmt.Printf("%s %v \n", "RsInCommit: ", submitter.PaperCommittedValue.r)
	fmt.Printf("%s %v \n", "RrInCommit: ", reviewer.paperCommittedValue.r)

//	reviewers := []Reviewer{reviewer}
//	submitters := []Submitter{submitter}

//	finalMatching(reviewers, submitters)
}

func TestGetPaperSubmissionCommit(t *testing.T) {
	r := ec.GetRandomInt(submitter.Keys.D)
	PaperBigInt := MsgToBigInt(EncodeToBytes(p))
	commit, _  := submitter.GetCommitMessagePaper(PaperBigInt, r)
	commit2, _  := submitter.GetCommitMessagePaper(PaperBigInt, r)

	commitMsg := CommitMsg{
		commit,
		commit2,
	}

	marshalledMsg, _ := json.Marshal(commitMsg)
	signedCommitMsg := SignzAndEncrypt(submitter.Keys, marshalledMsg, "")
	tree.Put("signedCommitMsg"+submitter.UserID, signedCommitMsg)
	tree.Put("signedCommitMsg"+submitter.UserID, signedCommitMsg)

	fmt.Sprintf("%v", commit)

	fmt.Printf("%v",(pc.GetPaperSubmissionCommit(&submitter)))
}
