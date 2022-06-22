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
	"swag/model"
)

var (
	Trae = NewTree(DefaultMinItems)
	Pc   = PC{
		nil,
		nil,
	}
)

func InitLocalPC() {
	pc := model.User{}
	err := model.GetPC(&pc)
	if err != nil {
		log.Fatalf("Couldn't retrieve PC")
	}
	var key ecdsa.PrivateKey
	pkeys := DecodeToStruct(pc.PublicKeys).(ecdsa.PublicKey)
	key = ecdsa.PrivateKey{
		PublicKey: pkeys,
		D:         big.NewInt(0),
	}
	Pc.Keys = &key
}

func InitLocalPCPaperList() {
	msgs := []model.Log{}
	model.GetAllLogMsgs(&msgs)
	for _, msg := range msgs {
		if strings.Contains(msg.LogMsg, "SignedSubmitMsg") {
			encodedSubmitMsg := msg.Value
			submitMsg := DecodeToStruct(encodedSubmitMsg).(SubmitMessage)
			fmt.Println(Pc.Keys.PublicKey.X.String())
			fmt.Println(submitMsg.EncryptedKpcs)
			decryptedKpcs := Decrypt(submitMsg.EncryptedKpcs, Pc.Keys.PublicKey.X.String())
			decryptedPaperAndRandomness := Decrypt(submitMsg.PaperAndRandomness, string(decryptedKpcs))
			paperAndRandomess := DecodeToStruct(decryptedPaperAndRandomness).(SubmitStruct)
			paper := paperAndRandomess.Paper
			Pc.AllPapers = append(Pc.AllPapers, paper)
			fmt.Printf("pc paper length: %v \n", len(Pc.AllPapers))
		}
	}
}

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

func InitGobs() {
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
	gob.Register(ValueSignature{})
	gob.Register(Message{})
	gob.Register(ShareReviewsMessage{})
}

func EncodeToBytes(p interface{}) []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
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

func Verify(pub *ecdsa.PublicKey, signature interface{}, hash []byte) bool {

	//signBytes := EncodeToBytes(signature)

	return ecdsa.VerifyASN1(pub, hash, signature.([]byte))
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
