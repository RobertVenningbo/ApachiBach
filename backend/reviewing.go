package backend

import (
	_ "errors"
	"fmt"
	"log"
	ec "swag/ec"
	"swag/model"
)

func (r *Reviewer) GetAssignedPaperFromPCLog() *Paper {
	str := fmt.Sprintf("DeliveredPaperForR%v", r.UserID)
	Kpcr := GenerateSharedSecret(&Pc, nil, r)
	var logmsg model.Log
	model.GetLogMsgByMsg(&logmsg, str)
	encodedPaper := Decrypt(logmsg.Value, Kpcr)
	decodedPaper := DecodeToStruct(encodedPaper).(Paper)

	isLegit := VerifySignature(str, encodedPaper, &Pc.Keys.PublicKey)
	if !isLegit {
		fmt.Printf("\nReviewer %v couldn't verify signature from PC ", r.UserID)
	} else {
		fmt.Printf("\nReviewer %v verifies signature from PC - recieves assigned paper %v", r.UserID, decodedPaper.Id)
	}

	return &decodedPaper

}

func (r *Reviewer) FinishReview(review string) { //step 8
	Kpcr := GenerateSharedSecret(&Pc, nil, r)

	reviewStruct := ReviewStruct{
		r.UserID,
		review,
		r.PaperCommittedValue.Paper.Id,
	}

	signAndEnc := SignsPossiblyEncrypts(r.Keys, EncodeToBytes(reviewStruct), Kpcr)
	str := fmt.Sprintf("Reviewer, %v, finish review on paper", r.UserID)

	logmsg := model.Log{
		State:      8,
		LogMsg:     str,
		FromUserID: r.UserID,
		Value:      signAndEnc[1],
		Signature:  signAndEnc[0],
	}
	model.CreateLogMsg(&logmsg)
	Trae.Put(str, signAndEnc)

}

func (r *Reviewer) SignReviewPaperCommit() { //step 9
	reviewSignedStruct := r.GetReviewSignedStruct(r.PaperCommittedValue.Paper.Id)
	reviewCommit := reviewSignedStruct.Commit

	nonce := reviewSignedStruct.Nonce

	reviewCommitNonce := ReviewCommitNonceStruct{
		reviewCommit,
		nonce,
	}
	rCommitSignature := SignsPossiblyEncrypts(r.Keys, EncodeToBytes(reviewCommitNonce), "")

	str := fmt.Sprintf("Reviewer %v signs paper review commit \n", r.UserID)
	logmsg := model.Log{
		State:      9,
		LogMsg:     str,
		FromUserID: r.UserID,
		Value:      rCommitSignature[1],
		Signature:  rCommitSignature[0],
	}
	model.CreateLogMsg(&logmsg)
	Trae.Put(str, rCommitSignature[1])
}

func (pc *PC) GenerateKeysForDiscussing() { //step 10
	tempStruct := ReviewKpAndRg{}
	for _, p := range pc.AllPapers {
		kp := NewKeys() //generating new group key

		rg := ec.GetRandomInt(pc.Keys.D) //generating new grade randomness rg for later commits.
		strPC := ""
		for _, r := range p.ReviewerList {
			Kpcr := pc.GetKPCRFromLog(r.UserID)

			GroupKeyAndRg := ReviewKpAndRg{
				kp,
				rg,
			}

			reviewKpAndRg := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(GroupKeyAndRg), Kpcr)
			str := fmt.Sprintf("PC signed and encrypted ReviewKpAndRg for revId%v", r.UserID)

			logmsg := model.Log{
				State:      10,
				LogMsg:     str,
				FromUserID: 4000,
				Value:      reviewKpAndRg[1],
				Signature:  reviewKpAndRg[0],
			}

			model.CreateLogMsg(&logmsg)
			Trae.Put(str, reviewKpAndRg[1])

			tempStruct = GroupKeyAndRg
		}
		strPC = fmt.Sprintf("Encrypted KpAndRg for PC, for Paper%v", p.Id)
		reviewKpAndRg := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(tempStruct), pc.Keys.D.String())

		logmsg := model.Log{
			State:      10,
			LogMsg:     strPC,
			FromUserID: 4000,
			Value:      reviewKpAndRg[1],
			Signature:  reviewKpAndRg[0],
		}
		model.CreateLogMsg(&logmsg)
		Trae.Put(strPC, reviewKpAndRg[1])
	}

}

func (pc *PC) GetKpAndRgPC(pId int) ReviewKpAndRg {
	strPC := fmt.Sprintf("Encrypted KpAndRg for PC, for Paper%v", pId)

	reviewKpAndRg := Trae.Find(strPC)
	if reviewKpAndRg == nil {
		CheckStringAgainstDB(strPC)
		reviewKpAndRg = Trae.Find(strPC)
	}

	bytes := reviewKpAndRg.value.([]byte)
	encodedReviewKpAndRg := Decrypt(bytes, pc.Keys.D.String())
	decodedReviewKpAndRg := DecodeToStruct(encodedReviewKpAndRg).(ReviewKpAndRg)

	return decodedReviewKpAndRg
}

