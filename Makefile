build:
	@go build -o bin/fs

run: build
	@bin/fs.exe

test:
	@go test ./... -v