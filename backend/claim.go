package backend

import (
	"fmt"
	"swag/model"
)


func (s *Submitter) ClaimPaper(pId int) { //step 19
	paper := s.PaperCommittedValue.Paper
	ri := s.SubmitterCommittedValue.Val
	//rii := s.GetPrivateBigInt("ri") new way of getting something private
	msg := ClaimMessage{
		paper,
		s,
		ri,
	}
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
			State: 20,
			LogMsg: putStr,
			FromUserID: 4000,
			Value: signature[1],
			Signature: signature[0],
		}
		model.CreateLogMsg(&logmsg)
		Trae.Put(putStr, signature)
    }
}

func (pc *PC) GetConfirmMessage(pId int) ([]byte, *ClaimMessage) { //returns signature from the submitter and the ClaimMessage
	claim := pc.GetClaimMessage(pId)
	getStr := fmt.Sprintf("PC confirms the ownership of paper, %v, to submitter: %v", claim.Paper.Id, claim.Submitter.UserID)
	item := Trae.Find(getStr)
	if item == nil {
		CheckStringAgainstDB(getStr)
		item = tree.Find(getStr)
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

	isLegit := VerifySignature(getStr, claimMsgBytes, &SPK)
	if !isLegit {
		fmt.Printf("PC couldn't verify signature from submitter who submitted paper %v \n", pId)
	} else {
		fmt.Printf("PC verifies signature from submitter who submitted %v \n", pId)
	}

	claimMsg := DecodeToStruct(claimMsgBytes).(ClaimMessage)

	return &claimMsg
}
