package backend

import (
	"crypto/ecdsa"
	"fmt"
	"log"
	Math "math"
	"swag/model"
)

func (r *Reviewer) SendSecretMsgToReviewers(input string) { //intended to be for step 12, repeated.
	KpAndRg := r.GetReviewKpAndRg()
	Kp := KpAndRg.GroupKey
<<<<<<< Updated upstream
	encryptedSignedMsg := SignsPossiblyEncrypts(r.Keys, []byte(input), Kp.D.String())
	logStr := fmt.Sprintf("Discussing message for paper: %v", r.PaperCommittedValue.Paper.Id)
	log.Println(logStr)
=======
	encryptedSignedMsg := SignsPossiblyEncrypts(r.Keys, EncodeToBytes(input), Kp.D.String())
	logStr := fmt.Sprintf("Sending msg to the log by reviewer: %v", r.UserID)
	logmsg := model.Log{
		State: 12,
		LogMsg: logStr,
		FromUserID: r.UserID,
		Value: encryptedSignedMsg[1],
		Signature: encryptedSignedMsg[0],
	}
	model.CreateLogMsg(&logmsg)
>>>>>>> Stashed changes
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
	data := DiscussingViewData{
		Title: r.PaperCommittedValue.Paper.Title,
		Msgs:  messages,
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
		State: 12, //unsure of state
		LogMsg: msg,
		FromUserID: r.UserID,
		Value: encryptedSignedGradeStruct[1],
		Signature: encryptedSignedGradeStruct[0],
	}
	model.CreateLogMsg(&logmsg)

	Trae.Put(msg, encryptedSignedGradeStruct)
	log.Println(msg)

}

func (r *Reviewer) GetGradeForReviewer(rId int) IndividualGrade {
	msg := fmt.Sprintf("Reviewer%v graded a paper", rId)
	gradeStruct := Trae.Find(msg)
	if gradeStruct == nil {
		CheckStringAgainstDB(msg)
		gradeStruct = Trae.Find(msg)
	}

	KpAndRg := r.GetReviewKpAndRg()
	Kp := KpAndRg.GroupKey
	bytes := gradeStruct.value.([]byte)
	encodedGradeStruct := Decrypt(bytes, Kp.D.String())
	decodedGradeStruct := DecodeToStruct(encodedGradeStruct).(IndividualGrade)
	return decodedGradeStruct
}

func AgreeOnGrade(paper *Paper) int {
	result := 0
	length := len(paper.ReviewerList)
	for _, r := range paper.ReviewerList {
		gradeStruct := r.GetGradeForReviewer(r.UserID)
		result += gradeStruct.Grade
	}
	avg := float64(result) / float64(length)
	grade := CalculateNearestGrade(avg)

	return grade
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
	if found == nil { //Unsure about checking against DB here, double check this when it's to be used
		KpAndRg := r.GetReviewKpAndRg()
		Rg := KpAndRg.Rg
		grade := AgreeOnGrade(r.PaperCommittedValue.Paper)
		GradeBigInt := MsgToBigInt(EncodeToBytes(grade))
		GradeCommit, _ := r.GetCommitMessageReviewGrade(GradeBigInt, Rg)

		logmsg := model.Log{
			State: 13, //unsure about state
			LogMsg: str,
			FromUserID: r.UserID,
			Value: EncodeToBytes(grade),
			Signature: nil, //nothing signed
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
	signedGradeReviewCommits := SignsPossiblyEncrypts(r.Keys, EncodeToBytes(gradeReviewCommits), "")
	str := fmt.Sprintf("Reviewer %v signed GradeReviewCommits", r.UserID)
	logmsg := model.Log{
		State: 13,
		LogMsg: str,
		FromUserID: r.UserID,
		Value: signedGradeReviewCommits[1],
		Signature: signedGradeReviewCommits[0],
	}
	model.CreateLogMsg(&logmsg)
	Trae.Put(str, signedGradeReviewCommits)

}

func (r *Reviewer) SignAndEncryptGrade() { //Expected to be called for every reviewer
	grade := AgreeOnGrade(r.PaperCommittedValue.Paper) //acquire agreed grade
	KpAndRg := r.GetReviewKpAndRg()
	Kp := KpAndRg.GroupKey
	signedGrade := SignsPossiblyEncrypts(r.Keys, EncodeToBytes(grade), Kp.D.String()) //Notice Kp.(ecdsa.PrivateKey).D.String() seems super fishy, plz work.
	submitStr := fmt.Sprintf("Reviewer %v signed and encrypted grade", r.UserID)

	logmsg := model.Log{
		State: 14,
		LogMsg: submitStr,
		FromUserID: r.UserID,
		Value: signedGrade[1],
		Signature: signedGrade[0],
	}
	model.CreateLogMsg(&logmsg)
	Trae.Put(submitStr, signedGrade)
}
