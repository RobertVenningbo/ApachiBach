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
	commit []byte
	keys   *[]ecdsa.PublicKey //This needs to be an array of keys once we can take more than 1 reviewer pr paper
	nonce  *big.Int
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

//TODO: TEST
func (r *Reviewer) getBiddedPaper() PaperBid { //TODO test this function

	Kpcr := generateSharedSecret(&pc, nil, r)
	EncryptedSignedBid := tree.Find("EncryptedSignedBids " + r.UserID)
	bytes := EncryptedSignedBid.value.([][]byte)
	_, enc := SplitSignatureAndMsg(bytes)
	decrypted := Decrypt([]byte(enc), Kpcr)
	decoded := DecodeToStruct(decrypted)
	bid := decoded.(PaperBid)
	fmt.Printf("%s %v \n", "reviewer: ", bid.Reviewer)
	return bid
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

//TODO: TEST
func (pc *PC) assignPaper(reviewerSlice []Reviewer) bool {
	assignedPaper := false
	tmpList := []PaperBid{}
	for i := range reviewerSlice { //loop to get list of all bidded papers
		p := reviewerSlice[i].getBiddedPaper()
		tmpList = append(tmpList, p)
		fmt.Println(reviewerSlice[i].UserID)
	}
	for _, bid := range tmpList { //loop through all bidded papers
		reviewerList := bid.Paper.ReviewerList
		if bid.Paper.Selected { //if a paper is already selected
			for _, p := range pc.allPapers { //find a paper that isn't selected
				if p.Id == bid.Paper.Id {
					if !p.Selected {
						reviewerList = append(reviewerList, *bid.Reviewer) //Add reviewer to papers list of reviewers
						p.Selected = true
						bid.Reviewer.PaperCommittedValue.Paper = *bid.Paper //Maybe pointer issue
						fmt.Println("Paper: " + fmt.Sprintf("%v", p.Id) + " assigned")
						assignedPaper = true
					}
				}
			}
		} else { //if a bidded paper is NOT selected, assign it to first reviewer
			reviewerList = append(reviewerList, *bid.Reviewer) //Add reviewer to papers list of reviewers
			bid.Reviewer.PaperCommittedValue.Paper = *bid.Paper
			bid.Reviewer.PaperCommittedValue.Paper.Selected = true
			for _, p := range pc.allPapers {
				if p.Id == bid.Paper.Id { //find bidded paper in all papers and set it to selected
					p.Selected = true
					fmt.Println("Paper: " + fmt.Sprintf("%v", p.Id) + " assigned")
					assignedPaper = true
				}
			}
		}
	}
	for _, p := range pc.allPapers { //Loop through all papers
		for _, r := range reviewerSlice { //loop through reviewers and find a reviewer without assigned paper
			reviewerList := p.ReviewerList
			if &r.PaperCommittedValue.Paper == &(Paper{}) { //assign paper to reviewer
				r.PaperCommittedValue.Paper = p
				fmt.Println("Paper: " + fmt.Sprintf("%v", p.Id) + " assigned")
				reviewerList = append(reviewerList, r) //Add reviewer to papers list of reviewers
				assignedPaper = true
			}
		}
	}
	return assignedPaper
}

func (pc *PC) matchPapers(reviewers []Reviewer, submitters []Submitter, papers []Paper) {
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
		//fmt.Printf("%s %v \n", "ReviewCommit: ", ReviewCommit)
		nonce, _ := rand.Int(rand.Reader, curve.Params().N) //nonce_r
		reviewStruct := ReviewSignedStruct{ //Struct for signing commit, reviewer keys and nonce
			EncodeToBytes(pc.reviewCommits[0]),
			&reviewerKeyList,
			nonce,
		}
		PCsignedReviewCommitKeysNonce := Sign(pc.Keys, reviewStruct)
		tree.Put("PCsignedReviewCommitKeysNonce"+fmt.Sprintf("%v", p.Id), PCsignedReviewCommitKeysNonce)

		for _, s := range submitters {
			fmt.Printf("\n %s %v \n ", "paperid: ", s.PaperCommittedValue.Paper.Id) //for testing delete later
			if s.PaperCommittedValue.Paper.Id == p.Id {
				rs := s.PaperCommittedValue.R
				PaperSubmissionCommit := pc.GetPaperSubmissionCommit(&s)                //C(P, rs)
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

	// paper := r.paperCommittedValue.paper
	// PaperBigInt := MsgToBigInt(EncodeToBytes(paper))
	// commit, _ := r.GetCommitMessageReviewPaper(PaperBigInt, rr) //C(P, rr)

	// nonce := ec.GetRandomInt(r.keys.D) //nonce_r

	// reviewStruct := ReviewSignedStruct{ //Struct for signing commit, reviewer keys and nonce
	// 	EncodeToBytes(commit),
	// 	&r.keys.PublicKey,
	// 	nonce,
	// }
	// for _, r := range reviewers {

	// 	PCsignedReviewCommitKeysNonce := Sign(pc.keys, reviewStruct)

	// 	tree.Put("PCsignedReviewCommitKeysNonce"+r.userID, PCsignedReviewCommitKeysNonce)
	// 	log.Println("PCsignedReviewCommitKeysNonce" + r.userID + " logged.")
	// 	for _, s := range submitters {
	// 		paperCommitSubmitter := s.paperCommittedValue.CommittedValue
	// 		paperCommitReviewer := r.paperCommittedValue.CommittedValue
	// 		if paperCommitSubmitter == paperCommitReviewer {
	// 			fmt.Println(1)
	// 			//PaperSubmissionCommit := tree.Find("PCsignedPaperCommit" + fmt.Sprintf("%s",(s.paperCommittedValue.Id)))

	// 			//	commit1 := Commitment{

	// 		}

	// 	}
	// }
}
