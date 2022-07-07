package backend

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	"swag/model"
)

func (s *Submitter) Submit(p *Paper) {
	InitLocalPC()

	Trae = DatabaseToTree()
	s.PaperCommittedValue.Paper = p
	curve1 := elliptic.P256()
	curve := curve1.Params()

	rr, _ := rand.Int(rand.Reader, curve.N)
	rs, _ := rand.Int(rand.Reader, curve.N)
	ri, _ := rand.Int(rand.Reader, curve.N)

	sharedKpcs := GenerateSharedSecret(&Pc, s, nil) //Shared secret key between Submitter and PC (Kpcs)

	PaperAndRandomness := SubmitStruct{ //Encrypted Paper and Random numbers
		p,
		rr,
		rs,
	}
	submitMsg := SubmitMessage{
		Encrypt(EncodeToBytes(PaperAndRandomness), sharedKpcs),
		Encrypt([]byte(sharedKpcs), Pc.Keys.PublicKey.X.String()),
	}

	SignedSubmitMsg := SignsPossiblyEncrypts(s.Keys, EncodeToBytes(submitMsg), "") //Signed and encrypted submit message 
	msg := fmt.Sprintf("SignedSubmitMsg%v", p.Id)

	logmsg := model.Log{
		State:      1,
		LogMsg:     msg,
		FromUserID: s.UserID,
		Value:      SignedSubmitMsg[1],
		Signature:  SignedSubmitMsg[0],
	}
	model.CreateLogMsg(&logmsg)

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

	logmsg2 := model.Log{
		State:      2,
		LogMsg:     msg,
		FromUserID: s.UserID,
		Value:      signedCommitMsg[1],
		Signature:  signedCommitMsg[0],
	}
	model.CreateLogMsg(&logmsg2)
	
	//TODO: Why are we logging KsString and PK?? they are the same thing
	KsString := fmt.Sprintf("SubmitterPublicKey with P%v", p.Id)
	logmsg3 := model.Log{
		State:      2,
		LogMsg:     KsString,
		FromUserID: s.UserID,
		Value:      EncodeToBytes(&s.Keys.PublicKey),
	}
	model.CreateLogMsg(&logmsg3)

	PK := fmt.Sprintf("SubmitterPublicKey %v", s.UserID)
	logmsg4 := model.Log{
		State:      2,
		LogMsg:     PK,
		FromUserID: s.UserID,
		Value:      EncodeToBytes(&s.Keys.PublicKey),
	}
	model.CreateLogMsg(&logmsg4)

	PCsignedPaperCommit := SignsPossiblyEncrypts(Pc.Keys, EncodeToBytes(PaperSubmissionCommit), "")
	str := fmt.Sprintf("PCsignedPaperCommit%v", p.Id)
	logmsg5 := model.Log{
		State:      3,
		LogMsg:     str,
		FromUserID: 4000,
		Value:      PCsignedPaperCommit[1],
		Signature:  PCsignedPaperCommit[0],
	}
	model.CreateLogMsg(&logmsg5)

	s.StorePrivateBigInt(ri, "ri")
}

func (s *Submitter) StorePrivateBigInt(i *big.Int, txt string) {
	str := fmt.Sprintf("Submitter %v privately stores a %s", s.UserID, txt)
	log.Println(str)
	Trae.Put(str, Encrypt(EncodeToBytes(i), s.Keys.D.String()))
}

func (s *Submitter) GetPrivateBigInt(txt string) *big.Int {
	str := fmt.Sprintf("Submitter %v privately stores a %s", s.UserID, txt)
	log.Println("GETTING:" + str)
	item := Trae.Find(str)
	if item == nil {
		CheckStringAgainstDB(str)
		item = Trae.Find(str)
	}
	bytes := item.value.([]byte)
	encodedBigInt := Decrypt(bytes, s.Keys.D.String())
	decodedBigInt := DecodeToStruct(encodedBigInt).(*big.Int)
	return decodedBigInt
}

