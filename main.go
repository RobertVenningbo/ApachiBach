package main

import (
	"log"
	"net/http"
	. "swag/components"
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func main() {
	app.Route("/", &Form{})
	app.RunWhenOnBrowser()
	http.Handle("/", &app.Handler{
		Name:        "Hello",
		Description: "An Hello World! example",
		Scripts: []string{
			"/web/script.js",
		},
	})


	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleGreet(ctx app.Context, a app.Action) {
	name, ok := a.Value.(string)
	if !ok {
		return
	}

	// Setting a state named "greet-name" with the name value.
	ctx.SetState("greet-name", name)
}