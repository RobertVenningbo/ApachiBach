package backend

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	ec "swag/ec"
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
	rr := ec.GetRandomInt(s.Keys.D)
	rs := ec.GetRandomInt(s.Keys.D)
	ri := ec.GetRandomInt(s.Keys.D) //TODO in the protocol description it says the submitter generates this

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

	SignedSubmitMsg := Sign(s.Keys, submitMsg)            //Signed and encrypted submit message --TODO is this what we need to return in the function?
	tree.Put("SignedSubmitMsg"+s.UserID, SignedSubmitMsg) //Signed and encrypted paper + randomness + shared kpcs logged (step 1 done)
	log.Println("SignedSubmitMsg from" + s.UserID + " - Encrypted Paper and Random Numbers logged")

	//submitter identity commit
	SubmitterBigInt := MsgToBigInt(EncodeToBytes(s))
	SubmitterIdenityCommit, _ := s.GetCommitMessage(SubmitterBigInt, ri)

	//paper submission commit
	PaperBigInt := MsgToBigInt(EncodeToBytes(p))
	PaperSubmissionCommit, _ := s.GetCommitMessagePaper(PaperBigInt, rs)

	commitMsg := CommitMsg{
		EncodeToBytes(SubmitterIdenityCommit),
		EncodeToBytes(PaperSubmissionCommit),
	}

	signedCommitMsg := SignzAndEncrypt(s.Keys, commitMsg, "")
	tree.Put("signedCommitMsg"+s.UserID, signedCommitMsg)
	log.Println("signedCommitMsg" + s.UserID + " logged") //Both commits signed and logged

	KsString := fmt.Sprintf("%v", EncodeToBytes(s.Keys.PublicKey))
	tree.Put(KsString+s.UserID, EncodeToBytes(s.Keys.PublicKey)) //Submitters public key (Ks) is revealed to all parties (step 2 done)
	log.Println("SubmitterPublicKey from submitter with userID: " + s.UserID + " logged.")

	PCsignedPaperCommit := SignzAndEncrypt(pc.Keys, PaperSubmissionCommit, "")
	tree.Put("PCsignedPaperCommit"+fmt.Sprintf("%v", (p.Id)), PCsignedPaperCommit)
	log.Println("PCsignedPaperCommit logged - The PC signed a paper commit.") //PC signed a paper submission commit (step 3 done)

	pc.allPapers = append(pc.allPapers, *p)
}

func (pc *PC) GetPaperSubmissionCommit(submitter *Submitter) ecdsa.PublicKey {

	signedCommitMsg := tree.Find("signedCommitMsg" + submitter.UserID)
	bytes := signedCommitMsg.value.([][]byte)
	_, commitMsg := SplitSignatureAndMsg(bytes)

	decodedCommitMsg := DecodeToStruct(commitMsg)

	encodedPaperCommit := decodedCommitMsg.(CommitMsg).PaperCommit
	decodedpaperCommit := DecodeToStruct(encodedPaperCommit)
	return decodedpaperCommit.(ecdsa.PublicKey)
}

func (pc *PC) GetPaperSubmissionSignature(submitter *Submitter) []byte {
	signedCommitMsg := tree.Find("signedCommitMsg" + submitter.UserID)
	bytes := signedCommitMsg.value.([][]byte)
	sig, _ := SplitSignatureAndMsg(bytes)
	return sig
}
