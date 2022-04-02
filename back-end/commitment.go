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
	val := s.submitterCommittedValue.val
	r := s.submitterCommittedValue.r

	return val, r
}

func (s *Submitter) GetDecommitMsgPaper() (*big.Int, *big.Int) {
	val := s.paperCommittedValue.CommittedValue.val
	r := s.paperCommittedValue.CommittedValue.r

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

func (s *Submitter) GetCommitMessage(val *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(s.keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r
	r := GetRandomInt(s.keys.D)

	s.submitterCommittedValue.r = r     //hiding factor?
	s.submitterCommittedValue.val = val //den value (random) vi comitter ting til
	x1 := ec.ExpBaseG(s.keys, val)
	x2 := ec.Exp(s.keys, &s.keys.PublicKey, r)
	comm := ec.Mul(s.keys, x1, x2)
	s.submitterCommittedValue.CommittedValue = comm

	return comm, nil
} //C(P, r)  C(S, r)

func (rev *Reviewer) GetCommitMessageReviewPaper(val *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(rev.keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r
	r := GetRandomInt(rev.keys.D)

	rev.paperCommittedValue.CommittedValue.r = r

	rev.paperCommittedValue.CommittedValue.val = val

	x1 := ec.ExpBaseG(rev.keys, val)
	x2 := ec.Exp(rev.keys, &rev.keys.PublicKey, r)
	comm := ec.Mul(rev.keys, x1, x2)
	rev.paperCommittedValue.CommittedValue.CommittedValue = comm
	fmt.Printf("\n %s, %s, %s", "R & Val (Reviewer): ", r, val)
	fmt.Printf("\n %s, %s", "comm (Reviewer)", comm)

	return comm, nil
} //C(P, r)  C(S, r)

func (rev *Reviewer) GetCommitMessageReviewGrade(val *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(rev.keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r
	r := GetRandomInt(rev.keys.D)

	rev.gradeCommittedValue.r = r
	rev.gradeCommittedValue.val = val
	
	
	x1 := ec.ExpBaseG(rev.keys, val)
	x2 := ec.Exp(rev.keys, &rev.keys.PublicKey, r)
	comm := ec.Mul(rev.keys, x1, x2)
	rev.gradeCommittedValue.CommittedValue = comm
	fmt.Printf("\n %s, %s, %s", "R & Val (Reviewer): ", r, val)
	fmt.Printf("\n %s, %s", "comm (Reviewer)", comm)

	return comm, nil
} //C(P, r)  C(S, r)

func (s *Submitter) GetCommitMessagePaper(val *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(s.keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r
	r := GetRandomInt(s.keys.D) //check up on this

	s.paperCommittedValue.CommittedValue.r = r

	s.paperCommittedValue.CommittedValue.val = val

	x1 := ec.ExpBaseG(s.keys, val)
	x2 := ec.Exp(s.keys, &s.keys.PublicKey, r)
	comm := ec.Mul(s.keys, x1, x2)
	s.paperCommittedValue.CommittedValue.CommittedValue = comm
	fmt.Printf("\n %s, %s, %s", "R & Val: (Submitter)", r, val)
	fmt.Printf("\n %s, %s", "comm (Submitter)", comm)
	return comm, nil
}

//verify
func (s *Submitter) VerifyTrapdoorSubmitter(trapdoor *big.Int) bool {
	h := ec.ExpBaseG(s.keys, trapdoor)
	return Equals(h, &s.keys.PublicKey)
	//Equals(key, &s.keys.PublicKey)
}

/*
//verify
func (s *Submitter) VerifyTrapdoorPaper(trapdoor *big.Int) bool {
	h:= ec.ExpBaseG(s.keys, s.keys.D)
	return Equals(h, &s.Pa)

	hx, hy := p.CommittedValue.Curve.ScalarBaseMult(trapdoor.Bytes())
	key := &ecdsa.PublicKey{p.CommittedValue.Curve, hx, hy}
	return key.Equal(p.CommittedValue)
	//Equals(key, &s.keys.PublicKey)

}*/