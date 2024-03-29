package backend

import (
	"crypto/ecdsa"
	_ "crypto/rand"
	"fmt"
	"log"
	ec "swag/ec"
	"swag/model"

	"github.com/mazen160/go-random"
)

//step 4
func (pc *PC) DistributePapers(reviewerSlice []Reviewer, paperSlice []*Paper) {
	for r := range reviewerSlice {
		Kpcr := pc.GetKPCRFromLog(reviewerSlice[r].UserID) //Shared key between R and PC (Kpcr) -
		for p := range paperSlice {
			SignedAndEncryptedPaper := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(paperSlice[p]), Kpcr)
			msg := fmt.Sprintf("SignedAndEncryptedPaper P%v for R%v", paperSlice[p].Id, reviewerSlice[r].UserID)
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
		if treeItem.value.([]byte) == nil {
			pList = append(pList, &Paper{
				Id:           0,
				Selected:     true,
				ReviewerList: []Reviewer{},
				Bytes:        []byte{},
				Title:        "ERROR: WAIT FOR PC ACTIONS AND REFRESH/F5",
			})
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
				UserID:              -1,
				Keys:                nil,
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

func (pc *PC) DeliverAssignedPaper() { //Unfortunately, reviewers get access to the entire paper reviewerlist this way
	for _, p := range pc.AllPapers {
		for _, r := range p.ReviewerList {
			str := fmt.Sprintf("DeliveredPaperForR%v", r.UserID)
			Kpcr := pc.GetKPCRFromLog(r.UserID)
			EncryptedPaper := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(p), Kpcr)
			logmsg := model.Log{
				State:      7,
				LogMsg:     str,
				FromUserID: 4000,
				Value:      EncryptedPaper[1],
				Signature:  EncryptedPaper[0],
			}
			model.CreateLogMsg(&logmsg)
		}
	}
}

func (pc *PC) AssignPaper(bidList []*PaperBid) {
	reviewersBidsTaken := []*Reviewer{}

	for _, p := range pc.AllPapers {
		if p.Selected {
			break
		}
		for _, bid := range bidList {
			if p.Id == bid.Paper.Id {
				if !p.Selected {
					if bid.Reviewer.PaperCommittedValue == nil {
						bid.Reviewer.PaperCommittedValue = &CommitStructPaper{}
					}
					p.ReviewerList = append(p.ReviewerList, *bid.Reviewer)
					p.Selected = true
				} else {
					reviewersBidsTaken = append(reviewersBidsTaken, bid.Reviewer)
				}
			}
		}
	}
	for _, r := range reviewersBidsTaken {
		x := false
		if r.PaperCommittedValue == nil {
			r.PaperCommittedValue = &CommitStructPaper{}
		}
		for _, p := range pc.AllPapers {
			if !p.Selected {
				x = true
				p.Selected = true
				p.ReviewerList = append(p.ReviewerList, *r)
				break
			}
		}
		if x {
			r.UserID = -1
			x = false
		}
	}
	for _, r := range reviewersBidsTaken {
		if r.UserID != -1 {
			r.PaperCommittedValue = &CommitStructPaper{}
			rand, err := random.IntRange(0, len(pc.AllPapers))
			if err != nil {
				log.Panicln("Panic in AssignPapers when generating random number")
			}
			pc.AllPapers[rand].ReviewerList = append(pc.AllPapers[rand].ReviewerList, *r)
			pc.AllPapers[rand].Selected = true
		}
	}
}

func (pc *PC) MatchPapers() {
	for _, p := range pc.AllPapers {
		fmt.Println("in match papers")
		PaperBigInt := MsgToBigInt(EncodeToBytes(p.Id)) //notice what we are creating our commitment from, maybe this ok.

		nonce_r := ec.GetRandomInt(pc.Keys.D)

		reviewerKeyList := []ecdsa.PublicKey{}
		for _, r := range p.ReviewerList {
			reviewerKeyList = append(reviewerKeyList, r.Keys.PublicKey)
		}
		rr := pc.GetPaperAndRandomness(p.Id).Rr
		commit := pc.GetPaperReviewCommitPC(PaperBigInt, rr) //paper review commit

		reviewStruct := ReviewSignedStruct{
			commit,
			reviewerKeyList,
			nonce_r,
		}

		signature := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(reviewStruct), "")

		msg := fmt.Sprintf("ReviewSignedStruct with P%v", p.Id)
		logmsg := model.Log{
			State:      7,
			LogMsg:     msg,
			FromUserID: 4000,
			Value:      signature[1],
			Signature:  signature[0],
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
	paperSubmissionCommit := pc.GetPaperSubmissionCommit(p.Id) //PaperSubmissionCommit generated in Submit.go
	rs := pc.GetPaperAndRandomness(p.Id).Rs                    //Rs generated in submit
	rr := pc.GetPaperAndRandomness(p.Id).Rr                    //Rr generated in submit
	reviewSignedStruct := pc.GetReviewSignedStruct(p.Id)
	reviewCommit := reviewSignedStruct.Commit //ReviewCommit generated by PC in matchPapers()
	nonce := reviewSignedStruct.Nonce         //Nonce from reviewSignedStruct

	PaperBigInt := MsgToBigInt(EncodeToBytes(p.Id)) //Converting the Paper to a big integer
	submitterPK := pc.GetPaperSubmitterPK(p.Id)     //Retriving the public keys of the submitter

	proof := NewEqProofP256(PaperBigInt, rs, rr, nonce, &submitterPK, &pc.Keys.PublicKey) //NIZK
	signedNIZK := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(proof), "")                //Signed NIZK

	logmsg := model.Log{
		State:      6,
		LogMsg:     fmt.Sprintf("NIZK Proving the review commit and paper submission commit hide the same paper for P%v", p.Id),
		FromUserID: 4000,
		Value:      signedNIZK[1],
		Signature:  signedNIZK[0],
	}
	model.CreateLogMsg(&logmsg) //NIZK published to log

	C1 := &Commitment{ //Converting commitments to structs
		paperSubmissionCommit.X,
		paperSubmissionCommit.Y,
	}
	C2 := &Commitment{
		reviewCommit.X,
		reviewCommit.Y,
	}

	return proof.OpenP256(C1, C2, nonce, &submitterPK, &pc.Keys.PublicKey)
}

func (pc *PC) GetKPCRFromLog(id int) string { 
	str := fmt.Sprintf("KPCR with PC and R%v", id)
	logmsg := model.Log{}
	model.GetLogMsgByMsg(&logmsg, str)
	EncodedKpcr := logmsg.Value
	DecodedKpcr := DecodeToStruct(EncodedKpcr).(string)
	return DecodedKpcr
}
