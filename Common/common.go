package common

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	. "swag/backend"
)

func getFile(filename string) *os.File {
	path := RootDir()
	path += "\\temp-files\\"
	path += filename
	file, err := os.Open(path)
	if err != nil {
		log.Panic("getFile doesn't work")
	}
	return file
}

func RootDir() string {
	_, b, _, _ := runtime.Caller(0)
	d := path.Join(path.Dir(b))
	return filepath.Dir(d)
}

func ReadFileByteData(filename string) ([]byte, error) {
	path := RootDir()
	path += "\\temp-files\\"
	path += filename
	plaintext, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return plaintext, nil
}

func main() {
	fmt.Println(EncodeToBytes(getFile("1614288602.pdf")))
	fmt.Println(ReadFileByteData("1614288602.pdf"))
}
