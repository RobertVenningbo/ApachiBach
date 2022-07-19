package backend

import (
	"fmt"
	"swag/model"
)

//confirm this isnt being called for rejected submitters, they dont need to claim anything.
// 18/07/2022: can confirm this is being called despite rejected, i.e. submitters claims their paper despite it being rejected
// also maybe bring emphasis on the submitter also revealing themselves, idk?
func (s *Submitter) ClaimPaper(pId int) { //step 19
	paper := s.PaperCommittedValue.Paper
	ri := s.SubmitterCommittedValue.Val
	msg := ClaimMessage{
		paper,
		s,
		ri,
	}
	SubmitterBigInt := MsgToBigInt(EncodeToBytes(s.UserID))
	fmt.Println("SubmitterBigInt in ClaimPaper" + fmt.Sprint(SubmitterBigInt)) //check. Probably need to reveal its own ID as bigInt for checking commitment havent changed. Since we do commitment with this bigInt
	str := fmt.Sprintf("Submitter claims paper %v by revealing paper and ri.", pId)
	signature := SignsPossiblyEncrypts(s.Keys, EncodeToBytes(msg), "")
	logmsg := model.Log{
		State:      19,
		LogMsg:     str,
		FromUserID: s.UserID,
		Value:      signature[1],
		Signature:  signature[0],
	}
	model.CreateLogMsg(&logmsg)
	Trae.Put(str, signature)
}

func (pc *PC) ConfirmOwnership(pId int) { //step 20
	if pc.GetClaimMessage(pId) == nil {
		return
	} else {
		claim := pc.GetClaimMessage(pId)

		claimBytes := EncodeToBytes(claim)
		signature := SignsPossiblyEncrypts(pc.Keys, claimBytes, "")

		putStr := fmt.Sprintf("PC confirms the ownership of paper, %v, to submitter: %v", claim.Paper.Id, claim.Submitter.UserID)

		logmsg := model.Log{
			State:      20,
			LogMsg:     putStr,
			FromUserID: 4000,
			Value:      signature[1],
			Signature:  signature[0],
		}
		model.CreateLogMsg(&logmsg)
		Trae.Put(putStr, signature)
	}
}

//Not used -- yet?
func (pc *PC) GetConfirmMessage(pId int) ([]byte, *ClaimMessage) { //returns signature from the submitter and the ClaimMessage
	claim := pc.GetClaimMessage(pId)
	getStr := fmt.Sprintf("PC confirms the ownership of paper, %v, to submitter: %v", claim.Paper.Id, claim.Submitter.UserID)
	item := Trae.Find(getStr)
	if item == nil {
		CheckStringAgainstDB(getStr)
		item = Trae.Find(getStr)
	}

	claimMsgBytes := item.value.([][]byte)
	sig, encoded := SplitSignatureAndMsg(claimMsgBytes)
	claimMsg := DecodeToStruct(encoded).(ClaimMessage)

	return sig, &claimMsg
}

func (pc *PC) GetClaimMessage(pId int) *ClaimMessage {
	//Probably needs error handling for when checking claim
	// message for a submitter which haven't submitted one

	getStr := fmt.Sprintf("Submitter claims paper %v by revealing paper and ri.", pId)
	item := Trae.Find(getStr)
	if item == nil {
		CheckStringAgainstDB(getStr)
		item = Trae.Find(getStr)
	}

	if item == nil {
		fmt.Println("Submitter hasn't claimed ownership of paper yet.")
		return nil
	}
	SPK := pc.GetPaperSubmitterPK(pId)
	claimMsgBytes := item.value.([]byte)
	hash, _ := GetMessageHash(claimMsgBytes)
	var sigmsg model.Log
	model.GetLogMsgByMsg(&sigmsg, getStr)
	sig := sigmsg.Signature
	isLegit := Verify(&SPK, sig, hash)
	if !isLegit {
		fmt.Printf("PC couldn't verify signature from submitter who submitted paper %v \n", pId)
	} else {
		fmt.Printf("PC verifies signature from submitter who submitted %v \n", pId)
	}

	claimMsg := DecodeToStruct(claimMsgBytes).(ClaimMessage)

	return &claimMsg
}
