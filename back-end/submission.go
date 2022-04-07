package backend

import (
	"fmt"
	"log"
	"strconv"
)

type SubmitMessage struct {
	PaperAndRandomness []byte
	EncryptedKpcs	   []byte
}

type CommitMsg struct {
	IdenityCommit	[]byte
	PaperCommit		[]byte
}

func (s *Submitter)Submit(p *Paper) {
	rr := GetRandomInt(s.keys.D)
	rs := GetRandomInt(s.keys.D)
	ri := GetRandomInt(s.keys.D) //TODO in the protocol description it says the submitter generates this
	
	sharedKpcs := generateSharedSecret(&pc, s, nil)  //Shared secret key between Submitter and PC (Kpcs)

	PaperAndRandomness := SubmitStruct{ //Encrypted Paper and Random numbers
		p,
		rr,
		rs,
	}
	
	submitMsg := SubmitMessage{
		Encrypt(EncodeToBytes(PaperAndRandomness), sharedKpcs),
		Encrypt(EncodeToBytes(sharedKpcs), pc.keys.PublicKey.X.String()),
	}

	SignedSubmitMsg := Sign(s.keys, submitMsg) //Signed and encrypted submit message --TODO is this what we need to return in the function?
	tree.Put("SignedSubmitMsg" + s.userID, SignedSubmitMsg) //Signed and encrypted paper + randomness + shared kpcs logged (step 1 done)
	log.Println("SignedSubmitMsg from" + s.userID + " - Encrypted Paper and Random Numbers logged")	

	//s.encrypted = Encrypt(EncodeToBytes(EncryptedPaperAndRandomness), s.keys.D.String()) //TODO: Do we need  this if we log it above??
	
	//submitter identity commit
	SubmitterIdenityCommit, _ := s.GetCommitMessage(ri) 

	//paper submission commit
	PaperSubmissionCommit, _ := s.GetCommitMessagePaper(rs)
	
	commitMsg := CommitMsg {
		EncodeToBytes(SubmitterIdenityCommit),
		EncodeToBytes(PaperSubmissionCommit),
	}
	signedCommitMsg := SignzAndEncrypt(s.keys, commitMsg, "")
	tree.Put("signedCommitMsg" + s.userID, signedCommitMsg)
	log.Println("signedCommitMsg" + s.userID + " logged") //Both commits signed and logged 

	KsString := fmt.Sprintf("%v", EncodeToBytes(s.keys.PublicKey))
	tree.Put(KsString + s.userID, EncodeToBytes(s.keys.PublicKey)) //Submitters public key (Ks) is revealed to all parties (step 2 done)
	log.Println("SubmitterPublicKey from submitter with userID: " + s.userID + " logged.") 

	PCsignedPaperCommit := SignzAndEncrypt(pc.keys, PaperSubmissionCommit, "")
	tree.Put("PCsignedPaperCommit" + strconv.Itoa(p.Id), PCsignedPaperCommit)
	log.Println("PCsignedPaperCommit logged - The PC signed a paper commit.") //PC signed a paper submission commit (step 3 done)
}