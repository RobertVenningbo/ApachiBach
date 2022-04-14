package backend

import (
	"fmt"
	"log"
	"math/big"
)

type ClaimMessage struct {
	paper     *Paper
	submitter *Submitter
	ri        *big.Int
}

func (s *Submitter) ClaimPaper() { //step 19
	paper := s.PaperCommittedValue.Paper
	ri := s.SubmitterCommittedValue.Val

	msg := ClaimMessage{
		&paper,
		s,
		ri,
	}
	str := fmt.Sprintf("Submitter, %s, claims paper by revealing paper and ri.", s.UserID)
	signature := Sign(s.Keys, msg)
	log.Println(str, signature)
	tree.Put(str, signature)
}

func (pc *PC) ConfirmOwnership(s *Submitter) { //step 20
	getStr := fmt.Sprintf("Submitter, %s, claims paper by revealing paper and ri.", s.UserID)
	item := tree.Find(getStr)

	claimMsg := item.value.(ClaimMessage)

	/*perhaps verify some of the properties of claimMsg*/

	signature := Sign(pc.Keys, claimMsg)

	putStr := fmt.Sprintf("PC confirms the ownership of paper, %v, to submitter: %s", claimMsg.paper.Id, s.UserID)
	log.Println(putStr, " with signature: ", signature)
	tree.Put(putStr, signature)
}
