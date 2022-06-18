package backend

import (
	"crypto/ecdsa"
	_ "crypto/rand"
	"fmt"
	"log"
	ec "swag/ec"
	"swag/model"
)

//step 4
func (pc *PC) DistributePapers(reviewerSlice []Reviewer, paperSlice []*Paper) {
	for r := range reviewerSlice {
		Kpcr := pc.GetKPCRFromLog(reviewerSlice[r].UserID) //Shared key between R and PC (Kpcr) -
		for p := range paperSlice {
			SignedAndEncryptedPaper := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(paperSlice[p]), Kpcr)
			_, val := SplitSignatureAndMsg(SignedAndEncryptedPaper)
			fmt.Sprintln(val)
			msg := fmt.Sprintf("SignedAndEncryptedPaper P%v for R%v", paperSlice[p].Id, reviewerSlice[r].UserID)
			fmt.Println(msg)
			Trae.Put(msg, SignedAndEncryptedPaper)
			logmsg := model.Log{
				State:      4,
				LogMsg:     msg,
				FromUserID: 4000,
				Value:      SignedAndEncryptedPaper[1],
				Signature:  SignedAndEncryptedPaper[0],
			}
			model.CreateLogMsg(&logmsg)
		}
	}
}

//paperSlice is only there for getting len(paperSlice) for forloop.
//Gets all papers for each reviewer from log.
//Expected to be called for every reviewer when reviewers want to see list of all papers on frontend.
func (r *Reviewer) GetPapersReviewer(paperSlice []*Paper) []*Paper {
	Kpcr := GenerateSharedSecret(&Pc, nil, r)
	pList := []*Paper{}
	for p := range paperSlice {
		GetMsg := fmt.Sprintf("SignedAndEncryptedPaper P%v for R%v", paperSlice[p].Id, r.UserID)
		fmt.Println("GetMsg: " + GetMsg)
		treeItem := Trae.Find(GetMsg)

		if treeItem == nil {
			CheckStringAgainstDB(GetMsg)
			treeItem = Trae.Find(GetMsg)
		}

		bytes := treeItem.value.([]byte)
		decrypted := Decrypt(bytes, Kpcr)
		decoded := DecodeToStruct(decrypted)
		paper := decoded.(Paper)
		pList = append(pList, &paper)
	}
	return pList
}

func (pc *PC) GetBiddedPaper(id int) *PaperBid { 
	Kpcr := pc.GetKPCRFromLog(id)
	msg := fmt.Sprintf("EncryptedSignedBids %v", id)
	EncryptedSignedBid := Trae.Find(msg)

	if EncryptedSignedBid == nil {
		CheckStringAgainstDB(msg)
		EncryptedSignedBid = Trae.Find(msg)
	}
	if EncryptedSignedBid == nil {
		return &PaperBid{
			nil,
			&Reviewer{ 
				UserID: -1,
				Keys: nil,
				PaperCommittedValue: nil,
			},
		}
	}
	bytes := EncryptedSignedBid.value.([]byte)
	decrypted := Decrypt(bytes, Kpcr)
	decoded := DecodeToStruct(decrypted)
	bid := decoded.(PaperBid)
	return &bid
}

func (pc *PC) GetAllBids() []*PaperBid {
	var users []model.User
	model.GetReviewers(&users)
	var bidList []*PaperBid
	for _, u := range users {
		bidList = append(bidList, pc.GetBiddedPaper(u.Id))
	}
	return bidList
}


func (r *Reviewer) GetBiddedPaper() *PaperBid { // possibly
	Kpcr := GenerateSharedSecret(&Pc, nil, r)
	msg := fmt.Sprintf("EncryptedSignedBids %v", r.UserID)
	EncryptedSignedBid := Trae.Find(msg)

	if EncryptedSignedBid == nil {
		CheckStringAgainstDB(msg)
		EncryptedSignedBid = Trae.Find(msg)
	}

	bytes := EncryptedSignedBid.value.([][]byte)
	_, enc := SplitSignatureAndMsg(bytes)
	decrypted := Decrypt([]byte(enc), Kpcr)
	decoded := DecodeToStruct(decrypted)
	bid := decoded.(PaperBid)
	fmt.Println(bid.Paper.Id) //only for testing
	return &bid
}

func (r *Reviewer) MakeBid(pap *Paper) *PaperBid {
	return &PaperBid{
		pap,
		r,
	}
}