func (pc *PC) CollectReviews() { //step 11
	revKpAndRg := ReviewKpAndRg{}
	for _, p := range pc.AllPapers {
		ReviewStructList := []ReviewStruct{}
		revKpAndRg = pc.GetKpAndRgPC(p.Id)
		Kp := revKpAndRg.GroupKey
		for _, r := range p.ReviewerList {
			reviewStruct, err := pc.GetReviewStruct(r)
			if err != nil {
				log.Println(err)
			}
			fmt.Printf("%#v \n", reviewStruct)
			ReviewStructList = append(ReviewStructList, reviewStruct)
		}
		listSignature := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(ReviewStructList), Kp.D.String())
		putStr := fmt.Sprintf("Sharing reviews with Reviewers matched to paper: %v", p.Id)

		logmsg := model.Log{
			State:      11,
			LogMsg:     putStr,
			FromUserID: 4000,
			Value:      listSignature[1],
			Signature:  listSignature[0],
		}
		err := model.CreateLogMsg(&logmsg)
		if err != nil {
			log.Println("error in (pc).CollectReviews")
		}
		Trae.Put(putStr, listSignature)
	}
}

func (r *Reviewer) GetReviewKpAndRg() ReviewKpAndRg { //Perhaps add verification of Pc signature.
	str := fmt.Sprintf("PC signed and encrypted ReviewKpAndRg for revId%v", r.UserID)
	reviewKpAndRg := Trae.Find(str)
	if reviewKpAndRg == nil {
		CheckStringAgainstDB(str)
		reviewKpAndRg = Trae.Find(str)
	}
	bytes := reviewKpAndRg.value.([]byte)

	Kpcr := GenerateSharedSecret(&Pc, nil, r)
	encodedReviewKpAndRg := Decrypt(bytes, Kpcr)
	decodedReviewKpAndRg := DecodeToStruct(encodedReviewKpAndRg).(ReviewKpAndRg)

	isLegit := VerifySignature(str, encodedReviewKpAndRg, &Pc.Keys.PublicKey)

	if !isLegit {
		fmt.Printf("\n Reviewer %v couldn't verify signature in GetReviewKpAndRg", r.UserID)
	} else {
		fmt.Printf("\n Reviewer %v verifies signature in GetReviewKpAndRg", r.UserID)
	}

	return decodedReviewKpAndRg
}

func (pc *PC) GetReviewStruct(reviewer Reviewer) (ReviewStruct, error) {
	str := fmt.Sprintf("Reviewer, %v, finish review on paper", reviewer.UserID)

	CheckStringAgainstDBStruct(str)
	signedReviewStruct := Trae.Find(str)

	reviewStructBytes := signedReviewStruct.value.(ValueSignature)
	Kpcr := pc.GetKPCRFromLog(reviewer.UserID)
	encodedReviewStructValue := Decrypt(reviewStructBytes.Value, Kpcr)
	decodedReviewStruct := DecodeToStruct(encodedReviewStructValue).(ReviewStruct)
	hash, _ := GetMessageHash(encodedReviewStructValue)

	isLegit := Verify(&reviewer.Keys.PublicKey, reviewStructBytes.Signature, hash)
	if decodedReviewStruct.Review == "" || !isLegit {
		err := fmt.Errorf("error in GetReviewStruct, Review is empty or verification failed")
		return ReviewStruct{}, err
	} else {
		fmt.Printf("PC verifies reviewer: %v's review.\n", reviewer.UserID)
	}
	return decodedReviewStruct, nil
}

func (r *Reviewer) GetReviewCommitNonceStruct() ReviewCommitNonceStruct {

	str := fmt.Sprintf("Reviewer %v signs paper review commit \n", r.UserID)
	TreeItem := Trae.Find(str)
	if TreeItem == nil {
		CheckStringAgainstDB(str)
		TreeItem = Trae.Find(str)
	}
	bytes := TreeItem.value.([]byte)

	theStruct := DecodeToStruct(bytes).(ReviewCommitNonceStruct)

	return theStruct
}

func (r *Reviewer) GetCollectedReviews() []ReviewStruct {
	kpAndRg := r.GetReviewKpAndRg()
	getStr := fmt.Sprintf("Sharing reviews with Reviewers matched to paper: %v", r.PaperCommittedValue.Paper.Id)

	treeItem := Trae.Find(getStr)
	if treeItem == nil {
		CheckStringAgainstDB(getStr)
		treeItem = Trae.Find(getStr)
	}

	bytes := treeItem.value.([]byte)
	encodedReviewStructList := Decrypt(bytes, kpAndRg.GroupKey.D.String())
	decodedReviewStructList := DecodeToStruct(encodedReviewStructList).([]ReviewStruct)

	isLegit := VerifySignature(getStr, encodedReviewStructList, &Pc.Keys.PublicKey)
	if !isLegit {
		fmt.Printf("\n Reviewer %v couldn't verify signature when collecting reviews", r.UserID)
	} else {
		fmt.Printf("\n Reviewer %v verifies PC signature when collecting reviews", r.UserID)
	}

	return decodedReviewStructList
}
