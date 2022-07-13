package backend

import (
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
	"swag/model"

	"github.com/0xdecaf/zkrp/ccs08"
)

func (pc *PC) GetKPCSFromLog(pId int) []byte {
	msg := fmt.Sprintf("SignedSubmitMsg %v", pId)
	item := Trae.Find(msg)
	if item == nil {
		CheckStringAgainstDB(msg)
		item = Trae.Find(msg)
	}

	
	bytes := item.value.([]byte)
	decodedSubmitMessage := DecodeToStruct(bytes)
	submitMessage := decodedSubmitMessage.(SubmitMessage)
	kpcs := Decrypt(submitMessage.EncryptedKpcs, pc.Keys.X.String())


	return kpcs
}

func (pc *PC) SendGrades2(pId int) { //Step 15 new
	GradeAndPaper := pc.GetGradeAndPaper(pId)
	reviews := pc.GetReviewsOnly(pId)
	Kpcs := pc.GetKPCSFromLog(pId)

	msgStruct := SendGradeStruct{
		reviews,
		int(GradeAndPaper.GradeBefore),
	}

	EncMsgStruct := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(msgStruct), string(Kpcs))
	str := fmt.Sprintf("PC sends grade and reviews to submitter who submitted paper , %v", pId)
	logmsg := model.Log{
		State:      15,
		LogMsg:     str,
		FromUserID: 4000,
		Value:      EncMsgStruct[1],
		Signature:  EncMsgStruct[0],
	}
	model.CreateLogMsg(&logmsg)
	Trae.Put(str, EncMsgStruct)
}

/*PC DECLINES PAPER PATH*/

func (pc *PC) RejectPaper(pId int) RejectMessage { //step 16
	GradeAndPaper := pc.GetGradeAndPaper(pId)

	KpAndRg := pc.GetKpAndRgPC(pId)
	Rg := KpAndRg.Rg
	ReviewSignedStruct := pc.GetReviewSignedStruct(pId)
	ReviewCommit := ReviewSignedStruct.Commit

	// type RejectMessage struct {
	// 	Commit *ecdsa.PublicKey
	// 	Grade  int
	// 	Rg     *big.Int
	// }
	// maybe check the commit like the protocol describes any third party can do
	rejectMsg := RejectMessage{
		ReviewCommit,
		int(GradeAndPaper.GradeBefore),
		Rg,
	}

	signature := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(rejectMsg), "")

	str := fmt.Sprintf("PC rejects Paper: %v", pId)
	logmsg := model.Log{
		State:      16,
		LogMsg:     str,
		FromUserID: 4000,
		Value:      signature[1],
		Signature:  signature[0],
	}
	model.CreateLogMsg(&logmsg)
	Trae.Put(str, signature)

	return rejectMsg
}

/*PC ACCEPTS PAPER PATH*/

var AcceptedPapers []RandomizeGradesForProofStruct //Global

func (pc *PC) CompileGrades() { //step 17
	if len(AcceptedPapers) == 0 {
		log.Println("****** \n SILENT CRASHING \n ******")
	}
	grades := []int{}
	for _, p := range AcceptedPapers {
		GradeAndPaper := pc.GetGradeAndPaper(p.PaperId)
		grades = append(grades, int(GradeAndPaper.GradeAfter))
	}

	signStr := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(grades), "")
	str := fmt.Sprint("PC compiles grades")
	logmsg := model.Log{
		State:      17,
		LogMsg:     str,
		FromUserID: 4000,
		Value:      signStr[1],
		Signature:  signStr[0],
	}
	model.CreateLogMsg(&logmsg)
	Trae.Put(str, signStr[1])
}

func (pc *PC) GetCompiledGrades() []int64 {
	getStr := fmt.Sprintf("PC compiles grades")
	item := Trae.Find(getStr)
	if item == nil {
		CheckStringAgainstDB(getStr)
		item = Trae.Find(getStr)
	}

	bytes := item.value.([]byte)
	DecodedGrades := DecodeToStruct(bytes).([]int)

	var i64 []int64
	for _, v := range DecodedGrades {
		i64 = append(i64, int64(v))
	}
	return i64
}

func (pc *PC) RevealAllAcceptedPapers() {

	// ***SETUP PHASE***
	params, errSetup := ccs08.SetupSet(pc.GetCompiledGrades())
	// *** *** *** ***

	for _, v := range AcceptedPapers {
		p := pc.GetPaperAndRandomness(v.PaperId)

		revealPaperMsg := RevealPaper{
			*p.Paper,
			p.Rs,
		}

		signature := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(revealPaperMsg), "")
		str := fmt.Sprintf("PC reveals accepted paper: %v", p.Paper.Id)
		logmsg := model.Log{
			State:      18,
			LogMsg:     str,
			FromUserID: 4000,
			Value:      signature[1],
			Signature:  signature[0],
		}
		model.CreateLogMsg(&logmsg)
		Trae.Put(str, signature[1])

		/*NIZK*/
		if errSetup != nil {
			log.Panicln(errSetup)
		}
		IntGrade := pc.GetGradeAndPaper(v.PaperId)
		r, _ := rand.Int(rand.Reader, elliptic.P256().Params().N)
		proof_out, _ := ccs08.ProveSet(IntGrade.GradeAfter, r, params)
		result, _ := ccs08.VerifySet(&proof_out, &params)
		if !result {
			log.Panicf("Assert failure: expected true, actual: %v", result)
		} else {
			log.Println("PC proves that grade is in set of compiled grades.")
		}
		nizkStr := fmt.Sprintf("PC uploads grade NIZK for P%v", p.Paper.Id)
		signatureNizk := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(revealPaperMsg), "")
		nizkLogMsg := model.Log{
			State:      18,
			LogMsg:     nizkStr,
			FromUserID: 4000,
			Value:      signatureNizk[1],
			Signature:  signatureNizk[0],
		}
		model.CreateLogMsg(&nizkLogMsg)
		Trae.Put(nizkStr, signatureNizk[1])
	}
}

