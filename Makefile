.PHONY: build run test swagger

build:
	go build -o bin/task_vault ./cmd/task_vault

run:
	go run ./cmd/task_vault

test:
	go test ./...

swagger:
	swag init -g cmd/task_vault/main.go --parseDependency --parseInternal
