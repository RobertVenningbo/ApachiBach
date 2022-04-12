package backend

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	ec "swag/ec"
)
func NewReceiver(key *ecdsa.PrivateKey) *Receiver {
	return &Receiver{
		keys: key,
	}
}

func SetCommitment(r *Receiver, comm *ecdsa.PublicKey) {
	r.commitment = comm
}

func GetTrapdoor(r *Receiver) *big.Int {
	return r.keys.D
}

// It returns values x and r (commitment was c = g^x * g^r).
func (s *Submitter) GetDecommitMsg() (*big.Int, *big.Int) {
	val := s.SubmitterCommittedValue.val
	r := s.SubmitterCommittedValue.r

	return val, r
}

func (s *Submitter) GetDecommitMsgPaper() (*big.Int, *big.Int) {
	val := s.PaperCommittedValue.val
	r := s.PaperCommittedValue.r

	return val, r
}

// When receiver receives a decommitment, CheckDecommitment verifies it against the stored value
// (stored by SetCommitment).
func (r *Receiver) CheckDecommitment(R, val *big.Int) bool {
	a := ec.ExpBaseG(r.keys, val)             // g^x
	b := ec.Exp(r.keys, &r.keys.PublicKey, R) // h^r
	c := ec.Mul(r.keys, a, b)                 // g^x * h^r

	return Equals(c, r.commitment)
}

func (s *Submitter) GetCommitMessage(val *big.Int, r *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(s.Keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r

	s.SubmitterCommittedValue.r = r     //hiding factor?
	s.SubmitterCommittedValue.val = val //den value (random) vi comitter ting til
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

	s.PaperCommittedValue.r = r

	s.PaperCommittedValue.val = val

	x1 := ec.ExpBaseG(s.Keys, val)
	x2 := ec.Exp(s.Keys, &s.Keys.PublicKey, r)
	comm := ec.Mul(s.Keys, x1, x2)
	s.PaperCommittedValue.CommittedValue = comm
	return comm, nil
}

func (pc *PC) GetCommitMessageReviewPaperTest(val *big.Int, r *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(pc.keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}
	//c = g^x * h^r
	comm := &ecdsa.PublicKey{}
	for i := range pc.reviewCommits {
		if pc.reviewCommits[i].r == nil {
			pc.reviewCommits[i].r = r
			pc.reviewCommits[i].val = val
			x1 := ec.ExpBaseG(pc.keys, val)
			x2 := ec.Exp(pc.keys, &pc.keys.PublicKey, r)
			comm = ec.Mul(pc.keys, x1, x2)
			pc.reviewCommits[i].CommittedValue = comm
		}
	}
	return comm, nil

} //C(P, r)  C(S, r)


func (rev *Reviewer) GetCommitMessageReviewPaper(val *big.Int, r *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(rev.keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r

	rev.paperCommittedValue.r = r

	rev.paperCommittedValue.val = val

	x1 := ec.ExpBaseG(rev.keys, val)
	x2 := ec.Exp(rev.keys, &rev.keys.PublicKey, r)
	comm := ec.Mul(rev.keys, x1, x2)
	rev.paperCommittedValue.CommittedValue = comm

	return comm, nil
} //C(P, r)  C(S, r)


func (rev *Reviewer) GetCommitMessageReviewGrade(val *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(rev.keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r
	r := ec.GetRandomInt(rev.keys.D)

	rev.gradeCommittedValue.r = r
	rev.gradeCommittedValue.val = val
	
	
	x1 := ec.ExpBaseG(rev.keys, val)
	x2 := ec.Exp(rev.keys, &rev.keys.PublicKey, r)
	comm := ec.Mul(rev.keys, x1, x2)
	rev.gradeCommittedValue.CommittedValue = comm

	return comm, nil
} //C(P, r)  C(S, r)


//verify
func (s *Submitter) VerifyTrapdoorSubmitter(trapdoor *big.Int) bool {
	h := ec.ExpBaseG(s.Keys, trapdoor)
	return Equals(h, &s.Keys.PublicKey)
}