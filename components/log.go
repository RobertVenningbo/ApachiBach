package components

import (
	"github.com/maxence-charriere/go-app/v9/pkg/app"
)

type Log struct {
	app.Compo
}


func (l *Log) Render() app.UI{
	return app.Div().ID("wrapper").Body(
		app.Div().ID("log"),
		app.Form().ID("form").Body(
			app.Input().Type("submit").Value("Send"),
			app.Input().Type("text").ID("msg").Size(64).AutoFocus(true),

		),
	)
}