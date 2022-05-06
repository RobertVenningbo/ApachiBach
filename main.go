package main

import (
	"database/sql"
	"log"
	"net/http"
	. "swag/components"

	_ "github.com/mattn/go-sqlite3"

	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

func main() {
	database, _ := sql.Open("sqlite3", "./apachi.db") 
	// query := "DROP TABLE IF EXISTS users;"
	// statement, _ := database.Prepare(query)
	// statement.Exec()
	statement, _ := database.Prepare("DROP TABLE IF EXISTS users;")
	statement.Exec()
	statement, _ = database.Prepare("CREATE TABLE users (UserID INTEGER PRIMARY KEY NOT NULL, Username TEXT NOT NULL, Hash TEXT NOT NULL, UserType TEXT NOT NULL, Secret TEXT DEFAULT NULL)")
    statement.Exec()
	  


	app.Route("/", &Log{})
	app.RunWhenOnBrowser()
	http.Handle("/", &app.Handler{
		Name:        "Hello",
		Description: "An Hello World! example",
		Scripts: []string{
			"/web/websocket.js",
		},
	})

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}


