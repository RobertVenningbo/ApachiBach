package main

import (
    "fmt"
    "net/http"
	"html/template"
)

type SwagPage struct {
	Title string
	News string
}

func swagHandler(w http.ResponseWriter, r *http.Request){
	p := SwagPage{Title: "yeahhh", News: "yup no news"}
	t, _ := template.ParseFiles("basictemplating.html")
	t.Execute(w, p)
}

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, you've requested: %s\n", r.URL.Path + "swag")
    })
	http.HandleFunc("/swag", swagHandler)
    http.ListenAndServe(":80", nil)
}
