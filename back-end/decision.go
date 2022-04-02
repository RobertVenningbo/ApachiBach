package backend

import (
	"fmt"
	"log"
)

type SendGradeStruct struct {
	reviews interface{} //idk if this type
	grade   interface{}
}

type RejectMessage struct {
	commit interface{}
	grade  interface{}
	rg     interface{}
}

//needs >>a little<< more love
func (pc *PC) SendGrades(subm *Submitter) { //maybe dont use *Submitter as parameter but call gRPC method later on which gets the pub key
	grade := "get the grade" //retrieve grade
	putStr := fmt.Sprint("Sharing reviews with Reviewers")
	listSignatureItem := tree.Find(putStr)              //these are the reviews
	listSignature := listSignatureItem.value.([]string) //cast this to list somehow
	Kpcr := generateSharedSecret(pc, subm, nil)         //jf. kommentar ved metodenavn
	list := []string{}
	for _, v := range listSignature {
		//maybe verify signature
		_, txt := SplitSignz(v)
		list = append(list, txt)
	}
	signatureAndTextOfStruct := SignzAndEncrypt(pc.keys, list, Kpcr)

	msgStruct := SendGradeStruct{
		signatureAndTextOfStruct,
		grade,
	}
	str := fmt.Sprintf("PC sends grades to submitter, %s", subm.userID)
	log.Println(str)
	tree.Put(str, msgStruct)
}

/*PC DECLINES PAPER PATH*/

func (pc *PC) RejectPaper(rUserID string, grade string) { //step 16
	grade = "get the grade somehow, right now it's assumed that it's given from somewhere else"
	getRgStr := fmt.Sprintf("PC sign and encrypt Rg with Kpcr between PC and reviewer id %s", rUserID)

	rgItem := tree.Find(getRgStr)
	sigAndRgVal := rgItem.value.(string)
	_, rg := SplitSignz(sigAndRgVal)
	getReviewCommitStruct := fmt.Sprintf("%s signs paper review commit \n", rUserID)

	reviewCommitStructItem := tree.Find(getReviewCommitStruct)

	_, commitStruct := SplitSignz(reviewCommitStructItem.value.(string))

	rejectMsg := RejectMessage{
		commitStruct, //notice commitStruct also contains nonce, might break security properties. Delete this comment when you find out.
		grade,
		rg,
	}

	signature := Sign(pc.keys, rejectMsg)

	logMsg := fmt.Sprintf("Following paper was rejected: %s", signature)
	log.Println(logMsg)
	tree.Put(logMsg, logMsg)

}

/*PC ACCEPTS PAPER PATH*/

func (pc *PC) CompileGrades() { //step 17
	grades := "get the grades somehow, right now it's assumed that it's given from somewhere else"

	signStr := Sign(pc.keys, grades)

	str := fmt.Sprint("PC compiles grades")
	log.Println(str)
	tree.Put(str, signStr)
}

func (pc *PC) RevealAcceptedPaperInfo() {
	/*
		Reveal Paper p and randomness rs
			Sign these values
		NIZK proof which proofs that grade g is port of the set of compiled grades
		of the accepted papers and that it's the hiding factor in the grade commit
		C(g, rg)
	*/

	p := "paper"
	rs := "rs"
	grades := "grades"

}
