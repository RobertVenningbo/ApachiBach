package backend

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	_ "errors"
	"fmt"
	"log"
	"math/big"
	"strings"
)

var (
	tree = NewTree(DefaultMinItems)
	pc   = PC{
		NewKeys(),
		nil,
	}
)

func GenerateSharedSecret(pc *PC, submitter *Submitter, reviewer *Reviewer) string {
	publicPC := pc.Keys.PublicKey
	var sharedHash [32]byte
	if reviewer == nil {
		privateS := submitter.Keys
		shared, _ := publicPC.Curve.ScalarMult(publicPC.X, publicPC.Y, privateS.D.Bytes())
		sharedHash = sha256.Sum256(shared.Bytes())
	} else {
		privateR := reviewer.Keys
		shared, _ := publicPC.Curve.ScalarMult(publicPC.X, publicPC.Y, privateR.D.Bytes())
		sharedHash = sha256.Sum256(shared.Bytes())
	}

	return string(sharedHash[:])
}

func NewKeys() *ecdsa.PrivateKey {
	a, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	return a
}

func GetMessageHash(xd []byte) ([]byte, error) {
	md := sha256.New()
	return md.Sum(xd), nil
}

// func ECPointToEcdsa(ec *crypto.ECPoint) (*ecdsa.PublicKey){
// 	return &ecdsa.PublicKey{ec.Curve(), ec[0], ec.coords[1]}
// }

func Equals(e *ecdsa.PublicKey, b *ecdsa.PublicKey) bool {
	return e.X.Cmp(b.X) == 0 && e.Y.Cmp(b.Y) == 0
}

//TODO make init func for registering when starting server
func EncodeToBytes(p interface{}) []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	gob.Register(Paper{})
	gob.Register(ReviewSignedStruct{})
	gob.Register(CommitStruct{})
	gob.Register(CommitStructPaper{})
	gob.Register(Submitter{})
	gob.Register(ecdsa.PublicKey{})
	gob.Register(elliptic.P256())
	gob.Register(SubmitStruct{})
	gob.Register(SubmitMessage{})
	gob.Register(CommitMsg{})
	gob.Register(big.Int{})
	gob.Register(PaperBid{})
	gob.Register(Reviewer{})
	gob.Register(ReviewStruct{})
	gob.Register([]ReviewStruct{})
	gob.Register(ReviewKpAndRg{})
	gob.Register(ReviewCommitNonceStruct{})
	gob.Register(IndividualGrade{})
	gob.Register(GradeReviewCommits{})
	gob.Register(Grade{})
	gob.Register(RevealPaper{})
	gob.Register(RejectMessage{})
	gob.Register(SendGradeStruct{})
	gob.Register(ClaimMessage{})
	err := enc.Encode(&p)
	if err != nil {
		log.Fatal(err)
	}
	//fmt.Println("uncompressed size (bytes): ", len(buf.Bytes()))
	return buf.Bytes()
}

func DecodeToStruct(s []byte) (x interface{}) { //Decodes encoded struct to struct https://gist.github.com/SteveBate/042960baa7a4795c3565
	i := x
	dec := gob.NewDecoder(bytes.NewReader(s))
	err := dec.Decode(&i)
	if err != nil {
		log.Fatal(err)
	}
	return i
}

func DecodeToStruct1(s []byte, x interface{}) interface{} { //Decodes encoded struct to struct https://gist.github.com/SteveBate/042960baa7a4795c3565
	i := x
	dec := gob.NewDecoder(bytes.NewReader(s))
	err := dec.Decode(&i)
	if err != nil {
		log.Fatal(err)
	}
	return i
}

func DecodeToPaper(s []byte) *Paper {
	p := Paper{}
	dec := gob.NewDecoder(bytes.NewReader(s))
	err := dec.Decode(&p)
	if err != nil {
		log.Fatal(err)
	}
	return &p
}

func Sign(priv *ecdsa.PrivateKey, plaintext interface{}) string { //TODO; current bug is that the hash within this function is not the same hash as when taking the hash of the returned plaintext
	formatted := fmt.Sprintf("%v", plaintext)
	bytes := []byte(formatted)
	hash, _ := GetMessageHash(bytes)
	//fmt.Printf("%s%v\n", "Hash from Sign func:", hash)
	signature, _ := ecdsa.SignASN1(rand.Reader, priv, hash)
	//fmt.Printf("%s%v \n", "Sig from sign func:", signature)
	return fmt.Sprintf("%v%s%v", signature, "|", plaintext)
}

func Verify(pub *ecdsa.PublicKey, signature interface{}, hash []byte) bool {

	//signBytes := EncodeToBytes(signature)

	return ecdsa.VerifyASN1(pub, hash, signature.([]byte))
}

// CURRENTLY DEPRECATED, USE /SignsPossiblyEncrypts & SplitSignatureAndMsg/
func SignzAndEncrypt(priv *ecdsa.PrivateKey, plaintext interface{}, passphrase string) string {

	bytes := EncodeToBytes(plaintext)

	hash, _ := GetMessageHash(bytes)
	signature, _ := ecdsa.SignASN1(rand.Reader, priv, hash)

	encrypted := Encrypt(bytes, passphrase)

	if passphrase == "" {
		return fmt.Sprintf("%v%s%v", signature, "|", bytes) // Check if "|" interfere with any binary?
	} else {
		return fmt.Sprintf("%v%s%v", signature, "|", encrypted) // Check if "|" interfere with any binary?
	}
	//return [213, 123, 12, 392...]|someEncryptedString
}

// CURRENTLY DEPRECATED, USE /SignsPossiblyEncrypts & SplitSignatureAndMsg/
func SplitSignz(str string) (string, string) { //returns splitArr[0] = signature, splitArr[1] = encrypted
	splitArr := strings.Split(str, "|")
	if len(splitArr) > 2 {
		log.Panic("panic len > 2")
	}
	return splitArr[0], splitArr[1]
}

func SignsPossiblyEncrypts(priv *ecdsa.PrivateKey, bytes []byte, passphrase string) [][]byte { //signs and possibly encrypts a message
	hash, _ := GetMessageHash(bytes)
	signature, _ := ecdsa.SignASN1(rand.Reader, priv, hash)

	x := [][]byte{}

	if passphrase == "" { //if passphrase is empty dont encrypt
		x = append(x, signature, bytes)
		return x
	} else {
		encrypted := Encrypt(bytes, passphrase)
		x = append(x, signature, encrypted)
		return x
	}
}

func SplitSignatureAndMsg(bytes [][]byte) ([]byte, []byte) { // returns signature and msg or encrypted msg
	sig, msg := bytes[0], bytes[1]
	return sig, msg
}