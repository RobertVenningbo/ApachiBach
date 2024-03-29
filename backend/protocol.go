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
		log.Println("Couldn't retrieve PC")
	}
	var key ecdsa.PrivateKey
	pkeys := DecodeToStruct(pc.PublicKeys).(ecdsa.PublicKey)
	key = ecdsa.PrivateKey{
		PublicKey: pkeys,
		D:         big.NewInt(0),
	}
	Pc.Keys = &key
}

var NoMultipleAppend bool

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
			if NoMultipleAppend { //nasty fix i know
				return
			}
			Pc.AllPapers = append(Pc.AllPapers, paper)
		}
	}
	NoMultipleAppend = true
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

func Equals(e *ecdsa.PublicKey, b *ecdsa.PublicKey) bool {
	return e.X.Cmp(b.X) == 0 && e.Y.Cmp(b.Y) == 0
}

func InitGobs() {
	gob.Register(Paper{})
	gob.Register([]Paper{})
	gob.Register([]*Paper{})
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
	gob.Register(RevealPaper{})
	gob.Register(RejectMessage{})
	gob.Register(SendGradeStruct{})
	gob.Register(ClaimMessage{})
	gob.Register(ValueSignature{})
	gob.Register(Message{})
	gob.Register(ShareReviewsMessage{})
	gob.Register(GradeAndPaper{})
	gob.Register(RandomizeGradesForProofStruct{})
	gob.Register(EqProof{})
}

func EncodeToBytes(p interface{}) []byte {
	buf := bytes.Buffer{}
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(&p)
	if err != nil {
		log.Fatal(err)
	}
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
	return ecdsa.VerifyASN1(pub, hash, signature.([]byte))
}

func (reviewer *Reviewer) VerifyDiscussingMessage(bytes []byte, logStr string) { //Verifies signature for each msg, bad thing is it does it for all messages when refreshing
	var isLegit bool
	for _, r := range reviewer.PaperCommittedValue.Paper.ReviewerList {
		hash, _ := GetMessageHash(bytes)
		var sigmsg model.Log
		model.GetLogMsgByMsg(&sigmsg, logStr)
		sig := sigmsg.Signature
		isLegit = Verify(&r.Keys.PublicKey, sig, hash)
		if isLegit {
			fmt.Printf("\nReviewer %v verifies disccusing message from Reviewer %v", reviewer.UserID, r.UserID)
			return
		}
	}
}

func (pc *PC) VerifyGradesFromReviewers(pId int, msg []byte, logstr string) {
	for _, p := range pc.AllPapers { //Verifies submitted grade for each reviewer, need to double check it works with multiple reviewers
		if pId == p.Id {
			for _, r := range p.ReviewerList {
				hash, _ := GetMessageHash(msg)
				var sigmsg model.Log
				model.GetLogMsgByMsg(&sigmsg, logstr)
				sig := sigmsg.Signature
				isLegit := Verify(&r.Keys.PublicKey, sig, hash)
				if isLegit {
					fmt.Printf("\nPC verifies signature for grade from reviewer %v", r.UserID)
				}
			}
		}
	}
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

func UserToReviewer(user model.User) Reviewer {
	keys := DecodeToStruct(user.PublicKeys).(ecdsa.PublicKey)
	return Reviewer{
		UserID: user.Id,
		Keys: &ecdsa.PrivateKey{
			PublicKey: keys,
			D:         big.NewInt(0),
		},
		PaperCommittedValue: &CommitStructPaper{},
	}
}

func UserToSubmitter(user model.User) Submitter {
	keys := DecodeToStruct(user.PublicKeys).(ecdsa.PublicKey)
	return Submitter{
		UserID: user.Id,
		Keys: &ecdsa.PrivateKey{
			PublicKey: keys,
			D:         big.NewInt(0),
		},
		SubmitterCommittedValue: &CommitStruct{},
		PaperCommittedValue:     &CommitStructPaper{},
		Receiver:                &Receiver{},
	}
}
