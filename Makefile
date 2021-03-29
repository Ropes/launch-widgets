.PHONY: mod
mod: 
	go mod download

.PHONY: test
test:
	go test -race ./pkg/...

.PHONY: run
run: mod
	go run cmd/main.go