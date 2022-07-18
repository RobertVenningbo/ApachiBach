package backend

import (
	"crypto/ecdsa"
	"math/big"
)

type SubmitMessage struct {
	PaperAndRandomness []byte
	EncryptedKpcs      []byte
}

type SubmitStruct struct {
	Paper *Paper
	Rr    *big.Int
	Rs    *big.Int
}

type CommitMsg struct {
	IdenityCommit []byte
	PaperCommit   []byte
}

type ValueSignature struct {
	Value     []byte
	Signature []byte
}

type Message struct {
	Title       string
	ReviewerIds []int
}

type ShareReviewsMessage struct {
	Reviews string
	Msgs    []Message
}

type CheckSubmissionsMessage struct {
	SubmittersLength int
	Submissions      int
	Submitters		 []string
}

//Review

type ReviewCommitNonceStruct struct {
	Commit *ecdsa.PublicKey
	Nonce  *big.Int
}

type ReviewStruct struct {
	ReviewerId int
	Review     string
	PaperId    int
}

type ReviewKpAndRg struct {
	GroupKey *ecdsa.PrivateKey
	Rg       *big.Int
}

type PaperBid struct {
	Paper    *Paper
	Reviewer *Reviewer
}
//Matching

type AllBids struct {
	PaperBidCount int
	Status        string
	ShowBool      bool
	UsersLength   int
}

type ReviewSignedStruct struct {
	Commit *ecdsa.PublicKey
	Keys   []ecdsa.PublicKey
	Nonce  *big.Int
}

//Discussion
type IndividualGrade struct {
	PaperId    int
	ReviewerId int
	Grade      int
}

type GradeReviewCommits struct {
	PaperReviewCommit *ecdsa.PublicKey
	GradeCommit       *ecdsa.PublicKey
	Nonce             *big.Int
}

type DiscussingViewData struct {
	Title   string
	Msgs    []string
	Reviews []ReviewStruct
}


//Decision

type SendGradeStruct struct {
	Reviews []ReviewStruct
	Grade   int
}


type RejectMessage struct {
	Commit *ecdsa.PublicKey
	Grade  int
	Rg     *big.Int
}

type RevealPaper struct {
	Paper Paper
	Rs    *big.Int
}

type RandomizeGradesForProofStruct struct {
	R           int64
	GradeBefore int64
	GradeAfter  int64
	PaperId		int
}

type GradeAndPaper struct {
	Grade int64
	Papir Paper
}

//Claim

type ClaimMessage struct {
	Paper     *Paper
	Submitter *Submitter
	Ri        *big.Int
}

//Agents

type Reviewer struct {
	UserID              int
	Keys                *ecdsa.PrivateKey
	PaperCommittedValue *CommitStructPaper
}

type Submitter struct {
	Keys                    *ecdsa.PrivateKey
	UserID                  int
	SubmitterCommittedValue *CommitStruct //commitstruct
	PaperCommittedValue     *CommitStructPaper
	Receiver                *Receiver
}

type CommitStruct struct {
	CommittedValue *ecdsa.PublicKey
	R              *big.Int
	Val            *big.Int
}

type CommitStructPaper struct {
	CommittedValue *ecdsa.PublicKey
	R              *big.Int
	Val            *big.Int
	Paper          *Paper
}

type PC struct {
	Keys      *ecdsa.PrivateKey
	AllPapers []*Paper //As long as this is only used for reference for withdrawel etc. then this is fine. We shouldn't mutate values within this.
}

type Paper struct {
	Id           int
	Selected     bool
	ReviewerList []Reviewer
	Bytes        []byte
	Title        string
}

type Receiver struct {
	Keys       *ecdsa.PrivateKey
	Commitment *ecdsa.PublicKey
}

// ZKP & Commitment

type EqProof struct {
	C  *big.Int
	D  *big.Int
	D1 *big.Int
	D2 *big.Int
}

type Commitment struct {
	X *big.Int
	Y *big.Int
}
