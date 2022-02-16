package main

import (
	"html/template"
	"log"
	"net/http"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("templates/*.gohtml"))
}

type SubmissionPage struct {
	FName  string
	LName  string
	Email  string
	Title  string
	Secret string
}

type LogPage struct {
	Timestamp int
	News      string
}

func main() {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/submission", submissionHandler)
	http.HandleFunc("/review", reviewHandler)
	http.HandleFunc("/log", logHandler)
	http.HandleFunc("/discussion", discussionHandler)
	//http.HandleFunc("/claim", swagHandler)
	http.ListenAndServe(":80", nil)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "home.gohtml", nil)
	
	if err != nil {
		log.Println("LOGGED", err)
		http.Error(w, "failuree", http.StatusInternalServerError)
	}
	
}

func submissionHandler(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "submission.gohtml", nil)

	if err != nil {
		log.Println("LOGGED", err)
		http.Error(w, "failuree", http.StatusInternalServerError)
	}
}

func reviewHandler(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "review.gohtml", nil)

	if err != nil {
		log.Println("LOGGED", err)
		http.Error(w, "failuree", http.StatusInternalServerError)
	}
}

func discussionHandler(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "discussion.gohtml", nil)

	if err != nil {
		log.Println("LOGGED", err)
		http.Error(w, "failuree", http.StatusInternalServerError)
	}
}

func logHandler(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "log.gohtml", nil)

	if err != nil {
		log.Println("LOGGED", err)
		http.Error(w, "failuree", http.StatusInternalServerError)
	}
}
