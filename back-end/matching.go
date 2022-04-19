package backend

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"
	ec "swag/ec"
)

type ReviewSignedStruct struct {
	Commit *ecdsa.PublicKey
	Keys   *[]ecdsa.PublicKey
	Nonce  *big.Int
}

//step 4
func (pc *PC) distributePapers(reviewerSlice []Reviewer, paperSlice []Paper) {
	//Find a way to retrieve a list of all Reviewers
	for r := range reviewerSlice {
		Kpcr := generateSharedSecret(pc, nil, &reviewerSlice[r]) //Shared key between R and PC (Kpcr) -
		for p := range paperSlice {
			SignedAndEncryptedPaper := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(paperSlice[p]), Kpcr)
			msg := fmt.Sprintf("SignedAndEncryptedPaper P%v for R%v", paperSlice[p].Id, reviewerSlice[r].UserID)
			tree.Put(msg, SignedAndEncryptedPaper)
			log.Println(msg)
		}
	}
}

//paperSlice is only there for getting len(paperSlice) for forloop.
//Gets all papers for each reviewer from log.
//Expected to be called for every reviewer when reviewers want to see list of all papers on frontend.
func (r *Reviewer) GetPapersReviewer(paperSlice []Paper) []Paper {
	Kpcr := generateSharedSecret(&pc, nil, r)

	pList := []Paper{}
	for i := 0; i < len(paperSlice); i++ {
		GetMsg := fmt.Sprintf("SignedAndEncryptedPaper P%v for R%v", paperSlice[i].Id, r.UserID)
		EncryptedSignedBid := tree.Find(GetMsg)
		bytes := EncryptedSignedBid.value.([][]byte)
		sig, enc := SplitSignatureAndMsg(bytes)
		decrypted := Decrypt(enc, Kpcr)
		hash, err := GetMessageHash(decrypted)
		if err != nil {
			log.Fatal(err)
		}
		isVerified := Verify(&pc.Keys.PublicKey, sig, hash) //casually verifying, cuz we can :)
		if !isVerified {
			log.Fatalf("Couldn't verify signature of paper: %v", (paperSlice[i].Id))
		}
		decoded := DecodeToStruct(decrypted)
		paper := decoded.(Paper)
		pList = append(pList, paper)
	}
	return pList
}

func (r *Reviewer) getBiddedPaper() *PaperBid {

	Kpcr := generateSharedSecret(&pc, nil, r)
	EncryptedSignedBid := tree.Find("EncryptedSignedBids " + r.UserID)
	bytes := EncryptedSignedBid.value.([][]byte)
	_, enc := SplitSignatureAndMsg(bytes)
	decrypted := Decrypt([]byte(enc), Kpcr)
	decoded := DecodeToStruct(decrypted)
	bid := decoded.(PaperBid)
	fmt.Printf("%s %v \n", "reviewer: ", bid.Reviewer)
	return &bid
}

func (r *Reviewer) makeBid(pap *Paper) *PaperBid {
	return &PaperBid{
		pap,
		r,
	}
}

//step 5
func (r *Reviewer) SignBidAndEncrypt(p *Paper) { //set encrypted bid list
	bid := r.makeBid(p)
	Kpcr := generateSharedSecret(&pc, nil, r) //Shared secret key between R and PC
	EncryptedSignedBid := SignsPossiblyEncrypts(r.Keys, EncodeToBytes(bid), Kpcr)
	tree.Put("EncryptedSignedBids "+r.UserID, EncryptedSignedBid)
	log.Println("EncryptedSignedBids" + r.UserID + "logged.")
}

func (pc *PC) replaceWithBids(reviewerSlice []*Reviewer) ([]*Paper, []*PaperBid) {
	bidList := []*PaperBid{}
	for i := range reviewerSlice { //loop to get list of all bidded papers
		p := reviewerSlice[i].getBiddedPaper()
		bidList = append(bidList, p)
	}

	for _, p := range pc.allPapers {
		for _, b := range bidList {
			if p.Id == b.Paper.Id {
				p = b.Paper

			}
		}
	}
	return pc.allPapers, bidList
}

