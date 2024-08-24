build:
	@go build -o ./bin/snippetbox ./cmd/web/.

test:
	@go test -v ./...

run: build
	@./bin/snippetbox
