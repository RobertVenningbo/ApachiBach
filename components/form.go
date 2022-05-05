package components

import "github.com/maxence-charriere/go-app/v9/pkg/app"

type Form struct {
	app.Compo
	name string
}

func(f *Form) Render() app.UI {
	return app.Div().Body(
		app.H1().Text("Hello " + f.name),
		app.P().Body(
			app.Input().
				Type("text").
				Value(f.name).
				Placeholder("What is your name?").
				AutoFocus(true).
				// Here the username is directly mapped from the input's change
				// event.
				OnChange(f.ValueTo(&f.name)),
		),
	)
}