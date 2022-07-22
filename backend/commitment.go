package backend

import (
	"crypto/ecdsa"
	"math/big"
	ec "swag/ec"
)

//////////////////////////////////////////////////////////////////////////////////
// 								modified code from                              //
// https://github.com/xlab-si/emmy/blob/master/crypto/ecpedersen/commitment.go  //
//////////////////////////////////////////////////////////////////////////////////

func NewReceiver(key *ecdsa.PrivateKey) *Receiver {
	return &Receiver{
		Keys: key,
	}
}

func SetCommitment(r *Receiver, comm *ecdsa.PublicKey) {
	r.Commitment = comm
}

func GetTrapdoor(r *Receiver) *big.Int {
	return r.Keys.D
}

func (s *Submitter) GetDecommitMsgPaper() (*big.Int, *big.Int) {
	val := s.PaperCommittedValue.Val
	r := s.PaperCommittedValue.R

	return val, r
}

// When receiver receives a decommitment, CheckDecommitment verifies it against the stored value
// (stored by SetCommitment).
func (r *Receiver) CheckDecommitment(R, val *big.Int) bool {
	a := ec.ExpBaseG(r.Keys, val)             // g^x
	b := ec.Exp(r.Keys, &r.Keys.PublicKey, R) // h^r
	c := ec.Mul(r.Keys, a, b)                 // g^x * h^r

	return Equals(c, r.Commitment)
}

func (s *Submitter) GetCommitMessage(val *big.Int, r *big.Int) (*ecdsa.PublicKey, error) {
	s.SubmitterCommittedValue.R = r     //hiding factor?
	s.SubmitterCommittedValue.Val = val //den value (random) vi comitter ting til
	x1 := ec.ExpBaseG(s.Keys, val)
	x2 := ec.Exp(s.Keys, &s.Keys.PublicKey, r)
	comm := ec.Mul(s.Keys, x1, x2)
	s.SubmitterCommittedValue.CommittedValue = comm

	return comm, nil
} //C(P, r)  C(S, r)

func (s *Submitter) GetPaperSubmissionCommit(val *big.Int, r *big.Int) (*ecdsa.PublicKey, error) {

	s.PaperCommittedValue.R = r

	s.PaperCommittedValue.Val = val

	x1 := ec.ExpBaseG(s.Keys, val)
	x2 := ec.Exp(s.Keys, &s.Keys.PublicKey, r)
	comm := ec.Mul(s.Keys, x1, x2)
	s.PaperCommittedValue.CommittedValue = comm
	return comm, nil
}

func (pc *PC) GetPaperReviewCommitPC(val *big.Int, r *big.Int) (*ecdsa.PublicKey) {

	x1 := ec.ExpBaseG(pc.Keys, val)
	x2 := ec.Exp(pc.Keys, &pc.Keys.PublicKey, r)
	return ec.Mul(pc.Keys, x1, x2)
	// c = g^x * h^r
	
} //C(P, r)  C(S, r)

func (rev *Reviewer) GetCommitMessageReviewGrade(val *big.Int, r *big.Int) (*ecdsa.PublicKey, error) {
	// c = g^x * h^r
	x1 := ec.ExpBaseG(rev.Keys, val)
	x2 := ec.Exp(rev.Keys, &rev.Keys.PublicKey, r)
	comm := ec.Mul(rev.Keys, x1, x2)

	return comm, nil
} //C(P, r)  C(S, r)

//verify
func (s *Submitter) VerifyTrapdoorSubmitter(trapdoor *big.Int) bool {
	h := ec.ExpBaseG(s.Keys, trapdoor)
	return Equals(h, &s.Keys.PublicKey)
}
