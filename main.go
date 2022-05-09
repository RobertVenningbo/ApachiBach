package main

import (
	"fmt"
	_ "log"
	"net/http"
	_ "net/http"
	"os"
	. "swag/controller"
	_ "swag/backend"
	_ "github.com/lib/pq"
	. "swag/db"
)




func main() {
	var ispctaken bool
	serverport := os.Args[2]
	if os.Args[1] == "submitter" {
		ConnectDB()
		http.HandleFunc("/", SubmissionHandler)
		http.ListenAndServe(":"+serverport, nil)
		// submitter := Submitter{
		// 	NewKeys(),
		// 	-1,
		// 	nil,
		// 	nil,
		// 	nil,
		// }
		
	} else if os.Args[1] == "reviewer" {
		ConnectDB()
		http.HandleFunc("/", SubmissionHandler)
		http.ListenAndServe(":"+serverport, nil)
	} else if os.Args[1] == "pc" {
		if !ispctaken {
			ispctaken = true //TODO: make this work
			ConnectDB()
			http.HandleFunc("/", PCHandler)
			http.ListenAndServe(":"+serverport, nil)
		} else {
			fmt.Println("PC is already running")
			os.Exit(1)
		}
	} else {
		fmt.Println("Wrong arguments given")
		os.Exit(1)
	}


}
