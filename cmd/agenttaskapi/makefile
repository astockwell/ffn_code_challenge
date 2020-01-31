ccnix:
	env GOOS=linux GOARCH=amd64 go build
	mv agenttaskapi agenttaskapi_linux_amd64

ccwin:
	env GOOS=windows GOARCH=amd64 go build
	mv agenttaskapi.exe agenttaskapi_windows_amd64.exe

ccosx:
	env GOOS=darwin GOARCH=amd64 go build
	mv agenttaskapi agenttaskapi_darwin_amd64

all: ccnix ccwin ccosx

.PHONY: ccnix ccwin ccosx all
