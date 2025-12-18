run:
	@go run ./cmd/main/main.go

build:
	@go build -o ./dist/main ./cmd/main/main.go
