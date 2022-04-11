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
	paper := s.paperCommittedValue.paper
	ri := s.submitterCommittedValue.val

	msg := ClaimMessage{
		&paper,
		s,
		ri,
	}
	str := fmt.Sprintf("Submitter, %s, claims paper by revealing paper and ri.", s.userID)
	signature := Sign(s.keys, msg)
	log.Println(str, signature)
	tree.Put(str, signature)
}

func (pc *PC) ConfirmOwnership(s *Submitter) { //step 20
	getStr := fmt.Sprintf("Submitter, %s, claims paper by revealing paper and ri.", s.userID)
	item := tree.Find(getStr)

	claimMsg := item.value.(ClaimMessage)

	/*perhaps verify some of the properties of claimMsg*/

	signature := Sign(pc.keys, claimMsg)

	putStr := fmt.Sprintf("PC confirms the ownership of paper, %v, to submitter: %s", claimMsg.paper.Id, s.userID)
	log.Println(putStr, " with signature: ", signature)
	tree.Put(putStr, signature)
}
