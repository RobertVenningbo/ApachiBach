package main

import (
	"fmt"
	"log"

	"github.com/bitfield/script"
)

func main() {
	reviewercount := 5000
	submittercount := 3000
	pcint := 4000
	outputrevstr := fmt.Sprintf("go run main.go reviewer %v", 0)
	outputsubstr := fmt.Sprintf("go run main.go submitter %v", 0)
	outputpcstr := fmt.Sprintf("go run main.go pc %v", pcint)

	fmt.Println("Welcome to Apachi, please input the amount of agents.")
	fmt.Println("Please input the amount of submitters: ")
	var subnumber int
	_, err := fmt.Scan(&subnumber)
	if err != nil {
		log.Fatal("expected a number")
	}

	fmt.Println("Please input the amount of reviewers: ")
	var revnumber int
	_, err = fmt.Scan(&revnumber)
	if err != nil {
		log.Fatal("expected a number")
	}
	// script.Exec(outputpcstr)
	str := "\nPC:             http://localhost:4000" + "\n"
	for i := submittercount; i < submittercount+subnumber; i++ {
		outputsubstr = fmt.Sprintf("go run main.go submitter %v", i)
		ii := fmt.Sprint(i)
		str += "Submitter " + ii + ": http://localhost:" + ii + " \n"
		script.Exec(outputsubstr)
	}

	for i := reviewercount; i < reviewercount+revnumber; i++ {
		outputrevstr = fmt.Sprintf("go run main.go reviewer %v", i)
		ii := fmt.Sprint(i)
		str += "Reviewer " + ii + ":  http://localhost:" + ii + " \n"
		script.Exec(outputrevstr)
	}
	fmt.Println(str)
	script.Exec(outputpcstr).Stdout()
}
