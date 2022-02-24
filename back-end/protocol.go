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

	random "github.com/mazen160/go-random"
)

var n25519, _ = new(big.Int).SetString("7237005577332262213973186563042994240857116359379907606001950938285454250989", 10)

type Reviewer struct {
	keys ecdsa.PrivateKey
}

type Submitter struct {
	keys   ecdsa.PrivateKey
	rndaom RandomNumber
	userID string
	encrypted []byte
}

type PC struct {
	keys   ecdsa.PrivateKey
	rndaom RandomNumber
}

type RandomNumber struct {
	Rs int
	Rr int
	Ri int
	Rg int
}

type Paper struct {
	Id int
}

var pc = PC{
	 *newKeys(),
	 RandomNumber{0,0,0,0},
	}

type SubmitStruct struct {
	msg []byte
	rr []byte
	rs []byte
	sharedKey []byte
}
// Commit to a value x
// H - Random secondary point on the curve
// r - Private key used as blinding factor
// x - The value (number of tokens)
func commitTo(H *ecdsa.PublicKey, r, x *ecdsa.Scalar) ecdsa.PublicKey {
	//ec.g.mul(r).add(H.mul(x));
	var result, rPoint, transferPoint ecdsa.PublicKey
	rPoint.ScalarMultBase(r)
	transferPoint.ScalarMult(H, x)
	result.Add(&rPoint, &transferPoint)
	return result
}

// Generate a random point on the curve
func generateH() ecdsa.PublicKey {
	var random ecdsa.Scalar
	var H ecdsa.PublicKey
	random.Rand()
	H.ScalarMultBase(&random)
	return H
}

// Subtract two commitments using homomorphic encryption
func Sub(cX, cY *ristretto.Point) ristretto.Point {
	var subPoint ristretto.Point
	subPoint.Sub(cX, cY)
	return subPoint
}

// Subtract two known values with blinding factors
//   and compute the committed value
//   add rX - rY (blinding factor private keys)
//   add vX - vY (hidden values)
func SubPrivately(H *ristretto.Point, rX, rY *ristretto.Scalar, vX, vY *big.Int) ristretto.Point {
	var rDif ristretto.Scalar
	var vDif big.Int
	rDif.Sub(rY, rX)
	vDif.Sub(vX, vY)
	vDif.Mod(&vDif, n25519)

	var vScalar ristretto.Scalar
	var rPoint ristretto.Point
	vScalar.SetBigInt(&vDif)

	rPoint.ScalarMultBase(&rDif)
	var vPoint, result ristretto.Point
	vPoint.ScalarMult(H, &vScalar)
	result.Add(&rPoint, &vPoint)
	return result
}
func Commit(msg []byte, numb *big.Int) {
	
}

func NIZK() {

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

func Submit(s *Submitter, p *Paper) *Submitter{
	s.keys = *newKeys()
	rr, _ := random.IntRange(1024, 2048)
	rs, _ := random.IntRange(1024, 2048)

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

// https://pkg.go.dev/github.com/coniks-sys/coniks-go
