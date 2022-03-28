package backend

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"strconv"
)

func Submit(s *Submitter, p *Paper) *Submitter {
	rr := GetRandomInt(s.keys.D)
	rs := GetRandomInt(s.keys.D)
	ri := GetRandomInt(s.keys.D)
	
	log.Printf("\n, %s", "Generate rr from s.keys.D and storing in btree/log")
	tree.Put("Rr", rr)
	log.Printf("\n, %s", "Generate rs from s.keys.D and storing in btree/log")
	tree.Put("Rs", rs)
	log.Printf("\n, %s", "Generate ri from s.keys.D and storing in btree/log")
	tree.Put("Ri", ri)
	
	sharedKpcs := generateSharedSecret(&pc, s, nil)  //Shared secret key between Submitter and PC (Kpcs)

	hashedPublicK := sha256.Sum256(EncodeToBytes(pc.keys.PublicKey.X)) //PC's hashed public key
	encryptedSharedKpcs := Encrypt([]byte(sharedKpcs), string(hashedPublicK[:])) //Encrypted Kpcs with PC's public key

	EncryptedPaperAndRandomness := SubmitStruct{ //Encrypted Paper and Random numbers
		Encrypt(EncodeToBytes(p), sharedKpcs),
		Encrypt(EncodeToBytes(rr), sharedKpcs),
		Encrypt(EncodeToBytes(rs), sharedKpcs),
		encryptedSharedKpcs,
	}
	LoggedMessage := fmt.Sprintf("%#v", EncryptedPaperAndRandomness)
	tree.Put(LoggedMessage, EncryptedPaperAndRandomness)
	log.Println(LoggedMessage + " - Encrypted Paper and Random Numbers logged")
	

	s.encrypted = Encrypt(EncodeToBytes(EncryptedPaperAndRandomness), s.keys.D.String()) //TODO: Do we need  this if we log it above??
	
	SubmissionSignature, _ := ecdsa.SignASN1(rand.Reader, s.keys, s.encrypted) //Entire message signed by submission private key
	SubmitterAsString := fmt.Sprintf("%#v", s)
	tree.Put(SubmitterAsString + s.userID + "SubmissionSignature", SubmissionSignature)
	log.Println(SubmitterAsString + s.userID + " " + string(SubmissionSignature) + " - message signed by submission private key")

	//submitter identity commit
	SubmitterIdenityCommit, _ := s.GetCommitMessage(ri)
	SubmitCommitAsString := fmt.Sprintf("%#v", SubmitterIdenityCommit)
	tree.Put(SubmitCommitAsString + s.userID, SubmitterIdenityCommit)
	log.Println(SubmitCommitAsString + s.userID +  " - SubmitterIdenityCommit logged")

	//paper submission commit
	PaperSubmissionCommit, _ := s.GetCommitMessagePaper(rs)
	PaperCommitAsString := fmt.Sprintf("%#v", PaperSubmissionCommit)
	tree.Put(PaperCommitAsString + strconv.Itoa(p.Id), PaperSubmissionCommit)
	log.Println(PaperCommitAsString + strconv.Itoa(p.Id) + " - PaperSubmissionCommit logged.")

	hashedIdentityCommit, _ := GetMessageHash([]byte(fmt.Sprintf("%v", s.submitterCommittedValue.CommittedValue)))
	hashedPaperCommit, _ := GetMessageHash([]byte(fmt.Sprintf("%v", s.paperCommittedValue.CommittedValue.CommittedValue)))

	SignatureSubmitterIdenityCommit, _ := ecdsa.SignASN1(rand.Reader, s.keys, hashedIdentityCommit) //Submitter Idenity Commit signed by submission private key
	tree.Put(SubmitterAsString + s.userID + "SignatureSubmitterIdentityCommit",  SignatureSubmitterIdenityCommit)
	log.Println("SignatureSubmitterIdenityCommit from userID: " + s.userID + " logged.")
	//putNextSignatureInMapSubmitter(s, signatureSubmit)

	SignaturePaperCommit, _ := ecdsa.SignASN1(rand.Reader, s.keys, hashedPaperCommit) //paper commit signed by submission private key
	tree.Put(SubmitterAsString + s.userID + "SignaturePaperCommit",  SignaturePaperCommit)
	log.Println("SignaturePaperCommit from userID: " + s.userID + " logged.")
	//putNextSignatureInMapSubmitter(s, signaturePaper)

	KsString := fmt.Sprintf("%#v", s.keys.PublicKey)
	tree.Put(KsString + s.userID, s.keys.PublicKey) //Submitters public key (Ks) is revealed to all parties
	log.Println("SubmitterPublicKey from submitter with userID: " + s.userID + " logged.") 

	hashedPaperPC, _ := GetMessageHash([]byte(fmt.Sprintf("%v", s.paperCommittedValue.CommittedValue.CommittedValue)))
	SignaturePaperCommitPC, _ := ecdsa.SignASN1(rand.Reader, pc.keys, hashedPaperPC) //PC Signs a paper commit, indicating that the paperis ready to be reviewed.
	PCsignatureAsString := fmt.Sprintf("%#v", SignaturePaperCommitPC)
	tree.Put(PCsignatureAsString + strconv.Itoa(p.Id), SignaturePaperCommitPC)
	log.Println("SignaturePaperCommitPC logged - The PC signed a paper commit.")
	//putNextSignatureInMapPC(&pc, signaturePaperPC)     

	paperList = append(paperList, *p) //List of papers, but what is it used for?? TODO

	return s
}