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
	Trae.Put(str, rCommitSignature)
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

func (pc *PC) CollectReviews1(pId int) { //step 11
	ReviewStructList := []ReviewStruct{}
	revKpAndRg := ReviewKpAndRg{}
	for _, p := range pc.AllPapers {
		if pId == p.Id {
			for _, r := range p.ReviewerList {
				reviewStruct, err := pc.GetReviewStruct(r)
				if err != nil {
					log.Panic(err)
				}
				ReviewStructList = append(ReviewStructList, reviewStruct)
				revKpAndRg = pc.GetKpAndRgPC(pId)
			}
		}
	}

	Kp := revKpAndRg.GroupKey

	listSignature := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(ReviewStructList), Kp.D.String())
	putStr := fmt.Sprintf("Sharing reviews with Reviewers matched to paper: %v", pId)

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

func (pc *PC) CollectReviews() { //step 11
	fmt.Println(1)
	ReviewStructList := []ReviewStruct{}
	fmt.Println(2)
	revKpAndRg := ReviewKpAndRg{}
	fmt.Println(3)
	for _, p := range pc.AllPapers {
		fmt.Println(4)
		ReviewStructList = []ReviewStruct{}
		fmt.Println(5)
		revKpAndRg = pc.GetKpAndRgPC(p.Id)
		fmt.Println(6)
		Kp := revKpAndRg.GroupKey
		fmt.Println(7)
		for _, r := range p.ReviewerList {
			fmt.Println(8)
			reviewStruct, err := pc.GetReviewStruct(r)
			fmt.Println(9)
			if err != nil {
				log.Println(err)
			}
			fmt.Printf("%#v \n", reviewStruct)
			ReviewStructList = append(ReviewStructList, reviewStruct)
		}
		fmt.Println(10)
		listSignature := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(ReviewStructList), Kp.D.String())
		fmt.Println(11)
		putStr := fmt.Sprintf("Sharing reviews with Reviewers matched to paper: %v", p.Id)

		logmsg := model.Log{
			State:      11,
			LogMsg:     putStr,
			FromUserID: 4000,
			Value:      listSignature[1],
			Signature:  listSignature[0],
		}
		fmt.Println(12)
		err := model.CreateLogMsg(&logmsg)
		if err != nil {
			log.Println("error in (pc).CollectReviews")
		}
		Trae.Put(putStr, listSignature)
	}
}

func (r *Reviewer) GetReviewKpAndRg() ReviewKpAndRg {
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

	return decodedReviewKpAndRg
}

func (pc *PC) GetReviewStruct(reviewer Reviewer) (ReviewStruct, error) {
	str := fmt.Sprintf("Reviewer, %v, finish review on paper", reviewer.UserID)
	signedReviewStruct := Trae.Find(str)

	if signedReviewStruct == nil {
		CheckStringAgainstDBStruct(str)
		signedReviewStruct = Trae.Find(str)
	}

	reviewStructBytes := signedReviewStruct.value.([]byte)
	valueSignature := DecodeToStruct(reviewStructBytes).(ValueSignature)
	Kpcr := pc.GetKPCRFromLog(reviewer.UserID)
	encodedReviewStructValue := Decrypt(valueSignature.Value, Kpcr)
	decodedReviewStruct := DecodeToStruct(encodedReviewStructValue).(ReviewStruct)
	hash, _ := GetMessageHash(encodedReviewStructValue)

	isLegit := Verify(&reviewer.Keys.PublicKey, valueSignature.Signature, hash)
	if decodedReviewStruct.Review == "" || !isLegit {
		err := fmt.Errorf("error in GetReviewStruct, Review is empty or verification failed")
		return ReviewStruct{}, err
	}
	return decodedReviewStruct, nil
}

func (r *Reviewer) GetReviewCommitNonceStruct() ReviewCommitNonceStruct {

	str := fmt.Sprintf("Reviewer %v signs paper review commit \n", r.UserID)
	log.Printf("Reviewer: %v gets ReviewCommitNonce \n", r.UserID)
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

	return decodedReviewStructList
}
