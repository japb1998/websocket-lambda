profile=personal

.PHONY: build clean deploy gomodgen

build:
	export GO111MODULE=on
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/wsocket ./wshandler/cmd/connection
	env GOARCH=amd64 GOOS=linux go build -ldflags="-s -w" -o bin/messageHandler ./wshandler/cmd/messageHandler

deploy: build
	
	sls deploy --verbose --aws-profile "$(profile)"

gomodgen:
	chmod u+x gomod.sh
	./gomod.sh
