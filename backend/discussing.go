package backend

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	Math "math"
	random "math/rand"
	"swag/model"
)

func (r *Reviewer) SendSecretMsgToReviewers(input string) { //intended to be for step 12, repeated.
	KpAndRg := r.GetReviewKpAndRg()
	Kp := KpAndRg.GroupKey
	encryptedSignedMsg := SignsPossiblyEncrypts(r.Keys, []byte(input), Kp.D.String())
	logStr := fmt.Sprintf("Discussing message for paper: %v", r.PaperCommittedValue.Paper.Id)
	log.Println(logStr)
	Trae.Put(logStr, encryptedSignedMsg)

	logmsg := model.Log{
		State:      12,
		LogMsg:     logStr,
		FromUserID: r.UserID,
		Value:      encryptedSignedMsg[1],
		Signature:  encryptedSignedMsg[0],
	}
	model.CreateLogMsg(&logmsg)
}

func (r *Reviewer) GetSecretMsgsFromReviewers() DiscussingViewData {
	logStr := fmt.Sprintf("Discussing message for paper: %v", r.PaperCommittedValue.Paper.Id)
	var messages []string
	var logmsgs []model.Log
	model.GetAllLogMsgsByMsg(&logmsgs, logStr)
	var logsmsgsnotbinded []model.Log

	logsmsgsnotbinded = append(logsmsgsnotbinded, logmsgs...)
	if len(logsmsgsnotbinded) > 0 {
		reviewkpandrg := r.GetReviewKpAndRg()
		Kp := reviewkpandrg.GroupKey
		for _, v := range logsmsgsnotbinded {
			bytes := Decrypt(v.Value, Kp.D.String())
			messages = append(messages, string(bytes))
		}
	}
	reviewStruct := r.GetCollectedReviews()
	data := DiscussingViewData{
		Title:   r.PaperCommittedValue.Paper.Title,
		Msgs:    messages,
		Reviews: reviewStruct,
	}

	return data
}

func (r *Reviewer) GradePaper(grade int) {
	gradeStruct := IndividualGrade{
		r.PaperCommittedValue.Paper.Id,
		r.UserID,
		grade,
	}

	KpAndRg := r.GetReviewKpAndRg()
	Kp := KpAndRg.GroupKey
	encryptedSignedGradeStruct := SignsPossiblyEncrypts(r.Keys, EncodeToBytes(gradeStruct), Kp.D.String())
	msg := fmt.Sprintf("Reviewer%v graded a paper", r.UserID)
	logmsg := model.Log{
		State:      12,
		LogMsg:     msg,
		FromUserID: r.UserID,
		Value:      encryptedSignedGradeStruct[1],
		Signature:  encryptedSignedGradeStruct[0],
	}
	model.CreateLogMsg(&logmsg)
	Trae.Put(msg, encryptedSignedGradeStruct[1])
}

func (r *Reviewer) GetGradeForReviewer(rId int) *IndividualGrade {
	msg := fmt.Sprintf("Reviewer%v graded a paper", rId)
	gradeStruct := Trae.Find(msg)
	if gradeStruct == nil {
		CheckStringAgainstDB(msg)
		gradeStruct = Trae.Find(msg)
		return nil
	}

	bytes := gradeStruct.value.([]byte)

	KpAndRg := r.GetReviewKpAndRg()
	Kp := KpAndRg.GroupKey
	encodedGradeStruct := Decrypt(bytes, Kp.D.String())
	decodedGradeStruct := DecodeToStruct(encodedGradeStruct).(IndividualGrade)
	return &decodedGradeStruct
}

func (r *Reviewer) AgreeOnGrade2(paper *Paper) GradeAndPaper {
	gradeandpaper := r.GetAgreedGrade(paper.Id)
	if gradeandpaper != nil {
		return *gradeandpaper
	}

	return *gradeandpaper
}

func (r *Reviewer) CheckAllSubmittedGrades() bool {
	for _, v := range r.PaperCommittedValue.Paper.ReviewerList {
		if r.GetGradeForReviewer(v.UserID) == nil {
			return false
		}
	}
	return true
}