//step 5
func (r *Reviewer) SignBidAndEncrypt(p *Paper) { //set encrypted bid list
	bid := r.MakeBid(p)
	Kpcr := GenerateSharedSecret(&Pc, nil, r) //Shared secret key between R and PC
	EncryptedSignedBid := SignsPossiblyEncrypts(r.Keys, EncodeToBytes(bid), Kpcr)
	sig, msgvalue := SplitSignatureAndMsg(EncryptedSignedBid)
	msg := fmt.Sprintf("EncryptedSignedBids %v", r.UserID)

	logMsg := model.Log{
		State:      5, //check
		LogMsg:     msg,
		FromUserID: bid.Reviewer.UserID,
		Value:      msgvalue,
		Signature:  sig,
	}

	model.CreateLogMsg(&logMsg)
	Trae.Put(msg, EncryptedSignedBid)
	log.Println(msg + "logged.")
}

// func (pc *PC) ReplaceWithBids(reviewerSlice []*Reviewer) ([]*Paper, []*PaperBid) { //Not used, maybe delete
// 	bidList := []*PaperBid{}
// 	for i := range reviewerSlice { //loop to get list of all bidded papers
// 		p := reviewerSlice[i].GetBiddedPaper()
// 		bidList = append(bidList, p)
// 	}

// 	for _, p := range pc.AllPapers {
// 		for _, b := range bidList {
// 			if p.Id == b.Paper.Id {
// 				p = b.Paper

// 			}
// 		}
// 	}
// 	return pc.AllPapers, bidList
// }

func (pc *PC) DeliverAssignedPaper() { //Unfortunately, reviewers get access to the entire paper reviewerlist this way
	for _, p := range pc.AllPapers {
		for _, r := range p.ReviewerList {
			str := fmt.Sprintf("DeliveredPaperForR%v", r.UserID)
			Kpcr := pc.GetKPCRFromLog(r.UserID)
			EncryptedPaper := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(p), Kpcr)
			logmsg := model.Log{
				State: 7,
				LogMsg: str,
				FromUserID: 4000,
				Value: EncryptedPaper[1],
				Signature: EncryptedPaper[0],
			}
			model.CreateLogMsg(&logmsg)
		}
	}
}

func (pc *PC) AssignPaper(bidList[]*PaperBid) {
	

	reviewersBidsTaken := []Reviewer{}

	for _, bid := range bidList {
		for _, p := range pc.AllPapers {
			if p.Id == bid.Paper.Id {
				if !p.Selected {
					if bid.Reviewer.PaperCommittedValue == nil {
						bid.Reviewer.PaperCommittedValue = &CommitStructPaper{}
					}
					p.Selected = true
					p.ReviewerList = append(p.ReviewerList, *bid.Reviewer)
					break
				} else {
					reviewersBidsTaken = append(reviewersBidsTaken, *bid.Reviewer)
					break
				}
			}
		}
	}
	for i, r := range reviewersBidsTaken {
		x := false
		if r.PaperCommittedValue == nil {
			r.PaperCommittedValue = &CommitStructPaper{}
		}
		for _, p := range pc.AllPapers {
			if !p.Selected {
				x = true
				p.Selected = true
				reviewer := &r
				p.ReviewerList = append(p.ReviewerList, *reviewer)
				break
			}
		}
		if x {
			reviewersBidsTaken[i].UserID = -1
			x = false
		}
	}
	for _, r := range reviewersBidsTaken {
		if ((r.PaperCommittedValue == nil) || (r.PaperCommittedValue == &CommitStructPaper{})) && (r.UserID != -1) {
			r.PaperCommittedValue = &CommitStructPaper{}
			for _, p := range pc.AllPapers {
				p.Selected = true
				p.ReviewerList = append(p.ReviewerList, r)
				break
			}
		}
	}
	
	//pc.SetReviewersPaper(reviewerSlice)
}

