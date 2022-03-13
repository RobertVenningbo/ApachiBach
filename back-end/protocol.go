package backend

import (
	"bytes"
	_ "bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	_ "errors"
	"fmt"
	_ "fmt"
	"log"
	_ "log"
	"math/big"
	ec "swag/ec"
)

type Reviewer struct {
	keys      *ecdsa.PrivateKey
	paperList []Paper
}

type Submitter struct {
	keys                    *ecdsa.PrivateKey
	userID                  string
	submitterCommittedValue *CommitStruct //commitstruct
	paperCommittedValue     *Paper
	receiver                *Receiver
	encrypted               []byte
	signatureMap            map[int][]byte
}

type CommitStruct struct {
	committedValue *ecdsa.PublicKey
	r              *big.Int
	val            *big.Int
}

type PC struct {
	keys         *ecdsa.PrivateKey
	signatureMap map[int][]byte
}

type Paper struct {
	id             int
	committedValue *CommitStruct
	selected       bool
}

var (
	pc = PC{
		newKeys(),
		nil,
	}
)

type SubmitStruct struct {
	msg       []byte
	rr        []byte
	rs        []byte
	sharedKey []byte
}

type Receiver struct {
	keys       *ecdsa.PrivateKey
	commitment *ecdsa.PublicKey
}

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

// When receiver receives a decommitment, CheckDecommitment verifies it against the stored value
// (stored by SetCommitment).
func (r *Receiver) CheckDecommitment(R, val *big.Int) bool {
	a := ec.ExpBaseG(r.keys, val)             // g^x
	b := ec.Exp(r.keys, &r.keys.PublicKey, R) // h^r
	c := ec.Mul(r.keys, a, b)                 // g^x * h^r

	return Equals(c, r.commitment)
}

func generateSharedSecret(pc *PC, submitter *Submitter) string {
	publicPC := pc.keys.PublicKey
	privateS := submitter.keys
	shared, _ := publicPC.Curve.ScalarMult(publicPC.X, publicPC.Y, privateS.D.Bytes())

	sharedHash := sha256.Sum256(shared.Bytes())

	return string(sharedHash[:])
}

func newKeys() *ecdsa.PrivateKey {
	a, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	return a
}

func Submit(s *Submitter, p *Paper, c elliptic.Curve) *Submitter {
	s.keys = newKeys()
	rr := GetRandomInt(s.keys.D)
	rs := GetRandomInt(s.keys.D)
	ri := GetRandomInt(s.keys.D)

	log.Println(rr) // shared between all parties
	log.Println(rs) // shared between S and PC
	log.Println(ri) // step 2

	sharedPCS := generateSharedSecret(&pc, s)

	hashedPublicK := sha256.Sum256(EncodeToBytes(pc.keys.PublicKey.X))
	encryptedSharedPCS := Encrypt([]byte(sharedPCS), string(hashedPublicK[:]))

	msg := SubmitStruct{
		Encrypt(EncodeToBytes(p), sharedPCS),
		Encrypt(EncodeToBytes(rr), sharedPCS),
		Encrypt(EncodeToBytes(rs), sharedPCS),
		encryptedSharedPCS,
	}

	s.encrypted = Encrypt(EncodeToBytes(msg), s.keys.D.String()) //encrypted paper and random numbers

	//submitter identity commit
	s.GetCommitMessage(ri)

	//paper identity commit
	s.GetCommitMessagePaper(rs)

	s.paperCommittedValue = p

	hashedMsgSubmit, _ := GetMessageHash([]byte(fmt.Sprintf("%v", s.submitterCommittedValue.committedValue)))
	hashedMsgPaper, _ := GetMessageHash([]byte(fmt.Sprintf("%v", s.paperCommittedValue.committedValue.committedValue)))

	signatureSubmit, _ := ecdsa.SignASN1(rand.Reader, s.keys, hashedMsgSubmit)
	putNextSignatureInMapSubmitter(s, signatureSubmit)

	signaturePaper, _ := ecdsa.SignASN1(rand.Reader, s.keys, hashedMsgPaper)
	putNextSignatureInMapSubmitter(s, signaturePaper)

	log.Printf("\n %s %s", "Ks is revealed to all parties", s.keys.PublicKey) //KS is logged/revealed to all parties??? or is it

	hashedPaperPC, _ := GetMessageHash([]byte(fmt.Sprintf("%v", s.paperCommittedValue.committedValue.committedValue)))
	signaturePaperPC, _ := ecdsa.SignASN1(rand.Reader, pc.keys, hashedPaperPC)
	putNextSignatureInMapPC(&pc, signaturePaperPC)                            //signal next fase
	log.Println("PC signed a paper (submission) " + string(signaturePaperPC)) //PC signed paper commit to indicate the PC will continue the process of getting the paper reviewed

	return s
}

func putNextSignatureInMapSubmitter(s *Submitter, slice []byte) { //not sure if works, test needed.
	for k, v := range s.signatureMap {
		if v == nil {
			s.signatureMap[k] = slice
		}
	}
}

func putNextSignatureInMapPC(p *PC, slice []byte) {
	for k, v := range p.signatureMap {
		if v == nil {
			pc.signatureMap[k] = slice
		}
	}
}

func GetMessageHash(xd []byte) ([]byte, error) {
	md := sha256.New()
	return md.Sum(xd), nil
}

//step 4
func assignPapers(r []Receiver, p []Paper) {

	/*

		foreach receiver in r
			PC signs the paper with its private key
			Generate shared key
			encrypt paper with shared secret key
				do assign papers to receiver
					receiver.paperList = p (but encrypted version)

	*/
}

//step 5
func createBid(r Receiver) {
	/*
		Decrypt papers
		flip boolean in paper struct signaling if they want it or not

		then sign (private key) and encrypt (shared secret key) the bid?

	*/
}

func (s *Submitter) GetCommitMessage(val *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(s.keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r
	r := GetRandomInt(s.keys.D)

	s.submitterCommittedValue.r = r     //hiding factor?
	s.submitterCommittedValue.val = val //den value (random) vi comitter ting til !TODO: ændre i strucsne så det til at finde rundt på
	x1 := ec.ExpBaseG(s.keys, val)
	x2 := ec.Exp(s.keys, &s.keys.PublicKey, r)
	comm := ec.Mul(s.keys, x1, x2)
	s.submitterCommittedValue.committedValue = comm

	return comm, nil
} //C(P, r)  C(S, r)

func (s *Submitter) GetCommitMessagePaper(val *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(s.keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r
	r := GetRandomInt(s.keys.D) //check up on this

	s.paperCommittedValue.committedValue.r = r
	s.paperCommittedValue.committedValue.val = val
	x1 := ec.ExpBaseG(s.keys, val)
	x2 := ec.Exp(s.keys, &s.keys.PublicKey, r)
	comm := ec.Mul(s.keys, x1, x2)
	s.paperCommittedValue.committedValue.committedValue = comm

	return comm, nil
} //C(P, r)  C(S, r)

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

func Equals(e *ecdsa.PublicKey, b *ecdsa.PublicKey) bool {
	return e.X.Cmp(b.X) == 0 && e.Y.Cmp(b.Y) == 0
}

func EncodeToBytes(p interface{}) []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(p)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("uncompressed size (bytes): ", len(buf.Bytes()))
	return buf.Bytes()
}

func GetRandomInt(max *big.Int) *big.Int {
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Fatal(err)
	}
	return n
}