func (r *Reviewer) RandomizeGrades(grade int64, paperId int) *RandomizeGradesForProofStruct {
	x := random.Int63n(1844674407370955161) //some random large number to generate, 1 bits smaller than int64 cap.
	return &RandomizeGradesForProofStruct{
		R:           x,
		GradeBefore: grade,
		GradeAfter:  grade + x,
		PaperId:     paperId,
	}
}

func (r *Reviewer) PublishAgreedGrade() {
	result := 0
	papir := r.PaperCommittedValue.Paper
	length := len(papir.ReviewerList)

	for _, r := range papir.ReviewerList {
		gradeStruct := r.GetGradeForReviewer(r.UserID)
		result += gradeStruct.Grade
	}

	avg := float64(result) / float64(length)
	grade := CalculateNearestGrade(avg)
	randomGradeStruct := r.RandomizeGrades(int64(grade), papir.Id)
	///CONTINIUE HERE
	logmsg := model.Log{
		State:      13,
		LogMsg:     "All grades have been submitted",
		FromUserID: r.UserID,
	}
	model.CreateLogMsg(&logmsg)
	KpAndRg := r.GetReviewKpAndRg()
	EncryptedGradeStruct := SignsPossiblyEncrypts(r.Keys, EncodeToBytes(msg), KpAndRg.GroupKey.D.String())
	str := fmt.Sprintf("Reviewers agreed on a grade for paper%v", 2)
	logmsg2 := model.Log{
		State:      13,
		LogMsg:     str,
		FromUserID: r.UserID,
		Value:      EncryptedGradeStruct[1],
		Signature:  EncryptedGradeStruct[0],
	}
	model.CreateLogMsg(&logmsg2)
	Trae.Put(str, EncryptedGradeStruct[1])

	return msg
}

func (r *Reviewer) AgreeOnGrade(paper *Paper) GradeAndPaper {
	gradeandpaper := r.GetAgreedGrade(paper.Id)

	if gradeandpaper != nil {
		return *gradeandpaper
	}

	var gradeStruct *IndividualGrade
	result := 0
	length := len(paper.ReviewerList)
	for _, r := range paper.ReviewerList {
		gradeStruct = r.GetGradeForReviewer(r.UserID)
		if gradeStruct == nil { //TODO refactor so log messages aren't needed here
			logmsg := model.Log{
				State:      12,
				LogMsg:     "Not all grades have been submitted",
				FromUserID: r.UserID,
			}
			model.CreateLogMsg(&logmsg)
			return GradeAndPaper{}
		}
		logmsg := model.Log{
			State:      13,
			LogMsg:     "All grades have been submitted",
			FromUserID: r.UserID,
		}
		model.CreateLogMsg(&logmsg)
		result += gradeStruct.Grade
	}
	avg := float64(result) / float64(length)
	grade := CalculateNearestGrade(avg)
	msg := GradeAndPaper{
		Grade: int64(grade),
		Papir: *paper,
	}
	KpAndRg := r.GetReviewKpAndRg()
	EncryptedGradeStruct := SignsPossiblyEncrypts(r.Keys, EncodeToBytes(msg), KpAndRg.GroupKey.D.String())
	str := fmt.Sprintf("Reviewers agreed on a grade for paper%v", paper.Id)
	logmsg := model.Log{
		State:      13,
		LogMsg:     str,
		FromUserID: r.UserID,
		Value:      EncryptedGradeStruct[1],
		Signature:  EncryptedGradeStruct[0],
	}
	model.CreateLogMsg(&logmsg)
	Trae.Put(str, EncryptedGradeStruct[1])

	return msg
}

