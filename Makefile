SWAG_VERSION := v1.16.6

.PHONY: swagger test build

swagger:
	go run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION) init -g cmd/main/main.go -o docs/swagger --parseInternal --parseDependency

test:
	go test -race ./...

build:
	go build ./...
