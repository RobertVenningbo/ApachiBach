package backend

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/ed25519"
	"encoding/binary"
	"errors"
	"fmt"
	_ "log"

	"github.com/mazen160/go-random"
)
	



type Reviewer struct {
	Keys Types
}

type Submitter struct {
	Keys Types
	rndaom RandomNumber
	userID string
}

type PC struct {
	Keys Types
	rndaom RandomNumber
}

type Types struct {
	PrivateK string
	PublicK string
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

func Commit1(){

}

func NIZK(){

}

func Submit(s *Submitter, p *Paper, rs int, rr int){
	rs, err := random.IntRange(1024, 2048)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(rs)

	rr, err1 := random.IntRange(1024, 2048)
	if err != nil {
		fmt.Println(err1)
	}
	fmt.Println(rr)

	
	a, _ := GenCommitmentKey()
	var bin_buf bytes.Buffer
	binary.Write(&bin_buf, binary.BigEndian, p)
	commit1 := Commit(s.userID, []byte(bin_buf.Bytes()), a)
	b, _ := GenCommitmentKey()
	commit2 := Commit(s.userID, []byte(bin_buf.Bytes()), b)	

	fmt.Print(commit1)
	fmt.Print(commit2)

	
	
}


/*
func Submit(Paper p, PC_SK){
	a,b,c := Encrypt(S.submit(p), rs, rr)

	K_pcs := Ebcr

	submitter := {
		keys := {
			private = crypto.generatenewkey
		}
		rndom := {
			rs
			rr
		}
	}

	PC := {
		keys
		rndom := {
			rs
			rr
		}
		data := {
				
		}
	}
	aes.
}
*/
// https://pkg.go.dev/github.com/coniks-sys/coniks-go



const (
	// commitmentKeyLen should be robust against the birthday attack.
	// One commitment is given for each leaf node throughout time.
	commitmentKeyLen = 16 // 128 bits of security, supports 2^64 nodes.
	// prefix is a string used to make the commitments from this package unique.
	prefix = "Key Transparency Commitment"
)

var (
	hashAlgo = sha256.New
	// key is publicly known random fixed key for use in the HMAC function.
	// This fixed key allows the commitment scheme to be modeled as a random oracle.
	fixedKey = []byte{0x19, 0x6e, 0x7e, 0x52, 0x84, 0xa7, 0xef, 0x93, 0x0e, 0xcb, 0x9a, 0x19, 0x78, 0x74, 0x97, 0x55}
	// ErrInvalidCommitment occurs when the commitment doesn't match the profile.
	ErrInvalidCommitment = errors.New("invalid commitment")
)

// GenCommitmentKey generates a commitment key for use in Commit. This key must
// be kept secret in order to prevent an adversary from learning what data has
// been committed to by a commitment. To unseal and verify a commitment,
// provide this key, along with the data under commitment to the client.
//
// In Key Transparency, the user generates this key, creates a commitment, and
// signs it.  The user uploads the signed commitment along with this key and
// the associated data to the server in order for the server to reveal the
// associated data to senders. This commitment scheme keeps the associated data
// from leeking to anyone that has not explicitly requested it from the server.
func GenCommitmentKey() ([]byte, error) {
	// Generate commitment nonce.
	nonce := make([]byte, commitmentKeyLen)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return nonce, nil
}

// Commit makes a cryptographic commitment under a specific userID to data.
func Commit(userID string, data, nonce []byte) []byte {
	mac := hmac.New(hashAlgo, fixedKey)
	mac.Write([]byte(prefix))
	mac.Write(nonce)

	// Message
	binary.Write(mac, binary.BigEndian, uint32(len(userID)))
	mac.Write([]byte(userID))
	mac.Write(data)

	return mac.Sum(nil)
}

// Verify customizes a commitment with a userID.
func Verify(userID string, commitment, data, nonce []byte) error {
	if got, want := Commit(userID, data, nonce),
		commitment; !hmac.Equal(got, want) {
		return ErrInvalidCommitment
	}
	return nil
}