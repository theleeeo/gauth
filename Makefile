build:
	go mod tidy
	go build -race -o bin/${APP} ./cmd

.PHONY: lint
lint:
	golangci-lint run --fix --timeout=120s ./...

.PHONY: test
test:
	 go test ./...

test-coverage:
	go install github.com/ory/go-acc@latest
	go-acc -o coverage.out ./... -- -v
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

gql:
	cd gql && go run github.com/99designs/gqlgen generate
