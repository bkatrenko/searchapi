.PHONY: build integration-test docker-up docker-down clear 

build:
	@go build

integration-test: docker-up dependency
	@go test -v ./...

docker-up:
	@docker-compose up -d

docker-down:
	@docker-compose down

clear: docker-down