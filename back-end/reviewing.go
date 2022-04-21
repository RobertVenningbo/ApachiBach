package backend

import (
	"crypto/ecdsa"
	_ "crypto/elliptic"
	_ "errors"
	"fmt"
	_ "fmt"
	"log"
	"math/big"
	ec "swag/ec"
)

type reviewCommitNonceStruct struct {
	commit *ecdsa.PublicKey
	nonce  *big.Int
}

type ReviewStruct struct {
	ReviewerId   int
	Review  	 string
	PaperId		 int
}
type ReviewKpAndRg struct {
	GroupKey	*ecdsa.PrivateKey
	Rg 			*big.Int
}

func (r *Reviewer) FinishReview(review string) { //step 8
	Kpcr := generateSharedSecret(&pc, nil, r)

	reviewStruct := ReviewStruct{
		r.UserID,
		review,
		r.PaperCommittedValue.Paper.Id,
	}

	signAndEnc := SignsPossiblyEncrypts(r.Keys, EncodeToBytes(reviewStruct), Kpcr) 
	str := fmt.Sprintf("Reviewer, %v, finish review on paper\n", r.UserID)
	log.Printf(str)
	tree.Put(str, signAndEnc)
}

func (r *Reviewer) SignReviewPaperCommit() { //step 9 
	reviewSignedStruct := r.GetReviewSignedStruct(r.PaperCommittedValue.Paper.Id)
	reviewCommit := reviewSignedStruct.Commit

	nonce := reviewSignedStruct.Nonce

	reviewCommitNonce := reviewCommitNonceStruct{
		reviewCommit,
		nonce,
	}
	rCommitSignature := SignsPossiblyEncrypts(r.Keys, EncodeToBytes(reviewCommitNonce), "") 

	str := fmt.Sprintf("Reviewer %v signs paper review commit \n", r.UserID)
	log.Println(str)
	tree.Put(str, rCommitSignature)
}

func (pc *PC) GenerateKeysForDiscussing(reviewers []Reviewer) { //step 10
	kp := newKeys() //generating new group key

	rg := ec.GetRandomInt(pc.Keys.D) //generating new grade randomness rg for later commits.

	for _, r := range reviewers {
		Kpcr := generateSharedSecret(pc, nil, &r)
		GroupKeyAndRg := ReviewKpAndRg{
			kp,
			rg,
		}
		
		reviewKpAndRg := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(GroupKeyAndRg), Kpcr)

		str := fmt.Sprintf("PC signed and encrypted ReviewKpAndRg for revId%v", r.UserID)
		log.Printf("\n%s",str)
		tree.Put(str, reviewKpAndRg)
	}
}


func (pc *PC) CollectReviews(pId int) { //step 11
	ReviewStructList := []ReviewStruct{}
	revKpAndRg := ReviewKpAndRg{}
	for _, p := range pc.allPapers {
		if pId == p.Id {
			for _, r := range p.ReviewerList {
				reviewStruct, err := pc.GetReviewStruct(r)
				if err != nil {
					log.Panic(err)
				}
				ReviewStructList = append(ReviewStructList, reviewStruct) 
				revKpAndRg = pc.GetReviewKpAndRg(r)
				
			}
		}
	}
	log.Println("PC collects reviews from log")
	log.Println("PC retrieves Kp")

	Kp := revKpAndRg.GroupKey
	
	listSignature := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(ReviewStructList), Kp.D.String())
	putStr := fmt.Sprintf("Sharing reviews with Reviewers matched to paper: %v",pId)
	log.Println(putStr)
	tree.Put(putStr, listSignature)
}

func (pc *PC) GetReviewKpAndRg(reviewer Reviewer) ReviewKpAndRg {
	str := fmt.Sprintf("PC signed and encrypted ReviewKpAndRg for revId%v", reviewer.UserID)
	pc.GetReviewSignedStruct(reviewer.UserID)
	reviewKpAndRg := tree.Find(str).value
	_, encryptedReviewKpAndRg := SplitSignatureAndMsg(reviewKpAndRg.([][]byte))
	Kpcr := generateSharedSecret(pc, nil, &reviewer)
	encodedReviewKpAndRg := Decrypt(encryptedReviewKpAndRg, Kpcr)
	decodedReviewKpAndRg := DecodeToStruct(encodedReviewKpAndRg).(ReviewKpAndRg)

	return decodedReviewKpAndRg

}

func (pc *PC) GetReviewStruct(reviewer Reviewer) (ReviewStruct, error){
	str := fmt.Sprintf("Reviewer, %v, finish review on paper\n", reviewer.UserID)
	signedReviewStruct := (tree.Find(str)).value
	sig, encryptedReviewStruct := SplitSignatureAndMsg(signedReviewStruct.([][]byte))
	Kpcr := generateSharedSecret(pc, nil, &reviewer)
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

