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

.PHONY: dev
dev:
	air .

.PHONY: build
build:
	@echo 'Building for Linux'
	go build -ldflags=${linker_flags} -o=./bin/gdv ./main.go
