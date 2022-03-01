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
	"io"
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
	keys               *ecdsa.PrivateKey
	random             RandomNumber
	userID             string
	SubmitterCommittedValue *ecdsa.PublicKey
	PaperCommittedValue Paper
	encrypted          []byte
	signatureMap	   map[int][]byte
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
	Id int
	CommittedValue *ecdsa.PublicKey
	random RandomNumber
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


func NIZK() {
	ec.NewPrivateKey()
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
	rr := ec.GetRandomInt(s.keys.D)
	rs := ec.GetRandomInt(s.keys.D)
	ri := ec.GetRandomInt(s.keys.D)
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
	p.GetCommitMessage(rs)
	
	s.PaperCommittedValue = *p

	hashedMsgSubmit, _ := GetMessageHash(EncodeToBytes(s.SubmitterCommittedValue))
	hashedMsgPaper, _ := GetMessageHash(EncodeToBytes(p.CommittedValue))

	signatureSubmit, _ := ecdsa.SignASN1(rand.Reader, s.keys, hashedMsgSubmit) //rand.Reader idk??
	signaturePaper, _ := ecdsa.SignASN1(rand.Reader, s.keys, hashedMsgPaper)//rand.Reader idk??


	ecdsa.VerifyASN1(&s.keys.PublicKey, hashedMsgSubmit, signatureSubmit) //testing
	ecdsa.VerifyASN1(&s.keys.PublicKey, hashedMsgPaper, signaturePaper) //testing

	//TODO log Ks (reveal Ks to all parties?)
	
	return s
}

func GetMessageHash(xd []byte) ([]byte, error){
	md := sha256.New()
	return md.Sum(xd), nil
}

func (s *Submitter) GetCommitMessage(val *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(s.keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r
	r := ec.GetRandomInt(s.keys.D)

	s.random.Rr = r   //nøglen til boksen?
	s.random.Rg = val //den value (random) vi comitter ting til
	x1, y1 := s.keys.PublicKey.Curve.ScalarBaseMult(val.Bytes())
	x2, y2 := s.keys.PublicKey.Curve.ScalarMult(s.keys.X, s.keys.Y, val.Bytes())
	comm1, comm2 := s.keys.Curve.Add(x1, y1, x2, y2)
	s.SubmitterCommittedValue = &ecdsa.PublicKey{s.keys.Curve, comm1, comm2}

	return s.SubmitterCommittedValue, nil
} //C(P, r)  C(S, r)


func (p *Paper) GetCommitMessage(val *big.Int) (*ecdsa.PublicKey, error) {
	if val.Cmp(n25519) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	// c = g^x * h^r
	r := ec.GetRandomInt(n25519) //check up on this

	p.random.Rr = r   //nøglen til boksen?
	p.random.Rg = val //den value (random) vi comitter ting til
	x1, y1 := p.CommittedValue.Curve.ScalarBaseMult(val.Bytes())
	x2, y2 := p.CommittedValue.Curve.ScalarMult(p.CommittedValue.X, p.CommittedValue.Y, val.Bytes())
	comm1, comm2 := p.CommittedValue.Curve.Add(x1, y1, x2, y2)
	p.CommittedValue = &ecdsa.PublicKey{p.CommittedValue.Curve, comm1, comm2}

	return p.CommittedValue, nil
} //C(P, r)  C(S, r)

//verify
func (s *Submitter) VerifyTrapdoor(trapdoor *big.Int) bool {
	hx, hy := s.keys.PublicKey.Curve.ScalarBaseMult(trapdoor.Bytes())
	key := &ecdsa.PublicKey{s.keys.Curve, hx, hy}
	return key.Equal(s.keys) 
	//Equals(key, &s.keys.PublicKey)
}


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
