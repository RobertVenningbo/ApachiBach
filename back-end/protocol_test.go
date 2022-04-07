package backend

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"math/big"
	"swag/ec"
	"testing"
	"github.com/stretchr/testify/assert"
)

var (
	paperListTest = []Paper{
		Paper{1, nil, false, nil},
		Paper{2, nil, false, nil},
	}
)

func TestGenerateSharedSecret(t *testing.T) {
	pc := PC{
		newKeys(),
		nil,
	}
	submitter := Submitter{
		newKeys(),
		"1", //userID
		&CommitStruct{},
		&Paper{},
		&Receiver{nil, nil},
		nil,
		nil,
	}

	got := generateSharedSecret(&pc, &submitter, nil)
	want := generateSharedSecret(&pc, &submitter, nil)
	assert.Equal(t, got, want, "Test failed")

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
	got := submitter.VerifyTrapdoorSubmitter(GetTrapdoor(submitter.receiver))

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

func TestPedersenCommitmentPaper(t *testing.T) {
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

	submitter.receiver = NewReceiver(submitter.keys)

	a := ec.GetRandomInt(submitter.keys.D)

	c, err := submitter.GetCommitMessagePaper(a)

	if err != nil {
		t.Errorf("Error in GetCommitMsg: %v", err)
	}

	SetCommitment(submitter.receiver, c)
	submittedVal, r := submitter.GetDecommitMsgPaper()

	success := submitter.receiver.CheckDecommitment(r, submittedVal)

	assert.Equal(t, true, success, "pedersen paper commitment failed")

}

/*
// ALSO ADD RECEIVER TO COMMITSTRUCT, THEN MAYBE REMOVE FROM SUBMISSIVE SUBMITTER
func TestPedersenCommitmentPaper1(t *testing.T) {
	keys := newKeys()
	submitter := Submitter{
		keys,
		"1", //userID
		&CommitStruct{},
		&Paper{
			1,
			&CommitStruct{},
			true},
		&Receiver{},
		nil,
		nil,
	}

	submitter.paperCommittedValue.CommittedValue.receiver = NewReceiver(submitter.keys)
	rec := *submitter.paperCommittedValue.CommittedValue.receiver
	a := ec.GetRandomInt(submitter.keys.D)

	c, err := submitter.GetCommitMessagePaper(a)

	if err != nil {
		t.Errorf("Error in GetCommitMsg: %v", err)
	}

	SetCommitment(&rec, c)
	submittedVal, r := submitter.GetDecommitMsgPaper()

	success := rec.CheckDecommitment(r, submittedVal)

	assert.Equal(t, true, success, "pedersen failed")

}
*/
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
	c, _ := s.GetCommitMessage(a)

	hashedMsgSubmit1, _ := GetMessageHash([]byte(fmt.Sprintf("%v", c)))

	signatureSubmit, _ := ecdsa.SignASN1(rand.Reader, s.keys, hashedMsgSubmit1) //rand.Reader idk??

	got := ecdsa.VerifyASN1(&s.keys.PublicKey, hashedMsgSubmit1, signatureSubmit) //testing

	assert.Equal(t, true, got, "Sign and Verify failed")
}

// func TestSubmit(t *testing.T) {
// 	keys := newKeys()
// 	p := Paper{
// 		1,
// 		&CommitStruct{},
// 		true,
// 		nil,
// 	}
// 	s := Submitter{
// 		keys,
// 		"1", //userID
// 		&CommitStruct{},
// 		&p,
// 		&Receiver{},
// 		nil,
// 		nil,
// 	}
// 	pc := PC{
// 		keys,
// 		nil,
// 	}

// 	got := Submit(&s, &p)

// 	//fmt.Println(pc)
// 	pc.signatureMap = nil //need this or it isn't used

// 	assert.Equal(t, got, got, "Submit failed") //Can't compare got to got, this test is useless
// }

func TestAssignPapersGetPaperList(t *testing.T) {
	pc := PC{
		newKeys(),
		nil,
	}
	reviewer1 := Reviewer{
		"reviewer1",
		newKeys(),
		nil,
		map[int][]byte{},
		nil,
		nil,
		nil,
		nil,
	}
	reviewer2 := Reviewer{
		"reviewer2",
		newKeys(),
		nil,
		map[int][]byte{},
		nil,
		nil,
		nil,
		nil,
	}
	assignPapers(&pc, []Reviewer{reviewer1, reviewer2}, paperListTest)
	got := getPaperList(&pc, &reviewer1)
	want := paperListTest
	assert.Equal(t, got, want, "TestAssignPapersGetPaperList failed")

}

/*
func TestSchnorrProof(t *testing.T) {
	p := Paper{
		1,
		&CommitStruct{},
		true,
		nil,
	}
	reviewer := Reviewer{
		"reviewer",
		newKeys(),
		nil,
		map[int][]byte{},
		nil,
		&p,
		nil,
	}
	submitter := Submitter{
		newKeys(),
		"1", //userID
		&CommitStruct{},
		&p,
		&Receiver{},
		nil,
		nil,
	}

	a := ec.GetRandomInt(submitter.keys.D)
	submitter.GetCommitMessagePaper(a)

	b := ec.GetRandomInt(reviewer.keys.D)
	reviewer.GetCommitMessageReviewPaper(b)

	proof := CreateProof(submitter.keys, reviewer.keys)

	got := VerifyProof(proof)

	want := true

	assert.Equal(t, want, got, "Proof failed")
}
*/

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
	keys := newKeys()
	c := CommitStruct{
		&keys.PublicKey,
		nil,
		nil}
	p := Paper{
		1,
		&c,
		true,
		nil,
	}
	EncodedStruct := EncodeToBytes(p)
	DecodedStruct := DecodeToStruct1(EncodedStruct, Paper{})
	assert.Equal(t, DecodedStruct, p, "Test failed")

}

func TestVerifyMethod(t *testing.T) {
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
	c, _ := s.GetCommitMessage(a)

	signatureAndPlaintext := Sign(s.keys, c) //TODO; current bug is that this hash within this function is not the same hash as when taking the hash of the returned plaintext
	fmt.Println(signatureAndPlaintext)

	signature, text := SplitSignz(signatureAndPlaintext)
	fmt.Println(signature)

	hashedText, _ := GetMessageHash(EncodeToBytes(text))
	got := Verify(&s.keys.PublicKey, signature, hashedText)

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
	tree = NewTree(DefaultMinItems)
	pc := PC{
		newKeys(),
		nil,
	}
	s := Submitter{
		newKeys(),
		"1", //userID
		&CommitStruct{},
		&Paper{},
		&Receiver{},
		nil,
		nil,
	}
	
	number := GetRandomInt(s.keys.D)
	bytes := EncodeToBytes(number)
	Kpcs := generateSharedSecret(&pc, &s, nil)
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

