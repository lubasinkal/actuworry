.PHONY: help build run test clean deploy

help:
	@echo "Available commands:"
	@echo "  make build   - Build the application"
	@echo "  make run     - Run the application locally"
	@echo "  make test    - Run tests"
	@echo "  make clean   - Clean build artifacts"
	@echo "  make deploy  - Deploy to Render"

build:
	go build -tags netgo -ldflags '-s -w' -o app

run:
	go run main.go

test:
	go test ./backend/...
	cd backend/tests && bash test_api.sh

clean:
	rm -f app *.exe *.out
	go clean

deploy:
	git add .
	git commit -m "Deploy updates"
	git push origin main
