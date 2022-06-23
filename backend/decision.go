package backend

import (
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"log"
	random "math/rand"
	"swag/model"

	"github.com/0xdecaf/zkrp/ccs08"
)

func (pc *PC) GetKPCSFromLog(pId int) []byte {
	msg := fmt.Sprintf("SignedSubmitMsg%v", pId)
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
	grade := pc.GetGrade(pId)
	reviews := pc.GetReviewsOnly(pId)
	Kpcs := pc.GetKPCSFromLog(pId)

	msgStruct := SendGradeStruct{
		reviews,
		grade,
	}

	EncMsgStruct := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(msgStruct), string(Kpcs))
	str := fmt.Sprintf("PC sends grade and reviews to submitter who submitted paper , %v", pId)
	logmsg := model.Log{
		State: 15,
		LogMsg: str,
		FromUserID: 4000,
		Value: EncMsgStruct[1],
		Signature: EncMsgStruct[0],
	}
	model.CreateLogMsg(&logmsg)
	Trae.Put(str, EncMsgStruct)
}

func (pc *PC) SendGrades(subm *Submitter) { //step 15
	grade := pc.GetGrade(subm.PaperCommittedValue.Paper.Id)
	reviews := pc.GetReviewsOnly(subm.PaperCommittedValue.Paper.Id)
	Kpcs := GenerateSharedSecret(pc, subm, nil)
	msgStruct := SendGradeStruct{
		reviews,
		grade,
	}

	EncMsgStruct := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(msgStruct), Kpcs)

	str := fmt.Sprintf("PC sends grade and reviews to submitter, %v", subm.UserID)

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

var AcceptedPapers []Paper //Global

func (pc *PC) CompileGrades() { //step 17
	grades := []int{}
	for _, p := range AcceptedPapers {
		grade := pc.GetGrade(p.Id)
		grades = append(grades, grade)
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
	Trae.Put(str, signStr)
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

func (pc *PC) RevealAcceptedPaperInfo(pId int) RevealPaper {

	p := pc.GetPaperAndRandomness(pId)
	grades := pc.GetCompiledGrades()

	revealPaperMsg := RevealPaper{
		*p.Paper,
		p.Rs,
	}

	signature := SignsPossiblyEncrypts(pc.Keys, EncodeToBytes(revealPaperMsg), "")
	str := fmt.Sprintf("PC reveals accepted paper: %v", p)
	logmsg := model.Log{
		State:      18,
		LogMsg:     str,
		FromUserID: 4000,
		Value:      signature[1],
		Signature:  signature[0],
	}
	model.CreateLogMsg(&logmsg)
	Trae.Put(str, signature)

	/*NIZK*/
	params, errSetup := ccs08.SetupSet(grades)
	if errSetup != nil {
		log.Panicln(errSetup)
	}
	var i64 int64
	IntGrade := pc.GetGrade(pId)
	i64 = int64(IntGrade)
	//TODO: NOTE THAT WE HAVE TO RANDOMIZE GRADES ish. No duplicates plz
	r, _ := rand.Int(rand.Reader, elliptic.P256().Params().N)
	proof_out, _ := ccs08.ProveSet(i64, r, params)
	result, _ := ccs08.VerifySet(&proof_out, &params)
	if !result {
		log.Panicf("Assert failure: expected true, actual: %v", result)
	} else {
		log.Println("PC proves that grade is in set of compiled grades.")
	}
	return revealPaperMsg
}

func (pc *PC) RandomizeGradesForProof() []RandomizeGradesForProofStruct {
	somethinglist := []RandomizeGradesForProofStruct{}
	grades := pc.GetCompiledGrades()
	for _, g := range grades {
		x := random.Int63n(1844674407370955161)  //some random large number to generate, 1 bits smaller than int64 cap.
		msg := RandomizeGradesForProofStruct{
			R:           x,
			GradeBefore: g,
			GradeAfter:  x + g,
		}
		somethinglist = append(somethinglist, msg)
	}
	return somethinglist
}



/*HELPER METHODS*/

func (pc *PC) AcceptPaper(pId int) { //Helper function, "step 16.5"
	for _, p := range pc.AllPapers {
		if p.Id == pId {
			AcceptedPapers = append(AcceptedPapers, *p)
		}
	}
}

func (pc *PC) GetGrade(pId int) int {
	KpAndRg := pc.GetKpAndRgPC(pId)
	holder := 0
	Kp := KpAndRg.GroupKey
	for _, v := range pc.AllPapers {
		if pId == v.Id {
			holder = v.ReviewerList[0].UserID
		}
	}
	GetStr := fmt.Sprintf("Reviewer %v signed and encrypted grade", holder)
	item := Trae.Find(GetStr)
	if item == nil {
		CheckStringAgainstDB(GetStr)
		item = Trae.Find(GetStr)
	}
	bytes := item.value.([]byte)
	encodedGrade := Decrypt(bytes, Kp.D.String())
	decodedGrade := DecodeToStruct(encodedGrade).(int)
	log.Printf("PC decrypts retrieved encrypted grade for paper %v \n", pId)

	return decodedGrade
}

func (pc *PC) GetReviewsOnly(pId int) []string {
	reviews := []string{}
	for _, v := range pc.AllPapers {
		if pId == v.Id {
			for _, r := range v.ReviewerList {
				result, _ := pc.GetReviewStruct(r)
				reviews = append(reviews, result.Review)
			}
		}
	}
	return reviews
}

//This is for when the application is distributed s.t. a submitter can retrieve its reviews and grade.
func (s *Submitter) RetrieveGrades() SendGradeStruct {
	Kpcs := GenerateSharedSecret(&Pc, s, nil)

	getStr := fmt.Sprintf("PC sends grade and reviews to submitter, %v", s.UserID)
	log.Println(getStr)
	item := Trae.Find(getStr)
	if item == nil {
		CheckStringAgainstDB(getStr)
		item = Trae.Find(getStr)
	}
	bytes := item.value.([]byte)
	encoded := Decrypt(bytes, Kpcs)
	decoded := DecodeToStruct(encoded).(SendGradeStruct)

	return decoded
}
