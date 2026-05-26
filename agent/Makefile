.PHONY: dev build clean install-deps

dev:
	wails dev

build:
	wails build

build-all:
	wails build -platform linux/amd64
	wails build -platform darwin/amd64
	wails build -platform darwin/arm64
	wails build -platform windows/amd64

clean:
	rm -rf build/ frontend/dist/

install-deps:
	go install github.com/wailsapp/wails/v2/cmd/wails@latest
	cd frontend && npm install

cli:
	go build -o notion-agent-cli ./cmd/...

test:
	go test ./...
