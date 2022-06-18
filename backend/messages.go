package backend

import (
	"crypto/ecdsa"
	"math/big"
)

type SubmitMessage struct {
	PaperAndRandomness []byte
	EncryptedKpcs      []byte
}

type CommitMsg struct {
	IdenityCommit []byte
	PaperCommit   []byte
}

type ValueSignature struct {
	Value     []byte
	Signature []byte
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

//Matching

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

type Grade struct {
	Grade      int
	Randomness int
}

type GradeReviewCommits struct {
	PaperReviewCommit *ecdsa.PublicKey
	GradeCommit       *ecdsa.PublicKey
	Nonce             *big.Int
}

//Decision

type SendGradeStruct struct {
	Reviews []string
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

type Paper2 struct {
	Id           int
	Selected     bool
	ReviewerList []Reviewer
	Bytes        []byte
}

type PaperBid struct {
	Paper    *Paper
	Reviewer *Reviewer
}

type SubmitStruct struct {
	Paper *Paper
	Rr    *big.Int
	Rs    *big.Int
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
