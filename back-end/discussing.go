package backend

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	ec "swag/ec"
)


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

/*
/// Probably also need function for determining when a grade should be settled
*/

func (r *Reviewer) CommitGrade() { //Step 13, assumed to be ran when reviewers have settled on a grade
	val := ec.GetRandomInt(r.keys.D)

	gradeCommit, _ := r.GetCommitMessageReviewGrade(val)

	fmt.Println(gradeCommit, "") //for compiler

	/*
		Step 6 aka finalMatching method from matching.go is dogwater and therefore i cba atm.
		For future dev: look at finalMatching() first and fix that shit before even trying  :)
	*/
}

func (r *Reviewer) SignAndEncryptGrade() {
	grade := "find real grade" //acquire agreed grade

	kpStr := fmt.Sprintf("PC sign and encrypt Kp with Kpcr between PC and reviewer id %s", r.userID)
	log.Printf("Getting cosigned Kp group key by reviewer: %s", r.userID)
	KpItem := tree.Find(kpStr)
	_, enc := SplitSignz(fmt.Sprint(KpItem))
	Kpcr := generateSharedSecret(&pc, nil, r)
	decryptedKp := Decrypt([]byte(enc), Kpcr)
	Kp := DecodeToStruct(decryptedKp)

	signAndEnc := SignzAndEncrypt(r.keys, grade, Kp.(ecdsa.PrivateKey).D.String()) //Notice Kp.(ecdsa.PrivateKey).D.String() seems super fishy, plz work.

	submitStr := fmt.Sprintf("Reviewer with id: %s, submits signed grade and enc kp", r.userID)
	log.Println(submitStr, ":", signAndEnc)
	tree.Put(submitStr, signAndEnc)
}