// This method is a little messy however it is not expected to be called on a lot of entities.
// **Finds every assigned reviewer for every paper and makes it bidirectional, such that a reviewer also has a reference to a paper**
// **Basically a fast reversal of assignPaper in terms of being bidirectional**
func (pc *PC) SetReviewersPaper(reviewerList []*Reviewer) {
	for _, p := range pc.AllPapers {
		for _, r := range p.ReviewerList {
			for _, r1 := range reviewerList {
				if r.UserID == r1.UserID {
					if r1.PaperCommittedValue == nil {
						r1.PaperCommittedValue = &CommitStructPaper{}
					}
					r1.PaperCommittedValue.Paper = p
				}
			}
		}
	}
}
func (pc *PC) MatchPapers() {
	for _, p := range pc.AllPapers {
		fmt.Println("in match papers")
		PaperBigInt := MsgToBigInt(EncodeToBytes(p.Id))

		nonce_r := ec.GetRandomInt(pc.Keys.D)

		reviewerKeyList := []ecdsa.PublicKey{}
		for _, r := range p.ReviewerList {
			reviewerKeyList = append(reviewerKeyList, r.Keys.PublicKey)
		}
		rr := pc.GetPaperAndRandomness(p.Id).Rr
		commit, err := pc.GetCommitMessagePaperPC(PaperBigInt, rr)
		if err != nil {
			log.Panic("matchPaperz error")
		}
		reviewStruct := ReviewSignedStruct{
			commit,
			reviewerKeyList,
			nonce_r,
		}

		signature := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(reviewStruct), "")

		msg := fmt.Sprintf("ReviewSignedStruct with P%v", p.Id)
		logmsg := model.Log{
			State: 7,
			LogMsg: msg,
			FromUserID: 4000,
			Value: signature[1],
			Signature: signature[0],
		}
		model.CreateLogMsg(&logmsg)
		Trae.Put(msg, signature[1])

		nizkBool := pc.SupplyNIZK(p)
		if !nizkBool {
			fmt.Println("NIZK Failed in MatchPapers")
		}
	}
}

func (pc *PC) GetReviewSignedStruct(pId int) ReviewSignedStruct {
	ret := ReviewSignedStruct{}
	msg := fmt.Sprintf("ReviewSignedStruct with P%v", pId)
	item := Trae.Find(msg)
	if item == nil {
		CheckStringAgainstDB(msg)
		item = Trae.Find(msg)
	}
	bytes := item.value.([]byte)
	decodedStruct := DecodeToStruct(bytes)
	ret = decodedStruct.(ReviewSignedStruct)
	fmt.Printf("%s %v \n", "Review Commit: ", ret.Commit)

	return ret
}

func (reviewer *Reviewer) GetReviewSignedStruct(pId int) ReviewSignedStruct {
	ret := ReviewSignedStruct{}
	msg := fmt.Sprintf("ReviewSignedStruct with P%v", pId)
	item := Trae.Find(msg)
	if item == nil {
		CheckStringAgainstDB(msg)
		item = Trae.Find(msg)
	}
	bytes := item.value.([]byte)
	decodedStruct := DecodeToStruct(bytes)
	ret = decodedStruct.(ReviewSignedStruct)
	fmt.Printf("%s %v \n", "Review Commit: ", ret.Commit)

	return ret
}

func (pc *PC) SupplyNIZK(p *Paper) bool {
	works := false                                             //for testing
	paperSubmissionCommit := pc.GetPaperSubmissionCommit(p.Id) //PaperSubmissionCommit generated in Submit.go
	reviewSignedStruct := pc.GetReviewSignedStruct(p.Id)
	reviewCommit := reviewSignedStruct.Commit //ReviewCommit generated in matchPapers
	nonce := reviewSignedStruct.Nonce
	rs := pc.GetPaperAndRandomness(p.Id).Rs //Rs generated in submit
	rr := pc.GetPaperAndRandomness(p.Id).Rr //Rr generated in submit

	PaperBigInt := MsgToBigInt(EncodeToBytes(p.Id))
	submitterPK := pc.GetPaperSubmitterPK(p.Id)
	proof := NewEqProofP256(PaperBigInt, rs, rr, nonce, &submitterPK, &pc.Keys.PublicKey)
	C1 := &Commitment{
		paperSubmissionCommit.X,
		paperSubmissionCommit.Y,
	}
	C2 := &Commitment{
		reviewCommit.X,
		reviewCommit.Y,
	}

	if !proof.OpenP256(C1, C2, nonce, &submitterPK, &pc.Keys.PublicKey) {
		works = false //for testing
		fmt.Println("Error: The review commit and paper submission commit does not hide the same paper")
	} else {
		works = true //for testing
		fmt.Println("The review commit and paper submission commit hides the same paper")
	}
	return works
}

func (pc *PC) GetKPCRFromLog(id int) string { //TODO: Maybe encrypt the KPCR when putting it on the log otherwise everyone can access it
	str := fmt.Sprintf("KPCR with PC and R%v", id)
	logmsg := model.Log{}
	model.GetLogMsgByMsg(&logmsg, str)
	EncodedKpcr := logmsg.Value
	DecodedKpcr := DecodeToStruct(EncodedKpcr).(string)
	return DecodedKpcr
}
