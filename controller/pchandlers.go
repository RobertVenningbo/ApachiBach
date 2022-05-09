package controller

import (
	"net/http"
	"text/template"
)

func PCHandler(w http.ResponseWriter, r *http.Request) {
	var tpl = template.Must(template.ParseFiles("templates/pc.html"))
	tpl.Execute(w, nil)
}
