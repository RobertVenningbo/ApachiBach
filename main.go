package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

//var server = GetInstanceOfServer()
var server = GetInstanceOfServer()

func main() {
	go server.Hub.Run()
	app.Route("/", &Log{})
	app.RunWhenOnBrowser()
	http.Handle("/", &app.Handler{
		Name:        "Hello",
		Description: "An Hello World! example",
		Scripts: []string{
			"/web/websocket.js",
		},
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(server.Hub, w, r)
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

type Log struct {
	app.Compo
}

func (l *Log) Render() app.UI {
	return app.Div().ID("wrapper").Body(
		app.Div().ID("log"),
		app.Form().ID("form").Body(
			app.Input().Type("submit").Value("Send"),
			app.Input().Type("text").ID("msg").Size(64).AutoFocus(true),
		),
		app.Button().OnClick(l.onClick),
	)
}

func (l *Log) onClick(ctx app.Context, e app.Event) {
	client, err := server.GetClientById("asd")
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(client.Id)
	fmt.Println("tis")
	
}
