build:
	GOARCH=wasm GOOS=js go build -o web/app.wasm *.go
	go build

run: build
	./swag