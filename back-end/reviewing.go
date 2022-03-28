package backend

import (
	"crypto/ecdsa"
	_ "crypto/elliptic"
	"crypto/rand"
	_ "errors"
	"fmt"
	_ "fmt"
	"log"
)

//planned to be called for every reviewer in the controller layer or whatever calls it
func (r *Reviewer) FinishReview() { //step 8
	Kpcr := generateSharedSecret(&pc, nil, r)

	signAndEnc := SignzAndEncrypt(r.keys, r.paperCommittedValue, Kpcr)
	str := fmt.Sprintf("\n, Reviewer, %s, signs and encrypts paper: %s", r.userID, signAndEnc)
	log.Printf(str)
	tree.Put(str, signAndEnc)
}

//planned to be called for every reviewer in the controller layer or whatever calls it
func (r *Reviewer) SignReviewPaperCommit() { //step 9
	hashedPaperCommit, err := GetMessageHash(EncodeToBytes(r.paperCommittedValue.CommittedValue.CommittedValue)) //hashing the paper assigned to a reviewer
	if err != nil {
		log.Fatal(err)
	}

	nonce := tree.Find("nonce") //find nonce (n_r) - probably wanna decrypt also as getting from log i.e. should be encrypted value

	hashedNonce, _ := GetMessageHash(EncodeToBytes(nonce))

	rCommitSignature, _ := ecdsa.SignASN1(rand.Reader, r.keys, hashedPaperCommit)
	rNonceSignature, _ := ecdsa.SignASN1(rand.Reader, r.keys, hashedNonce)

	str := fmt.Sprintf("%#v", r) //hmm?
	log.Printf("\n %s signs paper review commit  %s ", str, rCommitSignature)
	tree.Put(str+" signed commit signature of review paper", rCommitSignature)

	log.Printf("\n %v signs paper review NONCE %s", r, rCommitSignature)
	tree.Put(str+" signed commit signature NONCE", rNonceSignature)

}

func (pc *PC) GenerateKeysForDiscussing(reviewers *[]Reviewer) {
	kp := newKeys() //generating new group key

	rg := GetRandomInt(pc.keys.D) //generating new grade randomness rg for later commits.

	for _, r := range *reviewers {
		Kpcr := generateSharedSecret(pc, nil, &r)
		someSigKp := SignzAndEncrypt(pc.keys, kp, Kpcr) //return string([]byteSignature|someEncryptedString)

		str := fmt.Sprintf("PC sign and encrypt Rg with Kpcr between PC and reviewer id %s", r.userID)
		log.Printf("\n" + str + someSigKp)
		tree.Put(str, someSigKp)

		someSigRg := SignzAndEncrypt(pc.keys, rg, Kpcr) //return string([]byteSignature|someEncryptedString)
		str = fmt.Sprintf("PC sign and encrypt Rg with Kpcr between PC and reviewer id %s", r.userID)
		log.Printf("\n" + str + someSigRg)
		tree.Put(str, someSigRg)
	}
}
