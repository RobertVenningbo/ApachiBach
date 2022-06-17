package backend

import (
	_ "errors"
	"fmt"
	_ "fmt"
	"log"
	ec "swag/ec"
)

func (r *Reviewer) FinishReview(review string) { //step 8
	Kpcr := GenerateSharedSecret(&Pc, nil, r)

	reviewStruct := ReviewStruct{
		r.UserID,
		review,
		r.PaperCommittedValue.Paper.Id,
	}

	signAndEnc := SignsPossiblyEncrypts(r.Keys, EncodeToBytes(reviewStruct), Kpcr)
	str := fmt.Sprintf("Reviewer, %v, finish review on paper\n", r.UserID)
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
	log.Println(str)
	Trae.Put(str, rCommitSignature)
}

func (pc *PC) GenerateKeysForDiscussing(reviewers []Reviewer) { //step 10
	kp := NewKeys() //generating new group key

	rg := ec.GetRandomInt(pc.Keys.D) //generating new grade randomness rg for later commits.
	strPC := ""
	tempStruct := ReviewKpAndRg{}
	for _, r := range reviewers {
		Kpcr := GenerateSharedSecret(pc, nil, &r)
		GroupKeyAndRg := ReviewKpAndRg{
			kp,
			rg,
		}

		reviewKpAndRg := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(GroupKeyAndRg), Kpcr)
		fmt.Printf("%#v \n", GroupKeyAndRg)

		str := fmt.Sprintf("PC signed and encrypted ReviewKpAndRg for revId%v", r.UserID)
		log.Printf("\n%s", str)
		Trae.Put(str, reviewKpAndRg)
		tempStruct = GroupKeyAndRg
		strPC = fmt.Sprintf("Encrypted KpAndRg for PC, for Paper%v", r.PaperCommittedValue.Paper.Id)
	}

	reviewKpAndRg := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(tempStruct), pc.Keys.D.String())

	log.Printf("\n%s", strPC)
	Trae.Put(strPC, reviewKpAndRg)
}

func (pc *PC) GetKpAndRgPC(pId int) ReviewKpAndRg {
	strPC := fmt.Sprintf("Encrypted KpAndRg for PC, for Paper%v", pId)

	reviewKpAndRg := Trae.Find(strPC).value
	_, encryptedReviewKpAndRg := SplitSignatureAndMsg(reviewKpAndRg.([][]byte))
	encodedReviewKpAndRg := Decrypt(encryptedReviewKpAndRg, pc.Keys.D.String())
	decodedReviewKpAndRg := DecodeToStruct(encodedReviewKpAndRg).(ReviewKpAndRg)

	return decodedReviewKpAndRg
}

func (pc *PC) CollectReviews(pId int) { //step 11
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
	log.Println("PC collects reviews from log")
	log.Println("PC retrieves Kp")

	Kp := revKpAndRg.GroupKey

	listSignature := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(ReviewStructList), Kp.D.String())
	putStr := fmt.Sprintf("Sharing reviews with Reviewers matched to paper: %v", pId)
	log.Println(putStr)
	Trae.Put(putStr, listSignature)
}

func (r *Reviewer) GetReviewKpAndRg() ReviewKpAndRg {
	str := fmt.Sprintf("PC signed and encrypted ReviewKpAndRg for revId%v", r.UserID)
	log.Printf("Getting cosigned Kp group key by reviewer: %v\n", r.UserID)
	reviewKpAndRg := Trae.Find(str).value
	_, encryptedReviewKpAndRg := SplitSignatureAndMsg(reviewKpAndRg.([][]byte))
	Kpcr := GenerateSharedSecret(&Pc, nil, r)
	encodedReviewKpAndRg := Decrypt(encryptedReviewKpAndRg, Kpcr)
	decodedReviewKpAndRg := DecodeToStruct(encodedReviewKpAndRg).(ReviewKpAndRg)

	return decodedReviewKpAndRg
}

func (pc *PC) GetReviewStruct(reviewer Reviewer) (ReviewStruct, error) {
	str := fmt.Sprintf("Reviewer, %v, finish review on paper\n", reviewer.UserID)
	signedReviewStruct := (Trae.Find(str)).value
	sig, encryptedReviewStruct := SplitSignatureAndMsg(signedReviewStruct.([][]byte))
	Kpcr := GenerateSharedSecret(pc, nil, &reviewer)
	encodedReviewStruct := Decrypt(encryptedReviewStruct, Kpcr)
	decodedReviewStruct := DecodeToStruct(encodedReviewStruct).(ReviewStruct)
	hash, _ := GetMessageHash(encodedReviewStruct)

	isLegit := Verify(&reviewer.Keys.PublicKey, sig, hash)
	if decodedReviewStruct.Review == "" || !isLegit {
		err := fmt.Errorf("Error in GetReviewStruct, Review is empty or verification failed")
		return ReviewStruct{}, err
	}
	return decodedReviewStruct, nil
}

func (r *Reviewer) GetReviewCommitNonceStruct() ReviewCommitNonceStruct {

	str := fmt.Sprintf("Reviewer %v signs paper review commit \n", r.UserID)
	log.Printf("Reviewer: %v gets ReviewCommitNonce \n", r.UserID)
	TreeItem := Trae.Find(str)
	_, encodedTheStruct := SplitSignatureAndMsg(TreeItem.value.([][]byte))

	theStruct := DecodeToStruct(encodedTheStruct).(ReviewCommitNonceStruct)

	return theStruct
}

func (r *Reviewer) GetCollectedReviews() []ReviewStruct {
	kpAndRg := r.GetReviewKpAndRg()
	getStr := fmt.Sprintf("Sharing reviews with Reviewers matched to paper: %v", r.PaperCommittedValue.Paper.Id)

	treeItem := Trae.Find(getStr).value
	_, encryptedReviewStructList := SplitSignatureAndMsg(treeItem.([][]byte))
	encodedReviewStructList := Decrypt(encryptedReviewStructList, kpAndRg.GroupKey.D.String())
	decodedReviewStructList := DecodeToStruct(encodedReviewStructList).([]ReviewStruct)

	return decodedReviewStructList
}