/*HELPER METHODS*/

func (pc *PC) CheckAcceptedPapers(pId int) bool {

	for _, p := range AcceptedPapers {
		if p.PaperId == pId {
			return true
		}
	}
	return false
}

func (pc *PC) GetAcceptedPapers() []*Paper {
	var acceptedPapers []*Paper

	for _, p := range pc.AllPapers {
		if pc.CheckAcceptedPapers(p.Id) {
			acceptedPapers = append(acceptedPapers, p)
		}
	}
	return acceptedPapers

}

func (pc *PC) AcceptPaper(pId int) { //Helper function, "step 16.5"

	if pc.CheckAcceptedPapers(pId) { //such that we dont get any duplicates
		return
	}
	for _, p := range pc.AllPapers {
		if p.Id == pId {
			//AcceptedPapers = append(AcceptedPapers, *p)
			str := fmt.Sprintf("PC accepts Paper: %v", pId)
			logmsg := model.Log{
				State:      16,
				LogMsg:     str,
				FromUserID: 4000,
			}
			model.CreateLogMsg(&logmsg)
			GradeAndPaper := pc.GetGradeAndPaper(pId)
			AcceptedPapers = append(AcceptedPapers, GradeAndPaper)
		}
	}
}

func (pc *PC) CheckAllSignedGrades(pId int) bool {
	for _, v := range pc.AllPapers {
		if pId == v.Id {
			for _, r := range v.ReviewerList {
				GetStr := fmt.Sprintf("Reviewer %v signed and encrypted grade", r.UserID)
				exists := model.DoesLogMsgExist(GetStr)
				if !exists {
					log.Println("aboooort mission")
					return false
				}
			}
		}
	}
	return true
}

func (pc *PC) GetGradeAndPaper(pId int) RandomizeGradesForProofStruct {
	// AOK := pc.CheckAllSignedGrades(pId)
	// if !AOK {
	// 	return RandomizeGradesForProofStruct{R: -1, GradeBefore: -1, GradeAfter: -1, PaperId: -1}
	// }
	holder := 0
	for _, v := range pc.AllPapers {
		if pId == v.Id {
			holder = v.ReviewerList[0].UserID
		}
	}
	GetStr := fmt.Sprintf("Reviewer %v signed and encrypted grade", holder)
	KpAndRg := pc.GetKpAndRgPC(pId)
	Kp := KpAndRg.GroupKey

	item := Trae.Find(GetStr)
	if item == nil {
		CheckStringAgainstDB(GetStr)
		item = Trae.Find(GetStr)
	}
	bytes := item.value.([]byte)
	encodedGradeAndPaper := Decrypt(bytes, Kp.D.String())
	decodedGradeAndPaper := DecodeToStruct(encodedGradeAndPaper).(RandomizeGradesForProofStruct)
	pc.VerifyGradesFromReviewers(pId, encodedGradeAndPaper, GetStr)
	log.Printf("\nPC decrypts retrieved encrypted grade for paper %v \n", pId)

	return decodedGradeAndPaper
}


func (pc *PC) GetReviewsOnly(pId int) []ReviewStruct {
	reviews := []ReviewStruct{}
	for _, v := range pc.AllPapers {
		if pId == v.Id {
			for _, r := range v.ReviewerList {
				result, _ := pc.GetReviewStruct(r)
				reviews = append(reviews, result)
			}
		}
	}
	return reviews
}

func (s *Submitter) RetrieveGradeAndReviews() SendGradeStruct {
	Kpcs := GenerateSharedSecret(&Pc, s, nil)

	getStr := fmt.Sprintf("PC sends grade and reviews to submitter who submitted paper , %v", s.PaperCommittedValue.Paper.Id)
	log.Println(getStr)
	item := Trae.Find(getStr)
	if item == nil {
		CheckStringAgainstDB(getStr)
		item = Trae.Find(getStr)
	}

	bytes := item.value.([]byte)
	encoded := Decrypt(bytes, Kpcs)
	decoded := DecodeToStruct(encoded).(SendGradeStruct)

	isLegit := VerifySignature(getStr, encoded, &Pc.Keys.PublicKey)
	if !isLegit {
		fmt.Printf("\n Submitter %v couldn't verify PC signature when collecting grades & reviews", s.UserID)
	} else {
		fmt.Printf("\n Submitter %v verifies PC signature when collecting reviews", s.UserID)
	}


	return decoded
}
