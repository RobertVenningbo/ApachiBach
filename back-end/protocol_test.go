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
		{1, false, nil},
		{2, false, nil},
	}
	p = Paper{
		1,
		false,
		nil,

	}
	reviewer = Reviewer{
		1,
		newKeys(),
		&CommitStructPaper{},
	}
	reviewer2 = Reviewer{
		2,
		newKeys(),
		&CommitStructPaper{},
	}	
	reviewer3 = Reviewer{
		3,
		newKeys(),
		&CommitStructPaper{},
	}
	reviewer4 = Reviewer{
		4,
		newKeys(),
		&CommitStructPaper{},
	}
	submitter = Submitter{
		newKeys(),
		1, //userID
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

	signatureAndPlaintext := SignsPossiblyEncrypts(submitter.Keys, EncodeToBytes(c), "")

	signature, txt := SplitSignatureAndMsg(signatureAndPlaintext)

	hashedText, _ := GetMessageHash(txt)
	got := Verify(&submitter.Keys.PublicKey, signature, hashedText)

	assert.Equal(t, true, got, "Sign and Verify failed")
}


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

func TestGetPaperSubmissionCommit(t *testing.T) {
	r := ec.GetRandomInt(submitter.Keys.D)
	PaperBigInt := MsgToBigInt(EncodeToBytes(p))
	commit, _ := submitter.GetCommitMessagePaper(PaperBigInt, r)
	commit2, _ := submitter.GetCommitMessagePaper(PaperBigInt, r)

	commitMsg := CommitMsg{
		EncodeToBytes(commit),
		EncodeToBytes(commit2),
	}
	msg := fmt.Sprintf("signedCommitMsg%v", p.Id)
	signedCommitMsg := SignsPossiblyEncrypts(submitter.Keys, EncodeToBytes(commitMsg), "")
	
	tree.Put(msg, signedCommitMsg)

	foundCommit := pc.GetPaperSubmissionCommit(p.Id)
	assert.Equal(t, *commit, foundCommit, "TestGetPaperSubmissionCommit failed")
}

func TestGetPaperSubmissionSignature(t *testing.T) {
	r := ec.GetRandomInt(submitter.Keys.D)
	PaperBigInt := MsgToBigInt(EncodeToBytes(p))
	commit, _ := submitter.GetCommitMessagePaper(PaperBigInt, r)
	commit2, _ := submitter.GetCommitMessagePaper(PaperBigInt, r)

	commitMsg := CommitMsg{
		EncodeToBytes(commit),
		EncodeToBytes(commit2),
	}

	signedCommitMsg := SignsPossiblyEncrypts(submitter.Keys, EncodeToBytes(commitMsg), "")
	putStr := fmt.Sprintf("signedCommitMsg%v",submitter.UserID)
	tree.Put(putStr, signedCommitMsg)

	sig := pc.GetPaperSubmissionSignature(&submitter)

	_, txt := SplitSignatureAndMsg(signedCommitMsg)
	hash, _ := GetMessageHash(txt)
	got := Verify(&submitter.Keys.PublicKey, sig, hash)
	assert.Equal(t, true, got, "TestGetPaperSubmissionSignature failed")
}

func TestGetPaperAndRandomness(t *testing.T) {
	
	rr := ec.GetRandomInt(pc.Keys.D) 
    rs := ec.GetRandomInt(pc.Keys.D) 

	sharedKpcs := generateSharedSecret(&pc, &submitter, nil)
	PaperAndRandomness := SubmitStruct{ //Encrypted Paper and Random numbers
		&p,
		rr,
		rs,
	}
	submitMsg := SubmitMessage{
		Encrypt(EncodeToBytes(PaperAndRandomness), sharedKpcs),
		Encrypt(EncodeToBytes(sharedKpcs), pc.Keys.PublicKey.X.String()),
	}

	SignedSubmitMsg := SignsPossiblyEncrypts(submitter.Keys, EncodeToBytes(submitMsg), "")  //Signed and encrypted submit message --TODO is this what we need to return in the function?
	msg := fmt.Sprintf("SignedSubmitMsg%v", p.Id)
	tree.Put(msg, SignedSubmitMsg) //Signed and encrypted paper + randomness + shared kpcs logged (step 1 done)

	want := pc.GetPaperAndRandomness(p.Id)

	assert.Equal(t, PaperAndRandomness, want, "TestGetPaperAndRandomness failed")
}

func TestSubmit(t *testing.T) {
	submitter.Submit(&p)
	
}