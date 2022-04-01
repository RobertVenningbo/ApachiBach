package backend

import (
	"fmt"
	"log"
)

type SendGradeStruct struct{
	reviews interface{} //idk if this type
	grade interface{}
}


//needs more love
func (pc *PC) SendGrades(subm *Submitter) { //maybe dont use *Submitter as parameter but call gRPC method later on which gets the pub key
	grade := "get the grade"     //retrieve grade
	putStr := fmt.Sprint("Sharing reviews with Reviewers")
	listSignatureItem := tree.Find(putStr) //these are the reviews
	listSignature := listSignatureItem.value    //cast this to list somehow
	Kpcr := generateSharedSecret(pc, subm, nil) //jf. kommentar ved metodenavn
	list := []string{}
	for _, v := range listSignature {
		//maybe verify signature
		_, txt := SplitSignz(v)
		list = append(list, txt)
	}
	msgStruct := SendGradeStruct{
		signatureAndTextOfList,
		grade,
	}
	signatureAndTextOfStruct := SignzAndEncrypt(pc.keys, list, Kpcr)

	str := fmt.Sprintf("fix me")
	log.Println(str)
	tree.Put(str, signatureAndTextOfStruct)
}
