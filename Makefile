.PHONY: fmt fmt-check vet lint test gate

fmt:
	gofmt -w .

fmt-check:
	test -z "$$(gofmt -l .)"

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test ./... -race

gate: fmt-check vet lint test
