package backend

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"

	"github.com/0xdecaf/zkrp/ccs08"
)

type SendGradeStruct struct {
	reviews []string
	grade   int
}

type RejectMessage struct {
	commit *ecdsa.PublicKey
	grade  int
	rg     *big.Int
}

type RevealPaper struct {
	Paper Paper
	Rs    *big.Int
}

func (pc *PC) SendGrades(subm *Submitter) { //step 15
	grade := pc.GetGrade(subm.PaperCommittedValue.Paper.Id)
	reviews := pc.GetReviewsOnly(subm.PaperCommittedValue.Paper.Id)
	Kpcs := generateSharedSecret(pc, subm, nil)
	msgStruct := SendGradeStruct{
		reviews,
		grade,
	}

	EncMsgStruct := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(msgStruct), Kpcs)

	str := fmt.Sprintf("PC sends grade and reviews to submitter, %v", subm.UserID)
	log.Println(str)
	tree.Put(str, EncMsgStruct)
}

/*PC DECLINES PAPER PATH*/

func (pc *PC) RejectPaper(pId int) { //step 16
	Grade := pc.GetGrade(pId)

	KpAndRg := pc.GetKpAndRgPC(pId)
	Rg := KpAndRg.Rg
	ReviewSignedStruct := pc.GetReviewSignedStruct(pId)
	ReviewCommit := ReviewSignedStruct.Commit

	rejectMsg := RejectMessage{
		ReviewCommit,
		Grade,
		Rg,
	}

	signature := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(rejectMsg), "")

	logMsg := fmt.Sprintf("PC rejects Paper: %v", pId)
	log.Println(logMsg)
	tree.Put(logMsg, signature)
}

/*PC ACCEPTS PAPER PATH*/

var AcceptedPapers []Paper //Global

func (pc *PC) AcceptPaper(pId int) { //Helper function, "step 16.5"
	for _, p := range pc.allPapers {
		if p.Id == pId {
			AcceptedPapers = append(AcceptedPapers, *p)
		}
	}
}

func (pc *PC) CompileGrades() { //step 17
	grades := []int{}
	for _, p := range AcceptedPapers {
		grade := pc.GetGrade(p.Id)
		grades = append(grades, grade)
	}

	signStr := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(grades), "")
	str := fmt.Sprint("PC compiles grades")
	log.Println(str)
	tree.Put(str, signStr)
}

func (pc *PC) GetCompiledGrades() []int {
	getStr := fmt.Sprintf("PC compiles grades")
	item := tree.Find(getStr).value.([][]byte)
	_, EncodedGrades := SplitSignatureAndMsg(item)
	DecodedGrades := DecodeToStruct(EncodedGrades).([]int)
	return DecodedGrades
}

func (pc *PC) RevealAcceptedPaperInfo(s *Submitter) {

	p := pc.GetPaperAndRandomness(s.PaperCommittedValue.Paper.Id)

	revealPaperMsg := RevealPaper{
		*p.Paper,
		p.Rs,
	}

	signature := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(revealPaperMsg), "")
	str := fmt.Sprintf("PC reveals accepted paper: %v", p)
	log.Println(str)
	tree.Put(str, signature)

	/*NIZK*/
	// params := cc08.SetupSet()

}

func (pc *PC) GetGrade(pId int) int {
	KpAndRg := pc.GetKpAndRgPC(pId)
	holder := 0
	Kp := KpAndRg.GroupKey
	for _, v := range pc.allPapers {
		if pId == v.Id {
			holder = v.ReviewerList[0].UserID
		}
	}
	GetStr := fmt.Sprintf("Reviewer %v signed and encrypted grade", holder)
	item := tree.Find(GetStr).value.([][]byte)

	_, enc := SplitSignatureAndMsg(item)

	encodedGrade := Decrypt(enc, Kp.D.String())
	decodedGrade := DecodeToStruct(encodedGrade).(int)
	log.Printf("PC decrypts retrieved encrypted grade for paper %v \n", pId)

	return decodedGrade
}

func (pc *PC) GetReviewsOnly(pId int) []string {
	reviews := []string{}
	for _, v := range pc.allPapers {
		if pId == v.Id {
			for _, r := range v.ReviewerList {
				result, _ := pc.GetReviewStruct(r)
				reviews = append(reviews, result.Review)
			}
		}
	}
	return reviews
}

