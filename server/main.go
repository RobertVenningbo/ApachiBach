// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"text/template"
)

var (
	addr    = flag.String("addr", ":8080", "http service address")
	tpl     *template.Template
	funcMap template.FuncMap
)

func FuncMappings() { //TODO: SETUP FUNCMAP WITH `tpl` for passing data/functions to DOM
	funcMap = template.FuncMap{
		// The name "title" is what the function will be called in the template text.
		"submitter": strings.Title,
	}

}

func serveHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// err := tpl.ExecuteTemplate(w, "home.html", funcMap)
	// if err != nil {
	// 	log.Panic("Panic in serveHome")
	// }
	http.ServeFile(w, r, "home.html")
}

func testfunc(){
	fmt.Print("test")
}

func main() {
	FuncMappings()
	flag.Parse()
	hub := newHub()
	go hub.run()
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