func (pc *PC) assignPaper(reviewerSlice []*Reviewer) {
	reviewersBidsTaken := []Reviewer{}
	bidList := []*PaperBid{}
	for i := range reviewerSlice { //loop to get list of all bidded papers
		p := reviewerSlice[i].getBiddedPaper()
		bidList = append(bidList, p)
	}
	for _, bid := range bidList {
		for _, p := range pc.allPapers {
			if p.Id == bid.Paper.Id {
				if !p.Selected {
					if bid.Reviewer.PaperCommittedValue == nil {
						bid.Reviewer.PaperCommittedValue = &CommitStructPaper{}
					}
					p.Selected = true
					p.ReviewerList = append(p.ReviewerList, *bid.Reviewer)
					break
				} else {
					reviewersBidsTaken = append(reviewersBidsTaken, *bid.Reviewer)
					break
				}
			}
		}
	}
	for i, r := range reviewersBidsTaken {
		x := false
		if r.PaperCommittedValue == nil {
			r.PaperCommittedValue = &CommitStructPaper{}
		}
		for _, p := range pc.allPapers {
			if !p.Selected {
				x = true
				p.Selected = true
				reviewer := &r
				p.ReviewerList = append(p.ReviewerList, *reviewer)
				break
			}
		}
		if x {
			reviewersBidsTaken[i].UserID = "deleted"
			x = false
		}
	}
	for _, r := range reviewersBidsTaken {
		if ((r.PaperCommittedValue == nil) || (r.PaperCommittedValue == &CommitStructPaper{})) && (r.UserID != "deleted") {
			r.PaperCommittedValue = &CommitStructPaper{}
			for _, p := range pc.allPapers {
				p.Selected = true
				p.ReviewerList = append(p.ReviewerList, r)
				break
			}
		}
	}
	pc.SetReviewersPaper(reviewerSlice)
}

// This method is a little messy however it is not expected to be called on a lot of entities.
// **Finds every assigned reviewer for every paper and makes it bidirectional, such that a reviewer also has a reference to a paper**
// **Basically a fast reversal of assignPaper in terms of being bidirectional**
func (pc *PC) SetReviewersPaper(reviewerList []*Reviewer) {
	for _, p := range pc.allPapers {
		for _, r := range p.ReviewerList {
			for _, r1 := range reviewerList {
				if r.UserID == r1.UserID {
					if r1.PaperCommittedValue == nil {
						r1.PaperCommittedValue = &CommitStructPaper{}
					}
					r1.PaperCommittedValue.Paper = p
				}
			}
		}
	}
}
func (pc *PC) matchPaperz() {
	for _, p := range pc.allPapers {
		PaperBigInt := MsgToBigInt(EncodeToBytes(p))

		//TODO: rr should be retrieved from log (DELETE WHEN DONE)
		nonce_r := ec.GetRandomInt(pc.Keys.D)

		reviewerKeyList := []ecdsa.PublicKey{}
		for _, r := range p.ReviewerList {
			reviewerKeyList = append(reviewerKeyList, r.Keys.PublicKey)
		}

		rr := pc.GetPaperAndRandomness(p.Id).Rr

		commit, err := pc.GetCommitMessagePaperPC(PaperBigInt, rr)
		if err != nil {
			log.Panic("matchPaperz error")
		}

		reviewStruct := ReviewSignedStruct{
			commit,
			&reviewerKeyList,
			nonce_r,
		}

		signature := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(reviewStruct), "")

		msg := fmt.Sprintf("ReviewSignedStruct with P%v", p.Id)
		tree.Put(msg, signature)
	}
}

func (pc *PC) GetReviewSignedStruct(id int) ReviewSignedStruct {
	ret := ReviewSignedStruct{}
	for _, p := range pc.allPapers {
		if p.Id == id {
			msg := fmt.Sprintf("ReviewSignedStruct with P%v", p.Id)
			item := tree.Find(msg)
			_, encodedStruct := SplitSignatureAndMsg(item.value.([][]byte))
			decodedStruct := DecodeToStruct(encodedStruct)
			ret = decodedStruct.(ReviewSignedStruct)
			fmt.Printf("%s %v \n", "Review Commit: ", ret.Commit)
		}
	}
	return ret
}

