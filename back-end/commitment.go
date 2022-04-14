package backend

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	ec "swag/ec"
)
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

// It returns values x and r (commitment was c = g^x * g^r).
func (s *Submitter) GetDecommitMsg() (*big.Int, *big.Int) {
	val := s.SubmitterCommittedValue.Val
	r := s.SubmitterCommittedValue.R

	return val, r
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
	if val.Cmp(s.Keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r

	s.SubmitterCommittedValue.R = r     //hiding factor?
	s.SubmitterCommittedValue.Val = val //den value (random) vi comitter ting til
	x1 := ec.ExpBaseG(s.Keys, val)
	x2 := ec.Exp(s.Keys, &s.Keys.PublicKey, r)
	comm := ec.Mul(s.Keys, x1, x2)
	s.SubmitterCommittedValue.CommittedValue = comm

	return comm, nil
} //C(P, r)  C(S, r)

func (s *Submitter) GetCommitMessagePaper(val *big.Int, r *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(s.Keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r

	s.PaperCommittedValue.R = r

	s.PaperCommittedValue.Val = val

	x1 := ec.ExpBaseG(s.Keys, val)
	x2 := ec.Exp(s.Keys, &s.Keys.PublicKey, r)
	comm := ec.Mul(s.Keys, x1, x2)
	s.PaperCommittedValue.CommittedValue = comm
	return comm, nil
}

func (pc *PC) GetCommitMessageReviewPaperTest(val *big.Int, r *big.Int) error { //TODO test
	if val.Cmp(pc.Keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return err
	}
	//c = g^x * h^r
	//comm := &ecdsa.PublicKey{}
	
		x1 := ec.ExpBaseG(pc.Keys, val)
		x2 := ec.Exp(pc.Keys, &pc.Keys.PublicKey, r)
		comm := ec.Mul(pc.Keys, x1, x2)
		
		pc.reviewCommits = append(pc.reviewCommits, *comm)
	
		fmt.Printf("\n %s %v", "comm1: ", comm)
	return nil

} //C(P, r)  C(S, r)


func (rev *Reviewer) GetCommitMessageReviewPaper(val *big.Int, r *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(rev.Keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r

	rev.paperCommittedValue.R = r

	rev.paperCommittedValue.Val = val

	x1 := ec.ExpBaseG(rev.Keys, val)
	x2 := ec.Exp(rev.Keys, &rev.Keys.PublicKey, r)
	comm := ec.Mul(rev.Keys, x1, x2)
	rev.paperCommittedValue.CommittedValue = comm

	return comm, nil
} //C(P, r)  C(S, r)


func (rev *Reviewer) GetCommitMessageReviewGrade(val *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(rev.Keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r
	r := ec.GetRandomInt(rev.Keys.D)

	rev.gradeCommittedValue.R = r
	rev.gradeCommittedValue.Val = val
	
	
	x1 := ec.ExpBaseG(rev.Keys, val)
	x2 := ec.Exp(rev.Keys, &rev.Keys.PublicKey, r)
	comm := ec.Mul(rev.Keys, x1, x2)
	rev.gradeCommittedValue.CommittedValue = comm

	return comm, nil
} //C(P, r)  C(S, r)


//verify
func (s *Submitter) VerifyTrapdoorSubmitter(trapdoor *big.Int) bool {
	h := ec.ExpBaseG(s.Keys, trapdoor)
	return Equals(h, &s.Keys.PublicKey)
}