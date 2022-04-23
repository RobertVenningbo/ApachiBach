package backend

import (
	"crypto/ecdsa"
	"fmt"
	ec "swag/ec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFinishReview_And_GetReviewStruct(t *testing.T) {
	reviewer.PaperCommittedValue.Paper = &p //TODO: intended to be removed in later stages as the struct reference shouldn't exist.

	reviewer.FinishReview("Very nice paper (y)")

	ReviewStruct, _ := pc.GetReviewStruct(reviewer)

	assert.Equal(t, "Very nice paper (y)", ReviewStruct.Review, "TestFinishReview_And_GetReviewStruct FAILED")
	assert.Equal(t, reviewer.UserID, ReviewStruct.ReviewerId, "TestFinishReview_And_GetReviewStruct FAILED")
}

func TestGenerateKeysForDiscussing_And_GetReviewKpAndRg(t *testing.T) {
	reviewerList := []Reviewer{}
	reviewerList = append(reviewerList, reviewer, reviewer2)
	reviewer.PaperCommittedValue.Paper = &paperListTest[0]
	reviewer2.PaperCommittedValue.Paper = &paperListTest[0]
	pc.GenerateKeysForDiscussing(reviewerList)
	pc.allPapers = append(pc.allPapers, &paperListTest[0])
	pc.allPapers[0].ReviewerList = reviewerList

	GetStruct := pc.GetKpAndRgPC(pc.allPapers[0].Id)

	fmt.Printf("%#v \n", GetStruct)

	//tested with print statements from within GenerateKeysForDiscussing function also
	//to compare structs.
	// WORKS (Y)
}

func TestSignReviewPaperCommit_And_GetReviewCommitNonceStruct(t *testing.T) {
	//reviewer.PaperCommittedValue.Paper = &p //TODO: intended to be removed in later stages as the struct reference shouldn't exist.

	nonce_r := ec.GetRandomInt(pc.Keys.D)

	reviewStruct := ReviewSignedStruct{
		&reviewer.Keys.PublicKey, //ignore that it's not a commit ;-)
		nil,
		nonce_r,
	}

	signature := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(reviewStruct), "")

	msg := fmt.Sprintf("ReviewSignedStruct with P%v", p.Id)
	tree.Put(msg, signature)

	reviewer.SignReviewPaperCommit()

	structz := reviewer.GetReviewCommitNonceStruct()
	assert.Equal(t, reviewStruct.Commit, structz.Commit, "failz")
	assert.Equal(t, reviewStruct.Nonce, structz.Nonce, "failz")
}

func TestCollectReviews_OrEntireReviewing(t *testing.T) {
	reviewer.PaperCommittedValue.Paper = &paperListTest[0]
	reviewer2.PaperCommittedValue.Paper = &paperListTest[0]
	reviewer3.PaperCommittedValue.Paper = &paperListTest[1]
	reviewer4.PaperCommittedValue.Paper = &paperListTest[1]
	paperListTest[0].ReviewerList = append(paperListTest[0].ReviewerList, reviewer, reviewer2)
	paperListTest[1].ReviewerList = append(paperListTest[1].ReviewerList, reviewer3, reviewer4)
	pc.allPapers = append(pc.allPapers, &paperListTest[0], &paperListTest[1])

	nonce_r := ec.GetRandomInt(pc.Keys.D)

	reviewStruct := ReviewSignedStruct{
		&reviewer.Keys.PublicKey, //ignore that it's not a commit ;-)
		[]ecdsa.PublicKey{reviewer.Keys.PublicKey, reviewer2.Keys.PublicKey},
		nonce_r,
	}

	signature := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(reviewStruct), "")

	msg := fmt.Sprintf("ReviewSignedStruct with P%v", p.Id)
	tree.Put(msg, signature)

	// ^A lot of setup, manually setting up what the previous steps would've done^.

	reviewer.FinishReview("Pretty rad paper!")             //step 8
	reviewer2.FinishReview("Pretty dope")                  //step 8
	reviewer3.FinishReview("Noice")                        //step 8
	reviewer4.FinishReview("I didn't enjoy this paper :(") //step 8

	reviewer.SignReviewPaperCommit()  //step 9
	reviewer2.SignReviewPaperCommit() //step 9
	// reviewer3.SignReviewPaperCommit() //step 9 -- Shouldn't be called currently as GetReviewSignedStruct from the calling of this method would fail as we don't currently have something in the tree with the following string "ReviewSignedStruct with p%v", p.Id) for p.Id=2.
	// reviewer4.SignReviewPaperCommit() //step 9

	// Generating keys for both paperReview groups.
	pc.GenerateKeysForDiscussing(paperListTest[0].ReviewerList) //step 10
	pc.GenerateKeysForDiscussing(paperListTest[1].ReviewerList) //step 10

	//Fabricating expected structs

	ReviewStructList := []ReviewStruct{
		{
			reviewer.UserID,
			"Pretty rad paper!",
			pc.allPapers[0].Id,
		},
		{
			reviewer2.UserID,
			"Pretty dope",
			pc.allPapers[0].Id,
		},
	}

	pc.CollectReviews(pc.allPapers[0].Id) //step 11

	ActualReviewStructList := reviewer.GetCollectedReviews()

	assert.Equal(t, ReviewStructList, ActualReviewStructList, "failz")
}
