package backend

import (
	"bytes"
	_ "bytes"
	"crypto/ed25519"
	"encoding/gob"
	_ "errors"
	"fmt"
	_ "fmt"
	"log"
	_ "log"

	random "github.com/mazen160/go-random"
)

type Reviewer struct {
	keys Keys
}

type Submitter struct {
	keys   Keys
	rndaom RandomNumber
	userID string
}

type PC struct {
	keys   Keys
	rndaom RandomNumber
}

type Keys struct {
	PublicK  []byte
	PrivateK []byte
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

func Commit1() {

}

func NIZK() {

}

func newKeys() *Keys {
	a, b, _ := ed25519.GenerateKey(nil)
	return &Keys{a, b}
}

func (s *Submitter) Submit(p *Paper) {
	s.keys = *newKeys()
	rr, _ := random.IntRange(1024, 2048)
	rs, _ := random.IntRange(1024, 2048)

	Encrypt(EncodeToBytes(p), string(s.keys.PrivateK))
	Encrypt(EncodeToBytes(rr), string(s.keys.PrivateK))
	Encrypt(EncodeToBytes(rs), string(s.keys.PrivateK))
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

/*
func Submit(Paper p, PC_SK){
	a,b,c := Encrypt(S.submit(p), rs, rr)

	K_pcs := Ebcr

	submitter := {
		keys
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
