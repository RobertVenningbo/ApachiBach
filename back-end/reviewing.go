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

	hashedPaper, err := GetMessageHash(EncodeToBytes(r.paperCommittedValue)) //hashing the paper assigned to a reviewer
	if err != nil {
		log.Fatal(err)
	}
	rSignature, _ := ecdsa.SignASN1(rand.Reader, r.keys, hashedPaper)
	encrypted := Encrypt(rSignature, Kpcr)

	log.Printf("\n, Reviewer signs reviewed paper and encrypts signature of given paper %s", encrypted)
	tree.Put("Step 8: ", encrypted)
}

//planned to be called for every reviewer in the controller layer or whatever calls it
func (r *Reviewer) SignReviewPaperCommit() { //step 9
	hashedPaperCommit, err := GetMessageHash(EncodeToBytes(r.paperCommittedValue.CommittedValue.CommittedValue)) //hashing the paper assigned to a reviewer
	if err != nil {
		log.Fatal(err)
	}

	nonce := tree.Find("nonce") //find nonce (n_r)

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

	kpBytes := EncodeToBytes(kp)

	hashedKeys, err := GetMessageHash(kpBytes) 
	if err != nil {
		log.Fatal(err)
	}
	hashedKeysSignature, _ := ecdsa.SignASN1(rand.Reader, pc.keys, hashedKeys)
	
	rgBytes := EncodeToBytes(rg)

	hashedRg, err := GetMessageHash(rgBytes) 
	if err != nil {
		log.Fatal(err)
	}
	hashedRgSignature, _ := ecdsa.SignASN1(rand.Reader, pc.keys, hashedRg)

	strHashedKeys := fmt.Sprintf("PC signs Kp with signature %s", hashedKeysSignature) 
	log.Printf("\n"+strHashedKeys)
	tree.Put(strHashedKeys, hashedKeysSignature)

	
	strHashedRg := fmt.Sprintf("PC signs Rg with signature %s", hashedRgSignature) 
	log.Printf("\n"+strHashedRg, hashedRgSignature)
	tree.Put(strHashedRg, hashedRgSignature)

	for _, r := range *reviewers {
		Kpcr := generateSharedSecret(pc, nil, &r)

		encrypted := Encrypt(kpBytes, Kpcr)

		str := fmt.Sprintf("PC logs Kp encrypted with Kpcr between PC and reviewer id %s", r/*.userID*/) //gul streg pga r ikke er string (endnu)
		log.Printf("\n"+str)
		tree.Put(str, encrypted)

		str = fmt.Sprintf("PC logs Rg encrypted with Kpcr between PC and reviewer id %s", r/*.userID*/) //gul streg pga r ikke er string (endnu)
	
	}
}
