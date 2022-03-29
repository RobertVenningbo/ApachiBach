package backend

func (r *Reviewer) DetermineGrade(paperId int, grade int) (map[int]int) {
		pList := getPaperList(&pc, r) //This is the papers that the reviewer is grading right?
		for _, p := range pList {
			if p.Id == paperId {
				for paperId, v := range r.gradedPaperMap { //loop for mapping a grade to a paperID - unique per reviewer 
					if v == 0 {
						r.gradedPaperMap[paperId] = grade //map[paperID][grade]
					}
				}
			}
		}
		return r.gradedPaperMap //return map with suggested grades for papers
}

func AgreeOnGrade(reviewers []Reviewer) {
	//loop through all reviewers gradedPaperMap
	//find average grade on a specific paper and round to nearest 4, 7, 10, 12
	//agree on grade on paper and sign paper review commit and review nonce
}
