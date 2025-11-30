.PHONY: run download-deps docs install-mockgen create-mocks test dockerize docker-run

run:
	go run cmd/api/*
	
download-deps:
	@echo "Downloading dependencies..."
	@go mod download

docs:
	@echo "Creating Swagger documentation using local binary..."
	@if ! command -v swag >/dev/null 2>&1; then \
		echo "swag not found, downloading..."; \
		go install github.com/swaggo/swag/cmd/swag@latest; \
	fi
	@echo "Generating Swagger docs..."
	swag init --parseDependency -g internal/router/router.go -o docs

install-mockgen:
	which mockgen || go  install go.uber.org/mock/mockgen@v0.4.0

create-mocks: install-mockgen
	mockgen -source=./internal/repository/repository.go -destination=./internal/repository/mocks/repository_mock.go -package=mocks
	mockgen -source=./internal/repository/message.go -destination=./internal/repository/mocks/message_mock.go -package=mocks
	mockgen -source=./internal/repository/message_cache.go -destination=./internal/repository/mocks/message_cache_mock.go -package=mocks
	mockgen -source=./internal/services/message_sender.go -destination=./internal/services/mocks/message_sender_mock.go -package=mocks

test:
	@echo "Running tests..."
	go test -p 1 ./internal/... -race -covermode=atomic -coverprofile=coverage.out -v -tags=skip_coverage
	go tool cover -html=coverage.out -o coverage.html

dockerize:
	docker build -t go-template-microservice --no-cache -f ./Dockerfile .

docker-run:
	docker run -p 8080:8080 -d --env-file .env --name $(or $(CONTAINER_NAME),go-template-microservice) go-template-microservice