package backend

import (
	"bytes"
	_ "bytes"
	"crypto/ecdsa"
	_ "crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	_ "errors"
	"fmt"
	_ "fmt"
	"log"
	_ "log"
	"math/big"

	"github.com/binance-chain/tss-lib/crypto"
)


type Reviewer struct {
	userID				string
	keys                *ecdsa.PrivateKey
	biddedPaperMap      map[int][]byte
	paperMap            map[int][]byte
	signatureMap        map[int][]byte
	paperCommittedValue *Paper
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
	CommittedValue *ecdsa.PublicKey
	r              *big.Int
	val            *big.Int
}

type PC struct {
	keys         *ecdsa.PrivateKey
	signatureMap map[int][]byte
}

type Paper struct {
	Id                  int
	CommittedValue      *CommitStruct
	Selected            bool
	ReviewSignatureByPC []byte
}

var (
	tree = NewTree(DefaultMinItems)
	pc = PC{
		newKeys(),
		nil,
	}
	paperList     []Paper
	schnorrProofs []SchnorrProof
)

type SubmitStruct struct {
	Msg       []byte
	Rr        []byte
	Rs        []byte
	SharedKey []byte
}

type Receiver struct {
	keys       *ecdsa.PrivateKey
	commitment *ecdsa.PublicKey
}

func generateSharedSecret(pc *PC, submitter *Submitter, reviewer *Reviewer) string {
	publicPC := pc.keys.PublicKey
	var sharedHash [32]byte
	if reviewer == nil {
		privateS := submitter.keys
		shared, _ := publicPC.Curve.ScalarMult(publicPC.X, publicPC.Y, privateS.D.Bytes())
		sharedHash = sha256.Sum256(shared.Bytes())
	} else {
		privateR := reviewer.keys
		shared, _ := publicPC.Curve.ScalarMult(publicPC.X, publicPC.Y, privateR.D.Bytes())
		sharedHash = sha256.Sum256(shared.Bytes())
	}
	
	return string(sharedHash[:])
}

func newKeys() *ecdsa.PrivateKey {
	a, _ := ecdsa.GenerateKey(curve, rand.Reader)
	return a
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

func putNextSignatureInMapReviewer(r *Reviewer, slice []byte) {
	for k, v := range r.signatureMap {
		if v == nil {
			r.signatureMap[k] = slice
		}
	}
}

func putNextPaperInBidMapReviewer(r *Reviewer, slice []byte) {
	for k, v := range r.biddedPaperMap {
		if v == nil {
			r.biddedPaperMap[k] = slice
		}
	}
}

func GetMessageHash(xd []byte) ([]byte, error) {
	md := sha256.New()
	return md.Sum(xd), nil
}



func EcdsaToECPoint(pk *ecdsa.PublicKey) (*crypto.ECPoint, error) {
	return crypto.NewECPoint(pk.Curve, pk.X, pk.Y)
}

// func ECPointToEcdsa(ec *crypto.ECPoint) (*ecdsa.PublicKey){
// 	return &ecdsa.PublicKey{ec.Curve(), ec[0], ec.coords[1]}
// }

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

func DecodeToPaper(s []byte) Paper {

	p := Paper{}
	dec := gob.NewDecoder(bytes.NewReader(s))
	err := dec.Decode(&p)
	if err != nil {
		log.Fatal(err)
	}
	return p
}

func GetRandomInt(max *big.Int) *big.Int {
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		log.Fatal(err)
	}
	return n
}