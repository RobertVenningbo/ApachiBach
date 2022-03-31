package backend

import (
	"crypto/ecdsa"
	"fmt"
	"log"
)

func (r *Reviewer) DetermineGrade(paperId int, grade int) map[int]int {
	pList := getPaperList(&pc, r) //This is the papers that the reviewer is grading right?
	for _, p := range pList {
		if p.Id == paperId {
			for paperId, v := range r.gradedPaperMap { //loop for mapping a grade to a paperID - unique per reviewer
				if v == 0 {
					r.gradedPaperMap[paperId] = grade //map[paperID][grade]
				}
			}
		}
	}
	return r.gradedPaperMap //return map with suggested grades for papers
}

func AgreeOnGrade(reviewers []Reviewer) {
	//loop through all reviewers gradedPaperMap
	//find average grade on a specific paper and round to nearest 4, 7, 10, 12
	//agree on grade on paper and sign paper review commit and review nonce
}

func (r *Reviewer) SendSecretMsgToReviewers(input string) { //intended to be for step 12, repeated.
	signNtext := Sign(r.keys, input)

	kpStr := fmt.Sprintf("PC sign and encrypt Kp with Kpcr between PC and reviewer id %s", r.userID)
	log.Printf("Getting cosigned Kp group key by reviewer: %s", r.userID)
	KpItem := tree.Find(kpStr)
	_, enc := SplitSignz(fmt.Sprint(KpItem))
	Kpcr := generateSharedSecret(&pc, nil, r)
	decryptedKp := Decrypt([]byte(enc), Kpcr)
	Kp := DecodeToStruct(decryptedKp)

	logStr := fmt.Sprintf("Sending msg to the log by reviewer: %s", r.userID)
	log.Println(logStr)
	encryptedMsg := Encrypt([]byte(signNtext), Kp.(ecdsa.PrivateKey).D.String())
	tree.Put(logStr, encryptedMsg)
}