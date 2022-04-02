package backend

import (
	"crypto/ecdsa"
	_ "crypto/elliptic"
	_ "errors"
	"fmt"
	_ "fmt"
	"log"
	"math/big"
)

type reviewCommitNonceStruct struct {
	commit 			ecdsa.PublicKey
	nonce			big.Int
}

//TODO: Look at how we store values in the tree, keys.
//planned to be called for every reviewer in the controller layer or whatever calls it
func (r *Reviewer) FinishReview() { //step 8
	Kpcr := generateSharedSecret(&pc, nil, r)

	signAndEnc := SignzAndEncrypt(r.keys, r.paperCommittedValue, Kpcr)
	str := fmt.Sprintf("Reviewer, %s, finish review on paper\n", r.userID)
	log.Printf(str)
	tree.Put(str, signAndEnc)
}

// Logic should be OK, but casting, encoding and decoding might fuck it up.
func (pc *PC) CollectReviews(reviewers []Reviewer) { // step 11

	log.Println("PC collects reviews from log")
	list := []string{}
	Kpcr := ""
	KpcrRevId := ""
	for _, r := range reviewers {
		collectString := fmt.Sprintf("Reviewer, %s, finish review on paper\n", r.userID)
		result := tree.Find(collectString)
		// all of this is verifying pretty much
		signature, encrypted := SplitSignz(fmt.Sprintf("%v", result.value))
		Kpcr = generateSharedSecret(pc, nil, &r)
		KpcrRevId = r.userID
		decrypted := Decrypt([]byte(encrypted), Kpcr)
		hash, _ := GetMessageHash(decrypted)
		isLegit := Verify(&r.keys.PublicKey, signature, hash)
		if !isLegit {
			log.Panic("Signature couldn't be verified.")
		}
		// all of this is verifying pretty much

		list = append(list, fmt.Sprint(result.value)) //watch out for fmt.Sprint formatting differently than wanting
	}
	log.Println("PC retrieves Kp")
	str := fmt.Sprintf("PC sign and encrypt Rg with Kpcr between PC and reviewer id %s", KpcrRevId)
	KpSigAndEnc := tree.Find(str)
	_, enc := SplitSignz(fmt.Sprintf("%v", KpSigAndEnc.value)) //could verify signature, but idk if it's needed for every received value. It's more "it's there if u wanna verify it".
	plaintext := Decrypt([]byte(enc), Kpcr)
	Kp := DecodeToStruct(plaintext)
	listSignature := SignzAndEncrypt(pc.keys, list, Kp.(ecdsa.PrivateKey).D.String())
	putStr := fmt.Sprint("Sharing reviews with Reviewers")
	log.Println(putStr)
	tree.Put(putStr, listSignature)
}

//planned to be called for every reviewer in the controller layer or whatever calls it
func (r *Reviewer) SignReviewPaperCommit() { //step 9
	PaperCommit := r.paperCommittedValue.CommittedValue.CommittedValue

	nonce := tree.Find("nonce") //find nonce in reviewSignStruct
	reviewCommitNonce := reviewCommitNonceStruct{
		*PaperCommit,
		nonce.value.(big.Int),
	}

	rCommitSignature := Sign(r.keys, reviewCommitNonce) //

	str := fmt.Sprintf("\n %s signs paper review commit  %s ", r.userID, rCommitSignature)
	log.Printf("%s=%s", str, rCommitSignature)
	tree.Put(str+" signed commit signature of review paper", rCommitSignature)

}

func (pc *PC) GenerateKeysForDiscussing(reviewers []Reviewer) {
	kp := newKeys() //generating new group key

	rg := GetRandomInt(pc.keys.D) //generating new grade randomness rg for later commits.

	for _, r := range reviewers {
		Kpcr := generateSharedSecret(pc, nil, &r)
		someSigKp := SignzAndEncrypt(pc.keys, kp, Kpcr) //return string([]byteSignature|someEncryptedString)

		str := fmt.Sprintf("PC sign and encrypt Kp with Kpcr between PC and reviewer id %s", r.userID)
		log.Printf("\n" + str + someSigKp)
		tree.Put(str, someSigKp)

		someSigRg := SignzAndEncrypt(pc.keys, rg, Kpcr) //return string([]byteSignature|someEncryptedString)
		str = fmt.Sprintf("PC sign and encrypt Rg with Kpcr between PC and reviewer id %s", r.userID)
		log.Printf("\n" + str + someSigRg)
		tree.Put(str, someSigRg)
	}
}