func (pc *PC) supplyNIZK(p *Paper) bool {
	works := false                                             //for testing
	paperSubmissionCommit := pc.GetPaperSubmissionCommit(p.Id) //PaperSubmissionCommit generated in Submit.go
	reviewSignedStruct := pc.GetReviewSignedStruct(p.Id)
	reviewCommit := reviewSignedStruct.Commit //ReviewCommit generated in matchPapers

	nonce := reviewSignedStruct.Nonce
	rs := pc.GetPaperAndRandomness(p.Id).Rs //Rs generated in submit
	rr := pc.GetPaperAndRandomness(p.Id).Rr //Rr generated in submit

	PaperBigInt := MsgToBigInt(EncodeToBytes(p))

	submitterPK := pc.GetPaperSubmitterPK(p.Id)

	proof := NewEqProofP256(PaperBigInt, rr, rs, nonce, &submitterPK, &pc.Keys.PublicKey)

	C1 := Commitment{
		paperSubmissionCommit.X,
		paperSubmissionCommit.Y,
	}
	C2 := Commitment{
		reviewCommit.X,
		reviewCommit.Y,
	}

	if (!proof.OpenP256(&C1, &C2, nonce, &submitterPK, &pc.Keys.PublicKey)) {
		works = false //for testing
		fmt.Println("Error: The review commit and paper submission commit does not hide the same paper")
	} else {
		works = true //for testing
		fmt.Println("The review commit and paper submission commit hides the same paper")
	}
	return works
}

func (pc *PC) matchPapers(reviewers []Reviewer, submitters []Submitter, papers []*Paper) {
	for _, p := range papers {
		fmt.Println("Paper: " + fmt.Sprintf("%v", p.Id) + " looping")
		rr := ec.GetRandomInt(pc.Keys.D)
		PaperBigInt := MsgToBigInt(EncodeToBytes(p))
		reviewerList := p.ReviewerList
		reviewerKeyList := []ecdsa.PublicKey{}
		for _, r := range reviewerList {
			reviewerKeyList = append(reviewerKeyList, r.Keys.PublicKey)
		}
		pc.GetCommitMessageReviewPaperTest(PaperBigInt, rr) //C(P, rr)
		nonce, _ := rand.Int(rand.Reader, curve.Params().N) //nonce_r
		reviewStruct := ReviewSignedStruct{                 //Struct for signing commit, reviewer keys and nonce
			nil,
			&reviewerKeyList,
			nonce,
		}
		PCsignedReviewCommitKeysNonce := Sign(pc.Keys, reviewStruct)
		tree.Put("PCsignedReviewCommitKeysNonce"+fmt.Sprintf("%v", p.Id), PCsignedReviewCommitKeysNonce)
		for _, s := range submitters {
			fmt.Printf("\n %s %v \n ", "paperid: ", s.PaperCommittedValue.Paper.Id) //for testing delete later
			if s.PaperCommittedValue.Paper.Id == p.Id {
				rs := s.PaperCommittedValue.R
				PaperSubmissionCommit := pc.GetPaperSubmissionCommit(1)                  //C(P, rs)
				fmt.Printf("\n %s %v", "PaperSubmissionCommit: ", PaperSubmissionCommit) //for testing delete later
				proof := *NewEqProofP256(PaperBigInt, rr, rs, nonce, &s.Keys.PublicKey, &pc.Keys.PublicKey)
				C1 := Commitment{ //this is wrong, but trying for testing reasons, might need a for loop looping through reviewcommits
					pc.reviewCommits[0].X,
					pc.reviewCommits[0].Y,
				}
				fmt.Printf("\n %s %v ", "ReviewCommit: ", pc.reviewCommits[0])
				C2 := Commitment{
					PaperSubmissionCommit.X,
					PaperSubmissionCommit.Y,
				}
				if !proof.OpenP256(&C1, &C2, nonce, &s.Keys.PublicKey, &pc.Keys.PublicKey) {
					fmt.Println("Error: The review commit and paper submission commit does not hide the same paper")
				} else {
					fmt.Println("The review commit and paper submission commit hides the same paper")
				}
			}
		}
	}
}
