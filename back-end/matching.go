package backend

import (
	"crypto/ecdsa"
	"log"
	"math/big"
	"strconv"
)

type ReviewSignedStruct struct {
	commit []byte
	keys   *ecdsa.PublicKey //This needs to be an array of keys once we can take more than 1 reviewer pr paper
	nonce  *big.Int
}

//step 4
func assignPapers(pc *PC, reviewerSlice []Reviewer, paperSlice []Paper) { 
	for r := range reviewerSlice {
		Kpcr := generateSharedSecret(pc, nil, &reviewerSlice[r]) //Shared key between R and PC (Kpcr) - 
		for p := range paperSlice {

			SignedAndEncryptedPaper :=SignzAndEncrypt(pc.keys, paperSlice[p], Kpcr)
			tree.Put("SignedAndEncryptedPaper" + strconv.Itoa(paperSlice[p].Id), SignedAndEncryptedPaper)
			log.Println("SignedAndEncryptedPaper" + strconv.Itoa(paperSlice[p].Id))

			encryptedPaper := Encrypt(EncodeToBytes(paperSlice[p]), Kpcr)
			reviewerSlice[r].paperMap[p] = encryptedPaper //TODO this should be the SignedAndEncryptedPapers right??
		}
	}
}

func getPaperList(pc *PC, reviewer *Reviewer) []Paper {

	pMap := reviewer.paperMap
	Kpcr := generateSharedSecret(pc, nil, reviewer)
	pList := []Paper{}
	for _, v := range pMap {
		decrypted := Decrypt(v, Kpcr)
		p := DecodeToStruct(decrypted)
		pList = append(pList, p.(Paper))
	}
	return pList
}

func makeBid(r *Reviewer, pap *Paper) {
	pList := getPaperList(&pc, r)

	for _, p := range pList {
		if p.Id == pap.Id {
			p.Selected = true
		}
	}
}

//step 5
func setEncBidList(r *Reviewer) { //set encrypted bid list 
	//TODO: Checkup if we are actually doing what we are supposed to here
	pList := getPaperList(&pc, r)
	Kpcr := generateSharedSecret(&pc, nil, r) //Shared secret key between R and PC
	tmpPaperList := []Paper{}
	for _, p := range pList {
		if p.Selected == true {
			tmpPaperList = append(tmpPaperList, p)
			putNextPaperInBidMapReviewer(r, Encrypt(EncodeToBytes(p), Kpcr))
		}
	}

	EncryptedSignedBids := Encrypt(EncodeToBytes(Sign(r.keys, r.biddedPaperMap)), Kpcr)
	tree.Put("EncryptedSignedBids" + r.userID, EncryptedSignedBids)
	log.Println("EncryptedSignedBids" + r.userID + "logged.")

	//r.biddedPaperMap = Encrypt(EncodeToBytes(tmpPaperList), Kpcr) // What is this line?? TODO
}

func matchPaper(reviewerSlice []Reviewer) { //step 6 (some of it)

	pList := getPaperList(&pc, &reviewerSlice[0])
	for _, rev := range reviewerSlice {
		kcpr := generateSharedSecret(&pc, nil, &rev) //Shared secret key between PC and R
		for i := range rev.biddedPaperMap {
			decrypted := Decrypt(rev.biddedPaperMap[i], kcpr)
			paper := DecodeToStruct(decrypted)
			if paper.(Paper).Selected {
				for i, p := range pList {
					if paper.(Paper).Id == p.Id {
						rev.paperCommittedValue = paper.(*Paper)     //assigning paper
						pList[i] = Paper{-1, nil, false, nil} //removing paper from generic map
						break
					}
				}
			}
		}
	}
	// Case for if reviewer weren't assigned paper because
	// other reviewer might have gotten the paper beforehand
	for _, rev := range reviewerSlice {
		if rev.paperCommittedValue == &(Paper{}) {
			for i, v := range pList {
				if pList[i].Id != -1 { // checking for removed papers
					//if this statement is true we have a normal paper
					rev.paperCommittedValue = &v
					pList[i] = Paper{-1, nil, false, nil} //removing paper from generic map
					break
				}
			}
		} else {
			break
		}
	}
}

func finalMatching(reviewers []Reviewer, submitters []Submitter) {
	for _, r := range reviewers {
		commit, _ := r.GetCommitMessageReviewPaper(GetRandomInt(r.keys.D))
		reviewStruct := ReviewSignedStruct{ //Struct for signing commit, reviewer keys and nonce
			EncodeToBytes(commit), 
			&r.keys.PublicKey,
			GetRandomInt(r.keys.D),
		}
		PCsignedReviewCommitKeysNonce := Sign(pc.keys, reviewStruct)
		tree.Put("PCsignedReviewCommitKeysNonce" + r.userID, PCsignedReviewCommitKeysNonce)
		log.Println("PCsignedReviewCommitKeysNonce" + r.userID + " logged.")
		for _, s := range submitters {
			paperCommitSubmitter := s.paperCommittedValue.CommittedValue.CommittedValue
			paperCommitReviewer := r.paperCommittedValue.CommittedValue.CommittedValue
			if paperCommitSubmitter == paperCommitReviewer {
				schnorrProofs = append(schnorrProofs, *CreateProof(s.keys, r.keys)) //NOT CORRECT, WAIT FOR ANSWER FROM SUPERVISOR
				//TODO El Gamal NIZK
			}
		}
	}
}