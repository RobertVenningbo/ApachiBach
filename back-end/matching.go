package backend

import (
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"log"
	"strconv"
)

//step 4
func assignPapers(pc *PC, reviewerSlice []Reviewer, paperSlice []Paper) { 
	for r := range reviewerSlice {
		Kpcr := generateSharedSecret(pc, nil, &reviewerSlice[r]) //Shared key between R and PC (Kpcr) - 
		for p := range paperSlice {

			hashedPaper, _ := GetMessageHash(EncodeToBytes(paperSlice[p])) 
			SignedPaperPC, _ := ecdsa.SignASN1(rand.Reader, pc.keys, hashedPaper)
			PaperAsString := fmt.Sprintf("%#v", paperSlice[p])
			tree.Put(PaperAsString + strconv.Itoa(paperSlice[p].Id), SignedPaperPC) 
			log.Println(PaperAsString + strconv.Itoa(paperSlice[p].Id) + "SignedPaperPC logged - The PC signed a paper")
			//putNextSignatureInMapPC(pc, pcSignature)

			encryptedPaper := Encrypt(EncodeToBytes(paperSlice[p]), Kpcr)
			encryptedAsString :=fmt.Sprintf("%#v", encryptedPaper)
			tree.Put(encryptedAsString + strconv.Itoa(paperSlice[p].Id), encryptedPaper) //Encrypted paper logged in tree
			log.Printf("\n %s %s", encryptedPaper, " encrypted paper logged")
			
			reviewerSlice[r].paperMap[p] = encryptedPaper
		}
	}
}

func getPaperList(pc *PC, reviewer *Reviewer) []Paper {

	pMap := reviewer.paperMap
	Kpcr := generateSharedSecret(pc, nil, reviewer)
	pList := []Paper{}
	for _, v := range pMap {
		decrypted := Decrypt(v, Kpcr)
		p := DecodeToPaper(decrypted)
		pList = append(pList, p)
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
	//TODO This needs to be channged when we figure out how to sign and encrypt correctly, currently we are signing and encrypting in the wrong order
	pList := getPaperList(&pc, r)
	Kpcr := generateSharedSecret(&pc, nil, r) //Shared secret key between R and PC
	tmpPaperList := []Paper{}
	for _, p := range pList {
		if p.Selected == true {
			tmpPaperList = append(tmpPaperList, p)
			putNextPaperInBidMapReviewer(r, Encrypt(EncodeToBytes(p), Kpcr))
		}
	}

	hashedBiddedPaperList, _ := GetMessageHash(EncodeToBytes(r.biddedPaperMap)) //changed from tmpPaperList to r.biddedPaperMap
	rSignature, _ := ecdsa.SignASN1(rand.Reader, r.keys, hashedBiddedPaperList)
	tree.Put(r.userID + "SignedBidByReviewer", rSignature)
	log.Printf("\n %s %s", rSignature, "Signed bid from reviewer: " + r.userID + " logged.")
	//putNextSignatureInMapReviewer(r, rSignature)

	//r.biddedPaperMap = Encrypt(EncodeToBytes(tmpPaperList), Kpcr)
}

func matchPaper(reviewerSlice []Reviewer) { //step 6 (some of it)

	pList := getPaperList(&pc, &reviewerSlice[0])
	for _, rev := range reviewerSlice {
		kcpr := generateSharedSecret(&pc, nil, &rev) //Shared secret key between PC and R
		for i := range rev.biddedPaperMap {
			decrypted := Decrypt(rev.biddedPaperMap[i], kcpr)
			paper := DecodeToPaper(decrypted)
			if paper.Selected {
				for i, p := range pList {
					if paper.Id == p.Id {
						rev.paperCommittedValue = &paper      //assigning paper
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
		r.GetCommitMessageReviewPaper(GetRandomInt(r.keys.D))
		nonce := GetRandomInt(r.keys.D)
		reviewerAsString := fmt.Sprintf("%#v", r)
		tree.Put(reviewerAsString + "nonce", nonce) //Nonce logged
		log.Printf("\n, %s, %s", "Nonce from reviewer: " + r.userID + " - ", nonce)
		hash, _ := GetMessageHash(EncodeToBytes(r.keys.PublicKey))
		signature, _ := ecdsa.SignASN1(rand.Reader, pc.keys, hash)
		//TODO LOG SIGNATURE AND MESSAGE. CONCATENATED?
		putNextSignatureInMapPC(&pc, signature)
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