func (pc *PC) GetPaperSubmitterPK(pId int) ecdsa.PublicKey {
	KsString := fmt.Sprintf("SubmitterPublicKey with P%v", pId)
	item := Trae.Find(KsString)
	if item == nil {
		CheckStringAgainstDB(KsString)
		item = Trae.Find(KsString)
	}
	
	decodedPK := DecodeToStruct(item.value.([]byte))
	PK := decodedPK.(ecdsa.PublicKey)
	return PK
}

func (pc *PC) GetSubmitterPK(sUserID int) ecdsa.PublicKey {
	PK := fmt.Sprintf("SubmitterPublicKey %v", sUserID)
	
	item := Trae.Find(PK)
	if item == nil {
		CheckStringAgainstDB(PK)
		item = Trae.Find(PK)
	}
	decodedPK := DecodeToStruct(item.value.([]byte))
	REALPK := decodedPK.(ecdsa.PublicKey)

	return REALPK
}

func (pc *PC) GetPaperSubmissionCommit(id int) ecdsa.PublicKey {
	msg := fmt.Sprintf("signedCommitMsg%v", id)
	signedCommitMsg := Trae.Find(msg)
	if signedCommitMsg == nil {
		CheckStringAgainstDB(msg)
		signedCommitMsg = Trae.Find(msg)
	}

	bytes := signedCommitMsg.value.([]byte)
	decodedCommitMsg := DecodeToStruct(bytes)

	encodedPaperCommit := decodedCommitMsg.(CommitMsg).PaperCommit
	decodedpaperCommit := DecodeToStruct(encodedPaperCommit)


	SPK := pc.GetSubmitterPK(id)

    hash, _  := GetMessageHash(bytes)
	var sigmsg model.Log 
	model.GetLogMsgByMsg(&sigmsg, msg)
	sig := sigmsg.Signature
	isLegit := Verify(&SPK, sig, hash)
	if !isLegit {
		fmt.Printf("\nPC couldn't verify signature getting PaperSubmissionCommit for P%v \n", id)
	} else {
		fmt.Printf("\nPC verifies signature getting PaperSubmissionCommit for P%v \n", id)
	}

	return decodedpaperCommit.(ecdsa.PublicKey)
}

func (pc *PC) GetPaperSubmissionSignature(submitter *Submitter) []byte { //Not used for anything
	putStr := fmt.Sprintf("signedCommitMsg%v", submitter.UserID)
	signedCommitMsg := Trae.Find(putStr)
	if signedCommitMsg.value == nil {
		CheckStringAgainstDB(putStr)
		signedCommitMsg = Trae.Find(putStr)
	}
	bytes := signedCommitMsg.value.([][]byte)
	sig, _ := SplitSignatureAndMsg(bytes)
	return sig
}

func (pc *PC) GetPaperAndRandomness(pId int) SubmitStruct {
	msg := fmt.Sprintf("SignedSubmitMsg%v", pId)
	item := Trae.Find(msg)
	if item == nil {
		CheckStringAgainstDB(msg)
		item = Trae.Find(msg)
	}
	bytes := item.value.([]byte)
	decodedSubmitMessage := DecodeToStruct(bytes)
	submitMessage := decodedSubmitMessage.(SubmitMessage)
	kpcs := Decrypt(submitMessage.EncryptedKpcs, pc.Keys.X.String())

	decryptedPaperAndRandomness := Decrypt(submitMessage.PaperAndRandomness, string(kpcs))
	PaperAndRandomness := DecodeToStruct(decryptedPaperAndRandomness).(SubmitStruct)

	SPK := pc.GetSubmitterPK(pId)
	hash, _  := GetMessageHash(bytes)
	var sigmsg model.Log 
	model.GetLogMsgByMsg(&sigmsg, msg)
	sig := sigmsg.Signature
	isLegit := Verify(&SPK, sig, hash)
	if !isLegit {
		fmt.Printf("\nPC couldn't verify signature getting PaperAndRandomness for P%v \n", pId)
	} else {
		fmt.Printf("\nPC verifies signature getting PaperAndRandomness for P%v \n", pId)
	}
	return PaperAndRandomness
}
