BINARY_NAME=gdv

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N]' && read ans && [ $${ans:-N} = y ]

## audit: tidy dependencies, format, vet and test
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	@echo 'Running tests...'
	ENV=testing go test -race -vet=off ./...

## dev: run with hot-reloading
.PHONY: dev
dev:
	ENV=development air .

## build: build binary
.PHONY: build
build:
	@echo 'Building for Linux'
	go build -o=./bin/${BINARY_NAME} ./main.go

## deps: install external dependencies not used in source code
.PHONY: deps
deps: confirm
	@echo 'Installing `air` for hot-reloading'
	go install github.com/cosmtrek/air@latest
