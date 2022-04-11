package backend

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	ec "swag/ec"
)

type ReviewSignedStruct struct {
	commit []byte
	keys   *ecdsa.PublicKey //This needs to be an array of keys once we can take more than 1 reviewer pr paper
	nonce  *big.Int
}

//step 4
//TODO: TEST 
func (pc *PC) distributePapers(reviewerSlice []Reviewer, paperSlice []Paper) {
	for r := range reviewerSlice {
		Kpcr := generateSharedSecret(pc, nil, &reviewerSlice[r]) //Shared key between R and PC (Kpcr) -
		for p := range paperSlice {
			SignedAndEncryptedPaper := SignzAndEncrypt(pc.keys, paperSlice[p], Kpcr)
			tree.Put("SignedAndEncryptedPaper"+fmt.Sprintf("%v",(paperSlice[p].Id)), SignedAndEncryptedPaper)
			log.Println("SignedAndEncryptedPaper" + fmt.Sprintf("%v",(paperSlice[p].Id)))

		}
	}
}

// func getPaperList(pc *PC, reviewer *Reviewer) []Paper { -- måske kan vi bruge det her til at hente fra log 

// 	pMap := reviewer.paperMap
// 	Kpcr := generateSharedSecret(pc, nil, reviewer)
// 	pList := []Paper{}
// 	for _, v := range pMap {
// 		decrypted := Decrypt(v, Kpcr)
// 		p := DecodeToStruct(decrypted)
// 		pList = append(pList, p.(Paper))
// 	}
// 	return pList
// }

//TODO: TEST 
func (r *Reviewer) getBiddedPaper() PaperBid{ //TODO test this function
	Kpcr := generateSharedSecret(&pc, nil, r)
	EncryptedSignedBid := tree.Find("EncryptedSignedBids" + r.userID)
	str := EncryptedSignedBid.value.(string)
	_, enc :=SplitSignz(str)
	decrypted := Decrypt([]byte(enc), Kpcr)
	decoded := DecodeToStruct(decrypted)
	bid := decoded.(PaperBid)

	return bid
}

func (r *Reviewer) makeBid(pap *Paper) (PaperBid) {
	return PaperBid{
		pap,
		r,
	}
}

//step 5
func (r *Reviewer) SignBidAndEncrypt(p *Paper) { //set encrypted bid list
	bid := r.makeBid(p)
	Kpcr := generateSharedSecret(&pc, nil, r) //Shared secret key between R and PC
	EncryptedSignedBid := SignzAndEncrypt(r.keys, bid, Kpcr)
	tree.Put("EncryptedSignedBids"+r.userID, EncryptedSignedBid)
	log.Println("EncryptedSignedBids" + r.userID + "logged.")
}


//TODO: TEST 
func (pc *PC) assignPaper(reviewerSlice []Reviewer) {
	tmpList := []PaperBid{}
	for i := range reviewerSlice{	//loop to get list of all bidded papers
		p := reviewerSlice[i].getBiddedPaper()
		tmpList = append(tmpList, p)
	}
	for _, bid := range tmpList { //loop through all bidded papers
		reviewerList := bid.reviewer.paperCommittedValue.paper.ReviewerList
		if(bid.paper.Selected) { //if a paper is already selected
			for _, p := range pc.allPapers { //find a paper that isn't selected
				if p.Id == bid.paper.Id { 
					if !p.Selected {
						reviewerList = append(reviewerList, *bid.reviewer) //Add reviewer to papers list of reviewers
						p.Selected = true 
						bid.reviewer.paperCommittedValue.paper = *bid.paper //Maybe pointer issue
					}
				}
			}
		} else { //if a bidded paper is NOT selected, assign it to first reviewer
			reviewerList = append(reviewerList, *bid.reviewer) //Add reviewer to papers list of reviewers
			bid.reviewer.paperCommittedValue.paper = *bid.paper
			bid.reviewer.paperCommittedValue.paper.Selected = true
			for _, p := range pc.allPapers {
				if p.Id == bid.paper.Id { //find bidded paper in all papers and set it to selected
					p.Selected = true
				}
			}
		}
	} 
	for _, p := range pc.allPapers { //Loop through all papers
		for _, r := range reviewerSlice { //loop through reviewers and find a reviewer without assigned paper
			reviewerList := r.paperCommittedValue.paper.ReviewerList
			if &r.paperCommittedValue.paper == &(Paper{}) {  //assign paper to reviewer
				r.paperCommittedValue.paper = p
				reviewerList = append(reviewerList, r) //Add reviewer to papers list of reviewers
			}
		}
	}
}

func (pc *PC) matchPaper(reviewers []Reviewer, submitters []Submitter, p Paper) {
	rr := ec.GetRandomInt(pc.keys.D)

	//TODO: Loop through a Papers ReviewerList and find reviewer keys
	//TODO: who makes the commit?

	paper := r.paperCommittedValue.paper
	PaperBigInt := MsgToBigInt(EncodeToBytes(paper))
	commit, _ := r.GetCommitMessageReviewPaper(PaperBigInt, rr) //C(P, rr)

	nonce := ec.GetRandomInt(r.keys.D) //nonce_r

	reviewStruct := ReviewSignedStruct{ //Struct for signing commit, reviewer keys and nonce
		EncodeToBytes(commit),
		&r.keys.PublicKey,
		nonce,
	}
	for _, r := range reviewers {

		PCsignedReviewCommitKeysNonce := Sign(pc.keys, reviewStruct)

		tree.Put("PCsignedReviewCommitKeysNonce"+r.userID, PCsignedReviewCommitKeysNonce)
		log.Println("PCsignedReviewCommitKeysNonce" + r.userID + " logged.")
		for _, s := range submitters {
			paperCommitSubmitter := s.paperCommittedValue.CommittedValue
			paperCommitReviewer := r.paperCommittedValue.CommittedValue
			if paperCommitSubmitter == paperCommitReviewer {
				fmt.Println(1)
				//PaperSubmissionCommit := tree.Find("PCsignedPaperCommit" + fmt.Sprintf("%s",(s.paperCommittedValue.Id)))

				//	commit1 := Commitment{

			}

		}
	}
}
