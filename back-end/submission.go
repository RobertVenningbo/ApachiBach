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

func Submit(s *Submitter, p *Paper) *Submitter {
	rr := GetRandomInt(s.keys.D)
	rs := GetRandomInt(s.keys.D)
	ri := GetRandomInt(s.keys.D) //TODO in the protocol description it says the submitter generates this
	
	log.Printf("\n, %s", "Generate rr from s.keys.D and storing in btree/log")
	tree.Put("Rr", rr)
	log.Printf("\n, %s", "Generate rs from s.keys.D and storing in btree/log")
	tree.Put("Rs", rs)
	log.Printf("\n, %s", "Generate ri from s.keys.D and storing in btree/log")
	tree.Put("Ri", ri)
	
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

	SignedSubmitMsg := SignzAndEncrypt(s.keys, submitMsg, "") //Signed and encrypted submit message --TODO is this what we need to return in the function?

	LoggedMessage := fmt.Sprintf("%#v", submitMsg)
	tree.Put(LoggedMessage + s.userID, SignedSubmitMsg) //Signed and encrypted paper + randomness + shared kpcs logged (step 1 done)
	log.Println(LoggedMessage + " - Encrypted Paper and Random Numbers logged")	

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

	KsString := fmt.Sprintf("%#v", s.keys.PublicKey)
	tree.Put(KsString + s.userID, s.keys.PublicKey) //Submitters public key (Ks) is revealed to all parties (step 2 done)
	log.Println("SubmitterPublicKey from submitter with userID: " + s.userID + " logged.") 

	PCsignedPaperCommit := SignzAndEncrypt(pc.keys, PaperSubmissionCommit, "")
	tree.Put("PCsignedPaperCommit" + strconv.Itoa(p.Id), PCsignedPaperCommit)
	log.Println("PCsignedPaperCommit logged - The PC signed a paper commit.") //PC signed a paper submission commit (step 3 done)

	paperList = append(paperList, *p) //List of papers, but what is it used for?? TODO

	return s //TODO why do we return a submitter?
}