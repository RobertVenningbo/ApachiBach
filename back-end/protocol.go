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
	keys   *ecdsa.PrivateKey
	rndaom RandomNumber
	userID string
	ComittedValue *ecdsa.PublicKey
}

type PC struct {
	keys   ecdsa.PrivateKey
	rndaom RandomNumber
}

type RandomNumber struct {
	Rs *big.Int
	Rr *big.Int
	Ri *big.Int
	Rg *big.Int
}

type Paper struct {
	Id int
}

var( 
	
	pc = PC{
	 *newKeys(),
	 RandomNumber{0,0,0,0},
	}


)

type SubmitStruct struct {
	msg []byte
	rr []byte
	rs []byte
	sharedKey []byte
}



func Commit(msg []byte, numb *big.Int) {
	
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

func Submit(s *Submitter, p *Paper, c elliptic.Curve) *Submitter{
	s.keys = *newKeys()
	rr, _ := ecdsa.GenerateKey(c, rand.Reader)
	rs, _ := ecdsa.GenerateKey(c, rand.Reader)

	sharedPCS := generateSharedSecret(&pc, s)

	hashedPublicK := sha256.Sum256(EncodeToBytes(pc.keys.PublicKey.X))
	encryptedSharedPCS := Encrypt([]byte(sharedPCS), string(hashedPublicK[:]))

	msg := SubmitStruct{
		Encrypt(EncodeToBytes(p), sharedPCS),
		Encrypt(EncodeToBytes(rr), sharedPCS),
		Encrypt(EncodeToBytes(rs), sharedPCS),
		encryptedSharedPCS,
	}
	
	s.encrypted = Encrypt(EncodeToBytes(msg), s.keys.D.String())
	
	

	return s
}

func (s *Submitter) GetCommitMessage(val *big.Int) (*ecdsa.PublicKey, error){
	if val.Cmp(s.keys.D) == 1 || val.Cmp(big.NewInt(0)) == -1 {
		err := fmt.Errorf("the committed value needs to be in Z_q (order of a base point)")
		return nil, err
	}

	 // c = g^x * h^r
	r := ec.GetRandomInt(s.keys.D)

	s.rndaom.Rr = r //nøglen til boksen?
	s.rndaom.Rg = val //den value (random) vi comitter ting til
	x1, y1 := s.keys.PublicKey.Curve.ScalarBaseMult(val.Bytes())
	x2, y2 := s.keys.PublicKey.Curve.ScalarMult(s.keys.X, s.keys.Y, val.Bytes())
	comm1, comm2 := s.keys.Curve.Add(x1, y1, x2, y2) 
	s.ComittedValue = &ecdsa.PublicKey{nil, comm1, comm2}

	return s.ComittedValue, nil
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