func (r *Reviewer) GetAgreedGrade(pId int) *GradeAndPaper {
	str := fmt.Sprintf("Reviewers agreed on a grade for paper%v", pId)
	item := Trae.Find(str)
	if item == nil {
		CheckStringAgainstDB(str)
		item = Trae.Find(str)
	}
	bytes := item.value.([]byte)
	KpAndRg := r.GetReviewKpAndRg()
	encodedAgreedGrade := Decrypt(bytes, KpAndRg.GroupKey.D.String())
	agreedGrade := DecodeToStruct(encodedAgreedGrade).(GradeAndPaper)
	return &agreedGrade
}
func CalculateNearestGrade(avg float64) int {
	closest := 999
	minDiff := 999.0
	possibleGrades := []int{-3, 00, 02, 4, 7, 10, 12}
	var diff float64
	for _, v := range possibleGrades {
		diff = Math.Abs(float64(v) - avg)
		if minDiff > diff {
			minDiff = diff
			closest = v
		}
	}
	return closest
}

func (r *Reviewer) MakeGradeCommit() *ecdsa.PublicKey {
	str := fmt.Sprintf("GradeCommit for P%v has been made", r.PaperCommittedValue.Paper.Id)
	found := Trae.Find(str)
	if found == nil {
		KpAndRg := r.GetReviewKpAndRg()
		Rg := KpAndRg.Rg
		gradeStruct := r.AgreeOnGrade(r.PaperCommittedValue.Paper)
		GradeBigInt := MsgToBigInt(EncodeToBytes(gradeStruct.Grade)) //This Grade needs to be randomized
		GradeCommit, _ := r.GetCommitMessageReviewGrade(GradeBigInt, Rg)
		logmsg := model.Log{
			State:      13, //unsure about state
			LogMsg:     str,
			FromUserID: r.UserID,
			Value:      EncodeToBytes(gradeStruct.Grade),
			Signature:  nil, //nothing signed
		}
		model.CreateLogMsg(&logmsg)
		Trae.Put(str, EncodeToBytes(GradeCommit))
		return GradeCommit
	} else {
		EncodedGradeCommit := found.value.([]byte)
		GradeCommit := DecodeToStruct(EncodedGradeCommit).(ecdsa.PublicKey)
		return &GradeCommit
	}
}

func (r *Reviewer) SignCommitsAndNonce() { //Step 13, assumed to be ran when reviewers have settled on a grade
	GradeCommit := r.MakeGradeCommit()
	ReviewCommitNonceStruct := r.GetReviewCommitNonceStruct()
	PaperReviewCommit := ReviewCommitNonceStruct.Commit
	Nonce := ReviewCommitNonceStruct.Nonce

	gradeReviewCommits := GradeReviewCommits{
		PaperReviewCommit,
		GradeCommit,
		Nonce,
	}

	fmt.Printf("%#v\n", gradeReviewCommits)
	signedGradeReviewCommits := SignsPossiblyEncrypts(r.Keys, EncodeToBytes(gradeReviewCommits), "")
	str := fmt.Sprintf("Reviewer %v signed GradeReviewCommits", r.UserID)
	logmsg := model.Log{
		State:      13, //Maybe change the state to 14
		LogMsg:     str,
		FromUserID: r.UserID,
		Value:      signedGradeReviewCommits[1],
		Signature:  signedGradeReviewCommits[0],
	}
	model.CreateLogMsg(&logmsg)
	Trae.Put(str, signedGradeReviewCommits)
}

func (r *Reviewer) SignAndEncryptGrade() { //Expected to be called for every reviewer
	gradeStruct := r.AgreeOnGrade(r.PaperCommittedValue.Paper) //acquire agreed grade
	KpAndRg := r.GetReviewKpAndRg()
	Kp := KpAndRg.GroupKey
	gradeandstruct := GradeAndPaper{
		Grade: int64(gradeStruct.Grade), //Needs to be randomized
		Papir: *r.PaperCommittedValue.Paper,
	}
	signedGrade := SignsPossiblyEncrypts(r.Keys, EncodeToBytes(gradeandstruct), Kp.D.String())
	submitStr := fmt.Sprintf("Reviewer %v signed and encrypted grade", r.UserID)
	logmsg := model.Log{
		State:      14,
		LogMsg:     submitStr,
		FromUserID: r.UserID,
		Value:      signedGrade[1],
		Signature:  signedGrade[0],
	}
	model.CreateLogMsg(&logmsg)
	Trae.Put(submitStr, signedGrade[1])
}
