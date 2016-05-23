all:
	@@go build
linux:
	@@GOOS=linux GOARCH=386 go build
