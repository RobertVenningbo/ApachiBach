package backend

import (
	"fmt"
	"swag/model"
)

func (s *Submitter) ClaimPaper() { //step 19
	paper := s.PaperCommittedValue.Paper
	ri := s.SubmitterCommittedValue.Val
	//rii := s.GetPrivateBigInt("ri") new way of getting something private
	msg := ClaimMessage{
		paper,
		s,
		ri,
	}
	str := fmt.Sprintf("Submitter %v, claims paper by revealing paper and ri.", s.UserID)
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

func (pc *PC) ConfirmOwnership(s *Submitter) { //step 20

	sig, claim := GetClaimMessage(s)
	SPK := pc.GetSubmitterPK(s.UserID)
	claimBytes := EncodeToBytes(claim)
	hash, _ := GetMessageHash(claimBytes)
	isLegit := Verify(&SPK, sig, hash)
	if !isLegit {
		fmt.Printf("PC couldn't verify signature from submitter %v \n", s.UserID)
	}
	/*perhaps verify some of the properties of claimMsg*/

	signature := SignsPossiblyEncrypts(pc.Keys, claimBytes, "")

	putStr := fmt.Sprintf("PC confirms the ownership of paper, %v, to submitter: %v", claim.Paper.Id, s.UserID)
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

func GetConfirmMessage(s *Submitter) ([]byte, ClaimMessage) { //returns signature from the submitter and the ClaimMessage
	//Probably needs error handling for when checking claim
	// message for a submitter which haven't submitted one
	_, claim := GetClaimMessage(s)
	getStr := fmt.Sprintf("PC confirms the ownership of paper, %v, to submitter: %v", claim.Paper.Id, s.UserID)
	item := Trae.Find(getStr)

	claimMsgBytes := item.value.([][]byte)
	sig, encoded := SplitSignatureAndMsg(claimMsgBytes)
	claimMsg := DecodeToStruct(encoded).(ClaimMessage)

	return sig, claimMsg
}

func GetClaimMessage(s *Submitter) ([]byte, ClaimMessage) { //returns signature from the submitter and the ClaimMessage
	//Probably needs error handling for when checking claim
	// message for a submitter which haven't submitted one

	getStr := fmt.Sprintf("Submitter %v, claims paper by revealing paper and ri.", s.UserID)
	item := Trae.Find(getStr)
	if item == nil {
		CheckStringAgainstDB(getStr)
		item = Trae.Find(getStr)
	}

	claimMsgBytes := item.value.([]byte)

	var sigmsg model.Log
	model.GetLogMsgByMsg(&sigmsg, getStr)
	sig := sigmsg.Signature

	claimMsg := DecodeToStruct(claimMsgBytes).(ClaimMessage)

	return sig, claimMsg
}
