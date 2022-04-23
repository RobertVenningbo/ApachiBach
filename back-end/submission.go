package backend

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
)

type SubmitMessage struct {
	PaperAndRandomness []byte
	EncryptedKpcs      []byte
}

type CommitMsg struct {
	IdenityCommit []byte
	PaperCommit   []byte
}

func (s *Submitter) Submit(p *Paper) {

	curve1 := elliptic.P256()
	curve := curve1.Params()

	rr, _ := rand.Int(rand.Reader, curve.N)
	rs, _ := rand.Int(rand.Reader, curve.N)
	ri, _ := rand.Int(rand.Reader, curve.N)

	sharedKpcs := generateSharedSecret(&pc, s, nil) //Shared secret key between Submitter and PC (Kpcs)

	PaperAndRandomness := SubmitStruct{ //Encrypted Paper and Random numbers
		p,
		rr,
		rs,
	}
	submitMsg := SubmitMessage{
		Encrypt(EncodeToBytes(PaperAndRandomness), sharedKpcs),
		Encrypt(EncodeToBytes(sharedKpcs), pc.Keys.PublicKey.X.String()),
	}

	SignedSubmitMsg := SignsPossiblyEncrypts(s.Keys, EncodeToBytes(submitMsg), "") //Signed and encrypted submit message --TODO is this what we need to return in the function?
	msg := fmt.Sprintf("SignedSubmitMsg%v", p.Id)
	tree.Put(msg, SignedSubmitMsg) //Signed and encrypted paper + randomness + shared kpcs logged (step 1 done)
	log.Println("SignedSubmitMsg from" + fmt.Sprintf("%v", p.Id) + " - Encrypted Paper and Random Numbers logged")

	//submitter identity commit
	SubmitterBigInt := MsgToBigInt(EncodeToBytes(s))
	SubmitterIdenityCommit, err := s.GetCommitMessage(SubmitterBigInt, ri)
	if err != nil {
		fmt.Printf("Error in submission.go GetCommitMsg: %v\n", err)
	}

	//paper submission commit
	PaperBigInt := MsgToBigInt(EncodeToBytes(p.Id))
	PaperSubmissionCommit, err := s.GetCommitMessagePaper(PaperBigInt, rs)
	if err != nil {
		fmt.Printf("Error in submission.go GetCommitMsgPaper: %v\n", err)
	}

	commitMsg := CommitMsg{
		EncodeToBytes(SubmitterIdenityCommit),
		EncodeToBytes(PaperSubmissionCommit),
	}

	signedCommitMsg := SignsPossiblyEncrypts(s.Keys, EncodeToBytes(commitMsg), "")
	msg = fmt.Sprintf("signedCommitMsg%v", p.Id)
	tree.Put(msg, signedCommitMsg)
	log.Println(msg + " logged") //Both commits signed and logged

	KsString := fmt.Sprintf("SubmitterPublicKey with P%v", p.Id)
	tree.Put(KsString, EncodeToBytes(&s.Keys.PublicKey)) //Submitters public key (Ks) is revealed to all parties (step 2 done)
	log.Println(KsString + " logged.")

	PCsignedPaperCommit := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(PaperSubmissionCommit), "")
	tree.Put("PCsignedPaperCommit"+fmt.Sprintf("%v", (p.Id)), PCsignedPaperCommit)
	log.Println("PCsignedPaperCommit logged - The PC signed a paper commit.") //PC signed a paper submission commit (step 3 done)

	pc.allPapers = append(pc.allPapers, p)
}

func (pc *PC) GetPaperSubmitterPK(pId int) ecdsa.PublicKey {
	KsString := fmt.Sprintf("SubmitterPublicKey with P%v", pId)
	item := tree.Find(KsString)
	decodedPK := DecodeToStruct(item.value.([]byte))
	PK := decodedPK.(ecdsa.PublicKey)

	return PK
}

func (pc *PC) GetPaperSubmissionCommit(id int) ecdsa.PublicKey {
	msg := fmt.Sprintf("signedCommitMsg%v", id)
	signedCommitMsg := tree.Find(msg)
	bytes := signedCommitMsg.value.([][]byte)
	_, commitMsg := SplitSignatureAndMsg(bytes)

	decodedCommitMsg := DecodeToStruct(commitMsg)

	encodedPaperCommit := decodedCommitMsg.(CommitMsg).PaperCommit
	decodedpaperCommit := DecodeToStruct(encodedPaperCommit)
	return decodedpaperCommit.(ecdsa.PublicKey)
}

func (pc *PC) GetPaperSubmissionSignature(submitter *Submitter) []byte {
	putStr := fmt.Sprintf("signedCommitMsg%v",submitter.UserID)
	signedCommitMsg := tree.Find(putStr)
	bytes := signedCommitMsg.value.([][]byte)
	sig, _ := SplitSignatureAndMsg(bytes)
	return sig
}

func (pc *PC) GetPaperAndRandomness(pId int) SubmitStruct {
	msg := fmt.Sprintf("SignedSubmitMsg%v", pId)
	item := tree.Find(msg)
	_, encodedSubmitMessage := SplitSignatureAndMsg(item.value.([][]byte))
	decodedSubmitMessage := DecodeToStruct(encodedSubmitMessage)
	submitMessage := decodedSubmitMessage.(SubmitMessage)
	encodedKpcs := Decrypt(submitMessage.EncryptedKpcs, pc.Keys.X.String())
	kpcs := DecodeToStruct(encodedKpcs).(string)

	decryptedPaperAndRandomness := Decrypt(submitMessage.PaperAndRandomness, kpcs)
	PaperAndRandomness := DecodeToStruct(decryptedPaperAndRandomness).(SubmitStruct)
	return PaperAndRandomness

}
