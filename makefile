dev:
	wails dev

build:
	wails build

test:
	go test ./...

test-verbose:
	go test -v ./...