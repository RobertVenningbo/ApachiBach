package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	ab "swag/back-end"
	"unicode"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

var tpl *template.Template
var db *sql.DB

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
	var err error
	db, err = sql.Open("mysql", "root:password@tcp(localhost:3306)/testdb")
	if err != nil {
		panic(err.Error())
	}
	defer db.Close()
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/submission", submissionHandler)
	http.HandleFunc("/review", reviewHandler)
	http.HandleFunc("/log", logHandler)
	http.HandleFunc("/discussion", discussionHandler)
	http.HandleFunc("/upload", uploadFile)
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/registerauth", registerAuthHandler)
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

//https://github.com/GrowAdept/youtube/tree/main/gowebdev/register
// registerHandler serves form for registring new users
func registerHandler(w http.ResponseWriter, r *http.Request) {
	err := tpl.ExecuteTemplate(w, "register.gohtml", nil)

	if err != nil {
		log.Println("LOGGED", err)
		http.Error(w, "failuree", http.StatusInternalServerError)
	}
	fmt.Println("*****registerHandler running*****")
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
	xa := ab.Encrypt(fileBytes, "password") //HUSK
	// write this byte array to our temporary file
	tempFile.Write(ab.Decrypt(xa, "password")) //HUSK

	// return that we have successfully uploaded our file!
	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

//https://github.com/GrowAdept/youtube/tree/main/gowebdev/register
// registerAuthHandler creates new user in database
func registerAuthHandler(w http.ResponseWriter, r *http.Request) {
	/*
		1. check username criteria
		2. check password criteria
		3. check if username is already exists in database
		4. create bcrypt hash from password
		5. insert username and password hash in database
		(email validation will be in another video)
	*/
	fmt.Println("*****registerAuthHandler running*****")
	r.ParseForm()
	username := r.FormValue("username")
	// check username for only alphaNumeric characters
	var nameAlphaNumeric = true
	for _, char := range username {
		// func IsLetter(r rune) bool, func IsNumber(r rune) bool
		// if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
		if unicode.IsLetter(char) == false && unicode.IsNumber(char) == false {
			nameAlphaNumeric = false
		}
	}
	// check username length
	var nameLength bool
	if 5 <= len(username) && len(username) <= 50 {
		nameLength = true
	}
	// check password criteria
	password := r.FormValue("password")
	fmt.Println("password:", password, "\npswdLength:", len(password))
	// variables that must pass for password creation criteria
	var pswdLowercase, pswdUppercase, pswdNumber, pswdSpecial, pswdLength, pswdNoSpaces bool
	pswdNoSpaces = true
	for _, char := range password {
		switch {
		// func IsLower(r rune) bool
		case unicode.IsLower(char):
			pswdLowercase = true
		// func IsUpper(r rune) bool
		case unicode.IsUpper(char):
			pswdUppercase = true
		// func IsNumber(r rune) bool
		case unicode.IsNumber(char):
			pswdNumber = true
		// func IsPunct(r rune) bool, func IsSymbol(r rune) bool
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			pswdSpecial = true
		// func IsSpace(r rune) bool, type rune = int32
		case unicode.IsSpace(int32(char)):
			pswdNoSpaces = false
		}
	}
	if 11 < len(password) && len(password) < 60 {
		pswdLength = true
	}
	fmt.Println("pswdLowercase:", pswdLowercase, "\npswdUppercase:", pswdUppercase, "\npswdNumber:", pswdNumber, "\npswdSpecial:", pswdSpecial, "\npswdLength:", pswdLength, "\npswdNoSpaces:", pswdNoSpaces, "\nnameAlphaNumeric:", nameAlphaNumeric, "\nnameLength:", nameLength)
	if !pswdLowercase || !pswdUppercase || !pswdNumber || !pswdSpecial || !pswdLength || !pswdNoSpaces || !nameAlphaNumeric || !nameLength {
		tpl.ExecuteTemplate(w, "register.gohtml", "please check username and password criteria")
		return
	}
	// check if username already exists for availability
	stmt := "SELECT UserID FROM bcrypt WHERE username = ?"
	row := db.QueryRow(stmt, username)
	var uID string
	err := row.Scan(&uID)
	if err != sql.ErrNoRows {
		fmt.Println("username already exists, err:", err)
		tpl.ExecuteTemplate(w, "register.gohtml", "username already taken")
		return
	}
	// create hash from password
	var hash []byte
	// func GenerateFromPassword(password []byte, cost int) ([]byte, error)
	hash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Println("bcrypt err:", err)
		tpl.ExecuteTemplate(w, "register.gohtml", "there was a problem registering account")
		return
	}
	fmt.Println("hash:", hash)
	fmt.Println("string(hash):", string(hash))
	// func (db *DB) Prepare(query string) (*Stmt, error)
	var insertStmt *sql.Stmt
	insertStmt, err = db.Prepare("INSERT INTO bcrypt (Username, Hash) VALUES (?, ?);")
	if err != nil {
		fmt.Println("error preparing statement:", err)
		tpl.ExecuteTemplate(w, "register.gohtml", "there was a problem registering account")
		return
	}
	defer insertStmt.Close()
	var result sql.Result
	//  func (s *Stmt) Exec(args ...interface{}) (Result, error)
	result, err = insertStmt.Exec(username, hash)
	rowsAff, _ := result.RowsAffected()
	lastIns, _ := result.LastInsertId()
	fmt.Println("rowsAff:", rowsAff)
	fmt.Println("lastIns:", lastIns)
	fmt.Println("err:", err)
	if err != nil {
		fmt.Println("error inserting new user")
		tpl.ExecuteTemplate(w, "register.gohtml", "there was a problem registering account")
		return
	}
	fmt.Fprint(w, "Your account has been successfully created, "+username+".")

	http.Redirect(w, r, "/", 200) //virker ikke helt, men vil have den redirecter efter.
}
