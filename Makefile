all: ghbinmirror-linux-amd64

ghbinmirror-linux-amd64:
	@GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -v -o ghbinmirror-linux-amd64

