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

var n25519, _ = new(big.Int).SetString("7237005577332262213973186563042994240857116359379907606001950938285454250989", 10)

type Reviewer struct {
	keys *ecdsa.PrivateKey
}

type Submitter struct {
	keys                    *ecdsa.PrivateKey
	random                  RandomNumber
	userID                  string
	SubmitterCommittedValue *ecdsa.PublicKey //commitstruct
	PaperCommittedValue     Paper
	encrypted               []byte
	signatureMap            map[int][]byte
}

type CommitStruct struct{
	CommittedValue *ecdsa.PublicKey
	r *big.Int
	val *big.Int
}

type PC struct {
	keys   *ecdsa.PrivateKey
	random RandomNumber
}

type RandomNumber struct {
	Rs *big.Int
	Rr *big.Int
	Ri *big.Int
	Rg *big.Int
}

type Paper struct {
	Id             int
	CommittedValue *ecdsa.PublicKey
	random         RandomNumber
}

var (
	pc = PC{
		newKeys(),
		RandomNumber{nil, nil, nil, nil},
	}
)

type SubmitStruct struct {
	msg       []byte
	rr        []byte
	rs        []byte
	sharedKey []byte
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

	s.PaperCommittedValue = *p

	hashedMsgSubmit, _ := GetMessageHash(EncodeToBytes(s.SubmitterCommittedValue))
	hashedMsgPaper, _ := GetMessageHash(EncodeToBytes(p.CommittedValue))

	signatureSubmit, _ := ecdsa.SignASN1(rand.Reader, s.keys, hashedMsgSubmit) //rand.Reader idk??
	signaturePaper, _ := ecdsa.SignASN1(rand.Reader, s.keys, hashedMsgPaper)   //rand.Reader idk??

	ecdsa.VerifyASN1(&s.keys.PublicKey, hashedMsgSubmit, signatureSubmit) //testing
	ecdsa.VerifyASN1(&s.keys.PublicKey, hashedMsgPaper, signaturePaper)   //testing

	//TODO log Ks (reveal Ks to all parties?)

	return s
}

func GetMessageHash(xd []byte) ([]byte, error) {
	md := sha256.New()
	return md.Sum(xd), nil
}

func (s *Submitter) GetCommitMessage(val *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(s.keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r
	r := GetRandomInt(s.keys.D)

	s.random.Rr = r   //hiding factor?
	s.random.Rg = val //den value (random) vi comitter ting til !TODO: ændre i strucsne så det til at finde rundt på
	x1 := ec.ExpBaseG(s.keys, val)
	x2 := ec.Exp(s.keys, &s.keys.PublicKey, r)
	comm := ec.Mul(s.keys, x1, x2)
	s.SubmitterCommittedValue = comm

	return comm, nil
} //C(P, r)  C(S, r)

func (s *Submitter) GetCommitMessagePaper(val *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(s.keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r
	r := GetRandomInt(s.keys.D) //check up on this

	s.PaperCommittedValue.random.Rr = r
	s.PaperCommittedValue.random.Rg =  val
	x1 := ec.ExpBaseG(s.keys, val)
	x2 := ec.Exp(s.keys, &s.keys.PublicKey, r)
	comm := ec.Mul(s.keys, x1, x2)
	s.PaperCommittedValue.CommittedValue = comm

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
