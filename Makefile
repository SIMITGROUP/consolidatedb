run:
	go get .
	go build .
	@echo "Build successfully, run command with your own options"
	
windows:
	go get .
	GOOS=windows GOARCH=amd64 go build -o dist/consolidatedb-win.exe .
linux:
	go get .
	GOOS=linux GOARCH=amd64 go build -o dist/consolidatedb-linux.bin .
mac:
	go get .
	GOOS=darwin GOARCH=amd64 go build -o dist/consolidatedb-mac.bin .
mac-arm:
	go get .
	GOOS=darwin GOARCH=arm64 go build -o dist/consolidatedb-mac-arm.bin .