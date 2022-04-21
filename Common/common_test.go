package common

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"os"
	backend "swag/back-end"
	"testing"

	"github.com/clarketm/json"

	"github.com/stretchr/testify/assert"
)

type someStruct struct {
	Name string            `json:"name"`
	Id   int               `json:"id"`
	file *os.File          `json:"-"`
	keys *ecdsa.PrivateKey `json:"-"`
}

func newKeys() *ecdsa.PrivateKey {
	a, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	return a
}

func TestJson(t *testing.T) {
	s := someStruct{
		"thename",
		2,
		getFile("1614288602.pdf"),
		newKeys(),
	}

	fmt.Printf("%#v \n", s)

	bytes, _ := json.Marshal(s)
	fmt.Printf("%v \n", bytes)
	var sTemp someStruct
	json.Unmarshal(bytes, &sTemp)
	fmt.Printf("%v \n", sTemp)
	assert.Equal(t, s, sTemp, "Marshal failed")
}

func TestFileEncrypt(t *testing.T) {
	file := getFile("1614288602.pdf") //this is a previously uploaded file
	encodedFile := backend.EncodeToBytes(file)
	encryptedFile := backend.Encrypt(encodedFile, "privateKey")
	decodedFile := backend.Decrypt(encryptedFile, "privateKey")
	got := backend.DecodeToStruct(decodedFile)
	assert.Equal(t, file, got.(*os.File), "TestFileEncrypt Failed")
}
