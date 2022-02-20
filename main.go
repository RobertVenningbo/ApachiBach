package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	ab "swag/back-end"
)

var tpl *template.Template

func init() {
	tpl = template.Must(template.ParseGlob("front-end/templates/*.gohtml"))
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
	http.HandleFunc("/upload", uploadFile)
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

func uploadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")
	err := tpl.ExecuteTemplate(w, "upload.gohtml", nil)

	if err != nil {
		log.Println("LOGGED", err)
		http.Error(w, "failuree", http.StatusInternalServerError)
	}
	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("myFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)
	// Create a temporary file within our temp-files directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile("temp-files", "upload-*.pdf")
	if err != nil {
		fmt.Println(err)
	}
	defer tempFile.Close()

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	xa := ab.Encrypt(fileBytes, "password")
	// write this byte array to our temporary file
	tempFile.Write(ab.Decrypt(xa, "password"))

	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File\n")
}
